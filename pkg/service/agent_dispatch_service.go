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
	"fmt"

	"github.com/msmclass/samvaad/pkg/routing"
	"github.com/msmclass/samvaad/pkg/samvaad/agent"
	"github.com/msmclass/samvaad/pkg/proto/samvaad"
	"github.com/msmclass/samvaad/pkg/samvaad/logger"
	"github.com/msmclass/samvaad/pkg/samvaad/rpc"
	"github.com/msmclass/samvaad/pkg/samvaad/utils"
	"github.com/msmclass/samvaad/pkg/samvaad/utils/guid"
	"github.com/msmclass/samvaad/pkg/samvaad/psrpc"
)

type AgentDispatchService struct {
	agentDispatchClient rpc.TypedAgentDispatchInternalClient
	topicFormatter      rpc.TopicFormatter
	roomAllocator       RoomAllocator
	router              routing.MessageRouter
}

func NewAgentDispatchService(
	agentDispatchClient rpc.TypedAgentDispatchInternalClient,
	topicFormatter rpc.TopicFormatter,
	roomAllocator RoomAllocator,
	router routing.MessageRouter,
) *AgentDispatchService {
	return &AgentDispatchService{
		agentDispatchClient: agentDispatchClient,
		topicFormatter:      topicFormatter,
		roomAllocator:       roomAllocator,
		router:              router,
	}
}

func (ag *AgentDispatchService) CreateDispatch(ctx context.Context, req *samvaad.CreateAgentDispatchRequest) (*samvaad.AgentDispatch, error) {
	AppendLogFields(ctx, "room", req.Room, "request", logger.Proto(redactCreateAgentDispatchRequest(req)))
	err := EnsureAdminPermission(ctx, samvaad.RoomName(req.Room))
	if err != nil {
		return nil, twirpAuthError(err)
	}

	if err := agent.ValidateDeployment(req.GetDeployment()); err != nil {
		return nil, psrpc.NewError(psrpc.InvalidArgument, err)
	}

	if ag.roomAllocator.AutoCreateEnabled(ctx) {
		err := ag.roomAllocator.SelectRoomNode(ctx, samvaad.RoomName(req.Room), "")
		if err != nil {
			return nil, err
		}

		_, err = ag.router.CreateRoom(ctx, &samvaad.CreateRoomRequest{Name: req.Room})
		if err != nil {
			return nil, err
		}
	}

	dispatch := &samvaad.AgentDispatch{
		Id:            guid.New(guid.AgentDispatchPrefix),
		AgentName:     req.AgentName,
		Room:          req.Room,
		Metadata:      req.Metadata,
		RestartPolicy: req.RestartPolicy,
		Deployment:    req.Deployment,
	}
	return ag.agentDispatchClient.CreateDispatch(ctx, ag.topicFormatter.RoomTopic(ctx, samvaad.RoomName(req.Room)), dispatch)
}

func (ag *AgentDispatchService) DeleteDispatch(ctx context.Context, req *samvaad.DeleteAgentDispatchRequest) (*samvaad.AgentDispatch, error) {
	AppendLogFields(ctx, "room", req.Room, "request", logger.Proto(req))
	err := EnsureAdminPermission(ctx, samvaad.RoomName(req.Room))
	if err != nil {
		return nil, twirpAuthError(err)
	}

	return ag.agentDispatchClient.DeleteDispatch(ctx, ag.topicFormatter.RoomTopic(ctx, samvaad.RoomName(req.Room)), req)
}

func (ag *AgentDispatchService) ListDispatch(ctx context.Context, req *samvaad.ListAgentDispatchRequest) (*samvaad.ListAgentDispatchResponse, error) {
	AppendLogFields(ctx, "room", req.Room, "request", logger.Proto(req))
	err := EnsureAdminPermission(ctx, samvaad.RoomName(req.Room))
	if err != nil {
		return nil, twirpAuthError(err)
	}

	return ag.agentDispatchClient.ListDispatch(ctx, ag.topicFormatter.RoomTopic(ctx, samvaad.RoomName(req.Room)), req)
}

func redactCreateAgentDispatchRequest(req *samvaad.CreateAgentDispatchRequest) *samvaad.CreateAgentDispatchRequest {
	if req.Metadata == "" {
		return req
	}

	clone := utils.CloneProto(req)

	// replace with size of metadata to provide visibility on request size
	if clone.Metadata != "" {
		clone.Metadata = fmt.Sprintf("__size: %d", len(clone.Metadata))
	}

	return clone
}


