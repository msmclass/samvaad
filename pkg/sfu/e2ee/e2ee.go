// Copyright 2026 Samvaad Project, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package e2ee implements the zero-knowledge End-to-End Encryption key exchange
// layer for the Samvaad Bharat SFU.
//
// # Architecture
//
// The SFU acts as a pure routing fabric: it forwards encrypted RTP media and
// key-exchange packets without ever possessing the ability to decrypt either.
//
// Key exchange happens entirely in-band over WebRTC Reliable Data Channels
// (using the samvaad.DataPacket_Kind_USER message type). The SFU serialises and
// forwards KeyExchangePackets between participants without inspecting key material.
//
// # Protocol
//
//  1. Each participant generates an ephemeral X25519 ECDH key-pair via NewKeyManager().
//  2. The participant broadcasts their public key to all peers in the room by
//     wrapping it in a KeyExchangePacket and sending it over the reliable data channel.
//  3. Each receiving participant calls DeriveSharedSecret() with the sender's public
//     key to compute a 32-byte shared secret using X25519 ECDH.
//  4. The shared secret is expanded with HKDF-SHA256 into a 32-byte AES-256-GCM
//     session key using DeriveSessionKey().
//  5. Media frames are encrypted client-side with the session key; the SFU only
//     ever sees ciphertext.
//
// # FIPS Compliance
//
// When compiled with GOEXPERIMENT=boringcrypto the underlying crypto/ecdh and
// crypto/sha256 packages use BoringCrypto-backed implementations, satisfying
// FIPS 140-2 requirements for GOI / defence sector deployments.
package e2ee

import (
	"crypto/ecdh"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"io"

	"golang.org/x/crypto/hkdf"
)

// ----- Error sentinels -------------------------------------------------------

var (
	// ErrInvalidPublicKeyLen is returned when a peer public key byte slice has
	// an unexpected length (must be exactly 32 bytes for X25519).
	ErrInvalidPublicKeyLen = errors.New("e2ee: invalid public key length, expected 32 bytes (X25519)")

	// ErrNilPrivateKey is returned when DeriveSharedSecret is called on a
	// KeyManager that has not been initialised via NewKeyManager().
	ErrNilPrivateKey = errors.New("e2ee: key manager has no private key — call NewKeyManager()")

	// ErrInvalidPacket is returned when ParseKeyExchangePacket encounters a
	// malformed wire frame.
	ErrInvalidPacket = errors.New("e2ee: malformed key exchange packet")
)

// ----- Wire frame constants --------------------------------------------------

const (
	// keyExchangeMagic is a 2-byte magic number prefixing every KeyExchangePacket
	// on the wire, allowing receivers to distinguish E2EE frames from other
	// DataChannel payloads without a protobuf parser.
	keyExchangeMagic = uint16(0x5345) // "SE" — Samvaad E2EE

	// keyExchangeVersion is the protocol version byte.
	keyExchangeVersion = byte(1)

	// publicKeyLen is the fixed length of an X25519 public key in bytes.
	publicKeyLen = 32

	// minPacketLen is the smallest valid KeyExchangePacket on the wire:
	// 2 (magic) + 1 (version) + 4 (participant ID length) + publicKeyLen
	minPacketLen = 2 + 1 + 4 + publicKeyLen

	// hkdfInfoE2EE is the HKDF info string distinguishing Samvaad session keys
	// from any other key material derived from the same shared secret.
	hkdfInfoE2EE = "samvaad-e2ee-v1"

	// SessionKeyLen is the length of an AES-256-GCM session key in bytes.
	SessionKeyLen = 32
)

// ----- KeyManager ------------------------------------------------------------

// KeyManager holds one ephemeral X25519 ECDH key-pair for a single participant
// session. A new KeyManager should be created whenever a participant (re)joins a
// room, ensuring forward secrecy across sessions.
//
// KeyManager is NOT safe for concurrent use. Callers must synchronise access
// externally if the manager is shared across goroutines.
type KeyManager struct {
	privateKey *ecdh.PrivateKey
	publicKey  *ecdh.PublicKey
}

// NewKeyManager generates a fresh X25519 ECDH key-pair using the system's
// cryptographically secure random number generator.
func NewKeyManager() (*KeyManager, error) {
	priv, err := ecdh.X25519().GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}
	return &KeyManager{
		privateKey: priv,
		publicKey:  priv.PublicKey(),
	}, nil
}

// PublicKeyBytes returns the 32-byte X25519 public key that should be
// distributed to all peers in the room via a KeyExchangePacket.
func (km *KeyManager) PublicKeyBytes() ([]byte, error) {
	if km.privateKey == nil {
		return nil, ErrNilPrivateKey
	}
	return km.publicKey.Bytes(), nil
}

