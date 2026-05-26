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

package service

import (
	"context"
	"time"

	"github.com/msmclass/samvaad/pkg/proto/samvaad"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

// encapsulates CRUD operations for room settings
//
//counterfeiter:generate . ObjectStore
type ObjectStore interface {
	ServiceStore
	OSSServiceStore

	// enable locking on a specific room to prevent race
	// returns a (lock uuid, error)
	LockRoom(ctx context.Context, roomName samvaad.RoomName, duration time.Duration) (string, error)
	UnlockRoom(ctx context.Context, roomName samvaad.RoomName, uid string) error

	StoreRoom(ctx context.Context, room *samvaad.Room, internal *samvaad.RoomInternal) error

	StoreParticipant(ctx context.Context, roomName samvaad.RoomName, participant *samvaad.ParticipantInfo) error
	DeleteParticipant(ctx context.Context, roomName samvaad.RoomName, identity samvaad.ParticipantIdentity) error
}

//counterfeiter:generate . ServiceStore
type ServiceStore interface {
	LoadRoom(ctx context.Context, roomName samvaad.RoomName, includeInternal bool) (*samvaad.Room, *samvaad.RoomInternal, error)
	RoomExists(ctx context.Context, roomName samvaad.RoomName) (bool, error)

	// ListRooms returns currently active rooms. if names is not nil, it'll filter and return
	// only rooms that match
	ListRooms(ctx context.Context, roomNames []samvaad.RoomName) ([]*samvaad.Room, error)
}

type OSSServiceStore interface {
	DeleteRoom(ctx context.Context, roomName samvaad.RoomName) error
	HasParticipant(context.Context, samvaad.RoomName, samvaad.ParticipantIdentity) (bool, error)
	LoadParticipant(ctx context.Context, roomName samvaad.RoomName, identity samvaad.ParticipantIdentity) (*samvaad.ParticipantInfo, error)
	ListParticipants(ctx context.Context, roomName samvaad.RoomName) ([]*samvaad.ParticipantInfo, error)
}

//counterfeiter:generate . EgressStore
type EgressStore interface {
	StoreEgress(ctx context.Context, info *samvaad.EgressInfo) error
	LoadEgress(ctx context.Context, egressID string) (*samvaad.EgressInfo, error)
	ListEgress(ctx context.Context, roomName samvaad.RoomName, active bool) ([]*samvaad.EgressInfo, error)
	UpdateEgress(ctx context.Context, info *samvaad.EgressInfo) error
}

//counterfeiter:generate . IngressStore
type IngressStore interface {
	StoreIngress(ctx context.Context, info *samvaad.IngressInfo) error
	LoadIngress(ctx context.Context, ingressID string) (*samvaad.IngressInfo, error)
	LoadIngressFromStreamKey(ctx context.Context, streamKey string) (*samvaad.IngressInfo, error)
	ListIngress(ctx context.Context, roomName samvaad.RoomName) ([]*samvaad.IngressInfo, error)
	UpdateIngress(ctx context.Context, info *samvaad.IngressInfo) error
	UpdateIngressState(ctx context.Context, ingressId string, state *samvaad.IngressState) error
	DeleteIngress(ctx context.Context, info *samvaad.IngressInfo) error
}

//counterfeiter:generate . RoomAllocator
type RoomAllocator interface {
	AutoCreateEnabled(ctx context.Context) bool
	SelectRoomNode(ctx context.Context, roomName samvaad.RoomName, nodeID samvaad.NodeID) error
	CreateRoom(ctx context.Context, req *samvaad.CreateRoomRequest, isExplicit bool) (*samvaad.Room, *samvaad.RoomInternal, bool, error)
	ValidateCreateRoom(ctx context.Context, roomName samvaad.RoomName) error
}

//counterfeiter:generate . SIPStore
type SIPStore interface {
	StoreSIPTrunk(ctx context.Context, info *samvaad.SIPTrunkInfo) error
	StoreSIPInboundTrunk(ctx context.Context, info *samvaad.SIPInboundTrunkInfo) error
	StoreSIPOutboundTrunk(ctx context.Context, info *samvaad.SIPOutboundTrunkInfo) error
	LoadSIPTrunk(ctx context.Context, sipTrunkID string) (*samvaad.SIPTrunkInfo, error)
	LoadSIPInboundTrunk(ctx context.Context, sipTrunkID string) (*samvaad.SIPInboundTrunkInfo, error)
	LoadSIPOutboundTrunk(ctx context.Context, sipTrunkID string) (*samvaad.SIPOutboundTrunkInfo, error)
	ListSIPTrunk(ctx context.Context, opts *samvaad.ListSIPTrunkRequest) (*samvaad.ListSIPTrunkResponse, error)
	ListSIPInboundTrunk(ctx context.Context, opts *samvaad.ListSIPInboundTrunkRequest) (*samvaad.ListSIPInboundTrunkResponse, error)
	ListSIPOutboundTrunk(ctx context.Context, opts *samvaad.ListSIPOutboundTrunkRequest) (*samvaad.ListSIPOutboundTrunkResponse, error)
	DeleteSIPTrunk(ctx context.Context, sipTrunkID string) error

	StoreSIPDispatchRule(ctx context.Context, info *samvaad.SIPDispatchRuleInfo) error
	LoadSIPDispatchRule(ctx context.Context, sipDispatchRuleID string) (*samvaad.SIPDispatchRuleInfo, error)
	ListSIPDispatchRule(ctx context.Context, opts *samvaad.ListSIPDispatchRuleRequest) (*samvaad.ListSIPDispatchRuleResponse, error)
	DeleteSIPDispatchRule(ctx context.Context, sipDispatchRuleID string) error
}

//counterfeiter:generate . AgentStore
type AgentStore interface {
	StoreAgentDispatch(ctx context.Context, dispatch *samvaad.AgentDispatch) error
	DeleteAgentDispatch(ctx context.Context, dispatch *samvaad.AgentDispatch) error
	ListAgentDispatches(ctx context.Context, roomName samvaad.RoomName) ([]*samvaad.AgentDispatch, error)

	StoreAgentJob(ctx context.Context, job *samvaad.Job) error
	DeleteAgentJob(ctx context.Context, job *samvaad.Job) error
}


