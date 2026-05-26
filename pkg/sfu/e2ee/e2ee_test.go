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

package e2ee_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/msmclass/samvaad/pkg/sfu/e2ee"
)

// TestKeyManagerRoundTrip verifies that two participants performing ECDH arrive
// at the same shared secret.
func TestKeyManagerRoundTrip(t *testing.T) {
	alice, err := e2ee.NewKeyManager()
	require.NoError(t, err)

	bob, err := e2ee.NewKeyManager()
	require.NoError(t, err)

	alicePub, err := alice.PublicKeyBytes()
	require.NoError(t, err)
	require.Len(t, alicePub, 32)

	bobPub, err := bob.PublicKeyBytes()
	require.NoError(t, err)
	require.Len(t, bobPub, 32)

	// Alice derives shared secret using Bob's public key.
	sharedAlice, err := alice.DeriveSharedSecret(bobPub)
	require.NoError(t, err)

	// Bob derives shared secret using Alice's public key.
	sharedBob, err := bob.DeriveSharedSecret(alicePub)
	require.NoError(t, err)

	// ECDH property: both secrets must be equal.
	require.True(t, bytes.Equal(sharedAlice, sharedBob),
		"ECDH shared secrets must be equal")
}

// TestDeriveSessionKey verifies that identical inputs produce identical keys
// and that different salts produce different keys.
func TestDeriveSessionKey(t *testing.T) {
	secret := bytes.Repeat([]byte{0xAB}, 32)
	salt1 := []byte("samvaad-room-bharat")
	salt2 := []byte("samvaad-room-india")

	key1a, err := e2ee.DeriveSessionKey(secret, salt1)
	require.NoError(t, err)
	require.Len(t, key1a, e2ee.SessionKeyLen)

	// Deterministic: same inputs → same key.
	key1b, err := e2ee.DeriveSessionKey(secret, salt1)
	require.NoError(t, err)
	require.True(t, bytes.Equal(key1a, key1b), "HKDF must be deterministic")

	// Different salt → different key.
	key2, err := e2ee.DeriveSessionKey(secret, salt2)
	require.NoError(t, err)
	require.False(t, bytes.Equal(key1a, key2), "different salts must produce different keys")
}

// TestFullE2EEHandshake performs a complete two-party key exchange and session
// key derivation, simulating the DataChannel-based in-band key negotiation.
func TestFullE2EEHandshake(t *testing.T) {
	// --- Participant setup ---
	alice, err := e2ee.NewKeyManager()
	require.NoError(t, err)

	bob, err := e2ee.NewKeyManager()
	require.NoError(t, err)

	alicePub, err := alice.PublicKeyBytes()
	require.NoError(t, err)

	bobPub, err := bob.PublicKeyBytes()
	require.NoError(t, err)

	// --- Wire encoding (Alice broadcasts her public key) ---
	pktOut := &e2ee.KeyExchangePacket{
		ParticipantID: "alice-pid-12345",
		PublicKey:     alicePub,
	}
	wire, err := e2ee.MarshalKeyExchangePacket(pktOut)
	require.NoError(t, err)

	// Magic detection should fire.
	require.True(t, e2ee.IsKeyExchangePacket(wire))

	// --- Wire decoding (Bob receives Alice's packet) ---
	pktIn, err := e2ee.ParseKeyExchangePacket(wire)
	require.NoError(t, err)
	require.Equal(t, "alice-pid-12345", pktIn.ParticipantID)
	require.True(t, bytes.Equal(alicePub, pktIn.PublicKey))

	// --- ECDH ---
	sharedAlice, err := alice.DeriveSharedSecret(bobPub)
	require.NoError(t, err)

	sharedBob, err := bob.DeriveSharedSecret(pktIn.PublicKey)
	require.NoError(t, err)
	require.True(t, bytes.Equal(sharedAlice, sharedBob))

	// --- Session key derivation (both use the same room salt) ---
	roomSalt := []byte("room:bharat-vidya-conference-2026")

	keyAlice, err := e2ee.DeriveSessionKey(sharedAlice, roomSalt)
	require.NoError(t, err)

	keyBob, err := e2ee.DeriveSessionKey(sharedBob, roomSalt)
	require.NoError(t, err)

	require.True(t, bytes.Equal(keyAlice, keyBob),
		"session keys derived from equal shared secrets and salt must match")
}

// TestKeyExchangePacket_MalformedInputs validates that the parser rejects
// truncated and corrupted wire frames without panicking.
func TestKeyExchangePacket_MalformedInputs(t *testing.T) {
	t.Run("too short", func(t *testing.T) {
		_, err := e2ee.ParseKeyExchangePacket([]byte{0x53, 0x45, 0x01})
		require.ErrorIs(t, err, e2ee.ErrInvalidPacket)
	})

	t.Run("wrong magic", func(t *testing.T) {
		// Craft a packet with correct length but wrong magic bytes.
		buf := make([]byte, 2+1+4+32)
		buf[0], buf[1] = 0xDE, 0xAD
		_, err := e2ee.ParseKeyExchangePacket(buf)
		require.ErrorIs(t, err, e2ee.ErrInvalidPacket)
	})

	t.Run("empty slice", func(t *testing.T) {
		_, err := e2ee.ParseKeyExchangePacket(nil)
		require.ErrorIs(t, err, e2ee.ErrInvalidPacket)
	})

	t.Run("wrong public key length in marshal", func(t *testing.T) {
		pkt := &e2ee.KeyExchangePacket{
			ParticipantID: "test",
			PublicKey:     []byte{0x01, 0x02}, // only 2 bytes — too short
		}
		_, err := e2ee.MarshalKeyExchangePacket(pkt)
		require.ErrorIs(t, err, e2ee.ErrInvalidPublicKeyLen)
	})
}

// TestIsKeyExchangePacket covers edge cases for the magic-byte check.
func TestIsKeyExchangePacket(t *testing.T) {
	require.False(t, e2ee.IsKeyExchangePacket(nil))
	require.False(t, e2ee.IsKeyExchangePacket([]byte{0x53})) // only 1 byte
	require.False(t, e2ee.IsKeyExchangePacket([]byte{0x00, 0x00}))

	// Build a valid header and confirm detection.
	km, err := e2ee.NewKeyManager()
	require.NoError(t, err)
	pub, err := km.PublicKeyBytes()
	require.NoError(t, err)

	wire, err := e2ee.MarshalKeyExchangePacket(&e2ee.KeyExchangePacket{
		ParticipantID: "p1",
		PublicKey:     pub,
	})
	require.NoError(t, err)
	require.True(t, e2ee.IsKeyExchangePacket(wire))
}

// TestDeriveSharedSecret_InvalidKey ensures proper error handling for bad input.
func TestDeriveSharedSecret_InvalidKey(t *testing.T) {
	km, err := e2ee.NewKeyManager()
	require.NoError(t, err)

	// Wrong length.
	_, err = km.DeriveSharedSecret([]byte{0x01, 0x02, 0x03})
	require.ErrorIs(t, err, e2ee.ErrInvalidPublicKeyLen)
}