// DeriveSharedSecret performs the X25519 ECDH operation with the provided peer
// public key and returns the 32-byte raw shared secret.
//
// The returned secret MUST NOT be used directly as a cipher key; pass it to
// DeriveSessionKey() to produce a uniformly distributed key via HKDF.
func (km *KeyManager) DeriveSharedSecret(peerPublicKeyBytes []byte) ([]byte, error) {
	if km.privateKey == nil {
		return nil, ErrNilPrivateKey
	}
	if len(peerPublicKeyBytes) != publicKeyLen {
		return nil, ErrInvalidPublicKeyLen
	}

	peerKey, err := ecdh.X25519().NewPublicKey(peerPublicKeyBytes)
	if err != nil {
		return nil, err
	}

	secret, err := km.privateKey.ECDH(peerKey)
	if err != nil {
		return nil, err
	}
	return secret, nil
}

// ----- Session key derivation ------------------------------------------------

// DeriveSessionKey expands a raw ECDH shared secret into a SessionKeyLen-byte
// AES-256-GCM session key using HKDF-SHA256.
//
// salt should be a random, per-session value exchanged between peers (e.g. the
// room name hashed, or a fresh nonce). An empty salt causes HKDF to use a
// zero-length salt, which is still cryptographically sound per RFC 5869 §3.1
// but a random salt provides stronger security guarantees.
func DeriveSessionKey(sharedSecret, salt []byte) ([]byte, error) {
	h := hkdf.New(sha256.New, sharedSecret, salt, []byte(hkdfInfoE2EE))
	key := make([]byte, SessionKeyLen)
	if _, err := io.ReadFull(h, key); err != nil {
		return nil, err
	}
	return key, nil
}

// ----- Wire frame encoding / decoding ----------------------------------------

// KeyExchangePacket is the wire frame transmitted over WebRTC Reliable Data
// Channels to distribute a participant's X25519 public key in-band.
//
// Wire layout (little-endian):
//
//	Offset   Size  Field
//	     0      2  Magic       (0x5345 — "SE")
//	     2      1  Version     (currently 0x01)
//	     3      4  IDLen       uint32, byte length of ParticipantID
//	     7   IDLen ParticipantID UTF-8 string
//	 7+IDLen  32   PublicKey   X25519 public key bytes
type KeyExchangePacket struct {
	// ParticipantID uniquely identifies the sender within the room.
	// This matches samvaad.ParticipantID and is used by receivers to associate
	// the public key with the correct peer.
	ParticipantID string

	// PublicKey is the sender's 32-byte X25519 public key.
	PublicKey []byte
}

// MarshalKeyExchangePacket serialises a KeyExchangePacket into its compact
// binary wire representation suitable for transmission over a DataChannel.
func MarshalKeyExchangePacket(pkt *KeyExchangePacket) ([]byte, error) {
	if len(pkt.PublicKey) != publicKeyLen {
		return nil, ErrInvalidPublicKeyLen
	}

	idBytes := []byte(pkt.ParticipantID)
	idLen := len(idBytes)

	// Total: 2 (magic) + 1 (version) + 4 (idLen) + idLen + 32 (pubkey)
	buf := make([]byte, 2+1+4+idLen+publicKeyLen)

	binary.LittleEndian.PutUint16(buf[0:2], keyExchangeMagic)
	buf[2] = keyExchangeVersion
	binary.LittleEndian.PutUint32(buf[3:7], uint32(idLen))
	copy(buf[7:7+idLen], idBytes)
	copy(buf[7+idLen:], pkt.PublicKey)

	return buf, nil
}

// ParseKeyExchangePacket deserialises a raw DataChannel payload into a
// KeyExchangePacket. Returns ErrInvalidPacket when the frame is malformed.
//
// Callers should verify that the first two bytes equal keyExchangeMagic (0x5345)
// before calling this function to avoid misinterpreting non-E2EE DataChannel
// messages.
func ParseKeyExchangePacket(data []byte) (*KeyExchangePacket, error) {
	if len(data) < minPacketLen {
		return nil, ErrInvalidPacket
	}

	magic := binary.LittleEndian.Uint16(data[0:2])
	if magic != keyExchangeMagic {
		return nil, ErrInvalidPacket
	}

	version := data[2]
	if version != keyExchangeVersion {
		// Future versions may add fields; reject unknown versions to avoid
		// silently misinterpreting newer wire frames.
		return nil, ErrInvalidPacket
	}

	idLen := int(binary.LittleEndian.Uint32(data[3:7]))
	expectedTotal := 7 + idLen + publicKeyLen
	if len(data) < expectedTotal {
		return nil, ErrInvalidPacket
	}

	participantID := string(data[7 : 7+idLen])
	publicKey := make([]byte, publicKeyLen)
	copy(publicKey, data[7+idLen:7+idLen+publicKeyLen])

	return &KeyExchangePacket{
		ParticipantID: participantID,
		PublicKey:     publicKey,
	}, nil
}

// IsKeyExchangePacket returns true if the first two bytes of data match the
// Samvaad E2EE wire magic. Use this to route DataChannel payloads without
// fully parsing them.
func IsKeyExchangePacket(data []byte) bool {
	if len(data) < 2 {
		return false
	}
	return binary.LittleEndian.Uint16(data[0:2]) == keyExchangeMagic
}
