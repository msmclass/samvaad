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
	"errors"
	"time"

	"github.com/dennwc/iters"
	"github.com/twitchtv/twirp"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/msmclass/samvaad/pkg/proto/samvaad"
	"github.com/msmclass/samvaad/pkg/samvaad/logger"
	"github.com/msmclass/samvaad/pkg/samvaad/rpc"
	"github.com/msmclass/samvaad/pkg/samvaad/sip"
	"github.com/msmclass/samvaad/pkg/samvaad/utils"
	"github.com/msmclass/samvaad/pkg/samvaad/utils/guid"
	"github.com/msmclass/samvaad/pkg/samvaad/psrpc"

	"github.com/msmclass/samvaad/pkg/config"
	"github.com/msmclass/samvaad/pkg/telemetry"
)

type SIPService struct {
	conf        *config.SIPConfig
	nodeID      samvaad.NodeID
	bus         psrpc.MessageBus
	psrpcClient rpc.SIPClient
	store       SIPStore
	roomService samvaad.RoomService
}

func NewSIPService(
	conf *config.SIPConfig,
	nodeID samvaad.NodeID,
	bus psrpc.MessageBus,
	psrpcClient rpc.SIPClient,
	store SIPStore,
	rs samvaad.RoomService,
	ts telemetry.TelemetryService,
) *SIPService {
	return &SIPService{
		conf:        conf,
		nodeID:      nodeID,
		bus:         bus,
		psrpcClient: psrpcClient,
		store:       store,
		roomService: rs,
	}
}

func (s *SIPService) CreateSIPTrunk(ctx context.Context, req *samvaad.CreateSIPTrunkRequest) (*samvaad.SIPTrunkInfo, error) {
	if err := EnsureSIPAdminPermission(ctx); err != nil {
		return nil, twirpAuthError(err)
	}
	if s.store == nil {
		return nil, ErrSIPNotConnected
	}
	if len(req.InboundNumbersRegex) != 0 {
		return nil, twirp.NewError(twirp.InvalidArgument, "Trunks with InboundNumbersRegex are deprecated. Use InboundNumbers instead.")
	}

	// Keep ID empty, so that validation can print "<new>" instead of a non-existent ID in the error.
	info := &samvaad.SIPTrunkInfo{
		InboundAddresses: req.InboundAddresses,
		OutboundAddress:  req.OutboundAddress,
		OutboundNumber:   req.OutboundNumber,
		InboundNumbers:   req.InboundNumbers,
		InboundUsername:  req.InboundUsername,
		InboundPassword:  req.InboundPassword,
		OutboundUsername: req.OutboundUsername,
		OutboundPassword: req.OutboundPassword,
		Name:             req.Name,
		Metadata:         req.Metadata,
	}
	if err := info.Validate(); err != nil {
		return nil, twirp.WrapError(twirp.NewError(twirp.InvalidArgument, err.Error()), err)
	}

	// Validate all trunks including the new one first.
	it, err := ListSIPInboundTrunk(ctx, s.store, &samvaad.ListSIPInboundTrunkRequest{}, info.AsInbound())
	if err != nil {
		return nil, err
	}
	defer it.Close()
	if err = sip.ValidateTrunksIter(it); err != nil {
		return nil, err
	}

	// Now we can generate ID and store.
	info.SipTrunkId = guid.New(utils.SIPTrunkPrefix)
	if err := s.store.StoreSIPTrunk(ctx, info); err != nil {
		return nil, err
	}
	return info, nil
}

func (s *SIPService) CreateSIPInboundTrunk(ctx context.Context, req *samvaad.CreateSIPInboundTrunkRequest) (*samvaad.SIPInboundTrunkInfo, error) {
	if err := EnsureSIPAdminPermission(ctx); err != nil {
		return nil, twirpAuthError(err)
	}
	if s.store == nil {
		return nil, ErrSIPNotConnected
	}
	if err := req.Validate(); err != nil {
		return nil, twirp.WrapError(twirp.NewError(twirp.InvalidArgument, err.Error()), err)
	}

	info := req.Trunk
	if info.SipTrunkId != "" {
		return nil, twirp.NewError(twirp.InvalidArgument, "trunk ID must be empty")
	}
	AppendLogFields(ctx, "trunk", logger.Proto(info))

	// Keep ID empty still, so that validation can print "<new>" instead of a non-existent ID in the error.

	// Validate all trunks including the new one first.
	it, err := ListSIPInboundTrunk(ctx, s.store, &samvaad.ListSIPInboundTrunkRequest{
		Numbers: req.GetTrunk().GetNumbers(),
	}, info)
	if err != nil {
		return nil, err
	}
	defer it.Close()
	if err = sip.ValidateTrunksIter(it); err != nil {
		return nil, err
	}

	// Now we can generate ID and store.
	info.SipTrunkId = guid.New(utils.SIPTrunkPrefix)
	if err := s.store.StoreSIPInboundTrunk(ctx, info); err != nil {
		return nil, err
	}
	return info, nil
}

func (s *SIPService) CreateSIPOutboundTrunk(ctx context.Context, req *samvaad.CreateSIPOutboundTrunkRequest) (*samvaad.SIPOutboundTrunkInfo, error) {
	if err := EnsureSIPAdminPermission(ctx); err != nil {
		return nil, twirpAuthError(err)
	}
	if s.store == nil {
		return nil, ErrSIPNotConnected
	}
	if err := req.Validate(); err != nil {
		return nil, twirp.WrapError(twirp.NewError(twirp.InvalidArgument, err.Error()), err)
	}

	info := req.Trunk
	if info.SipTrunkId != "" {
		return nil, twirp.NewError(twirp.InvalidArgument, "trunk ID must be empty")
	}
	AppendLogFields(ctx, "trunk", logger.Proto(info))

	// No additional validation needed for outbound.
	info.SipTrunkId = guid.New(utils.SIPTrunkPrefix)
	if err := s.store.StoreSIPOutboundTrunk(ctx, info); err != nil {
		return nil, err
	}
	return info, nil
}

func (s *SIPService) UpdateSIPInboundTrunk(ctx context.Context, req *samvaad.UpdateSIPInboundTrunkRequest) (*samvaad.SIPInboundTrunkInfo, error) {
	if err := EnsureSIPAdminPermission(ctx); err != nil {
		return nil, twirpAuthError(err)
	}
	if s.store == nil {
		return nil, ErrSIPNotConnected
	}
	if err := req.Validate(); err != nil {
		return nil, twirp.WrapError(twirp.NewError(twirp.InvalidArgument, err.Error()), err)
	}

	AppendLogFields(ctx,
		"request", logger.Proto(req),
		"trunkID", req.SipTrunkId,
	)

	// Validate all trunks including the new one first.
	info, err := s.store.LoadSIPInboundTrunk(ctx, req.SipTrunkId)
	if err != nil {
		if errors.Is(err, ErrSIPTrunkNotFound) {
			return nil, twirp.NewError(twirp.NotFound, err.Error())
		}
		return nil, err
	}
	switch a := req.Action.(type) {
	default:
		return nil, errors.New("missing or unsupported action")
	case samvaad.UpdateSIPInboundTrunkRequestAction:
		info, err = a.Apply(info)
		if err != nil {
			return nil, err
		}
	}

	it, err := ListSIPInboundTrunk(ctx, s.store, &samvaad.ListSIPInboundTrunkRequest{
		Numbers: info.Numbers,
	})
	if err != nil {
		return nil, err
	}
	defer it.Close()
	if err = sip.ValidateTrunksIter(it, sip.WithTrunkReplace(func(t *samvaad.SIPInboundTrunkInfo) *samvaad.SIPInboundTrunkInfo {
		if req.SipTrunkId == t.SipTrunkId {
			return info // updated one
		}
		return t
	})); err != nil {
		return nil, err
	}
	if err := s.store.StoreSIPInboundTrunk(ctx, info); err != nil {
		return nil, err
	}
	return info, nil
}

func (s *SIPService) UpdateSIPOutboundTrunk(ctx context.Context, req *samvaad.UpdateSIPOutboundTrunkRequest) (*samvaad.SIPOutboundTrunkInfo, error) {
	if err := EnsureSIPAdminPermission(ctx); err != nil {
		return nil, twirpAuthError(err)
	}
	if s.store == nil {
		return nil, ErrSIPNotConnected
	}
	if err := req.Validate(); err != nil {
		return nil, twirp.WrapError(twirp.NewError(twirp.InvalidArgument, err.Error()), err)
	}

	AppendLogFields(ctx,
		"request", logger.Proto(req),
		"trunkID", req.SipTrunkId,
	)

	info, err := s.store.LoadSIPOutboundTrunk(ctx, req.SipTrunkId)
	if err != nil {
		if errors.Is(err, ErrSIPTrunkNotFound) {
			return nil, twirp.NewError(twirp.NotFound, err.Error())
		}
		return nil, err
	}
	switch a := req.Action.(type) {
	default:
		return nil, errors.New("missing or unsupported action")
	case samvaad.UpdateSIPOutboundTrunkRequestAction:
		info, err = a.Apply(info)
		if err != nil {
			return nil, err
		}
	}
	// No additional validation needed for outbound.
	if err := s.store.StoreSIPOutboundTrunk(ctx, info); err != nil {
		return nil, err
	}
	return info, nil
}

func (s *SIPService) GetSIPInboundTrunk(ctx context.Context, req *samvaad.GetSIPInboundTrunkRequest) (*samvaad.GetSIPInboundTrunkResponse, error) {
	if err := EnsureSIPAdminPermission(ctx); err != nil {
		return nil, twirpAuthError(err)
	}
	if s.store == nil {
		return nil, ErrSIPNotConnected
	}
	if req.SipTrunkId == "" {
		return nil, twirp.NewError(twirp.InvalidArgument, "trunk ID is required")
	}
	AppendLogFields(ctx, "trunkID", req.SipTrunkId)

	trunk, err := s.store.LoadSIPInboundTrunk(ctx, req.SipTrunkId)
	if err != nil {
		if errors.Is(err, ErrSIPTrunkNotFound) {
			return nil, twirp.NewError(twirp.NotFound, err.Error())
		}
		return nil, err
	}

	return &samvaad.GetSIPInboundTrunkResponse{Trunk: trunk}, nil
}

func (s *SIPService) GetSIPOutboundTrunk(ctx context.Context, req *samvaad.GetSIPOutboundTrunkRequest) (*samvaad.GetSIPOutboundTrunkResponse, error) {
	if err := EnsureSIPAdminPermission(ctx); err != nil {
		return nil, twirpAuthError(err)
	}
	if s.store == nil {
		return nil, ErrSIPNotConnected
	}
	if req.SipTrunkId == "" {
		return nil, twirp.NewError(twirp.InvalidArgument, "trunk ID is required")
	}
	AppendLogFields(ctx, "trunkID", req.SipTrunkId)

	trunk, err := s.store.LoadSIPOutboundTrunk(ctx, req.SipTrunkId)
	if err != nil {
		if errors.Is(err, ErrSIPTrunkNotFound) {
			return nil, twirp.NewError(twirp.NotFound, err.Error())
		}
		return nil, err
	}

	return &samvaad.GetSIPOutboundTrunkResponse{Trunk: trunk}, nil
}

// deprecated: ListSIPTrunk will be removed in the future
func (s *SIPService) ListSIPTrunk(ctx context.Context, req *samvaad.ListSIPTrunkRequest) (*samvaad.ListSIPTrunkResponse, error) {
	if err := EnsureSIPAdminPermission(ctx); err != nil {
		return nil, twirpAuthError(err)
	}
	if s.store == nil {
		return nil, ErrSIPNotConnected
	}
	it := samvaad.ListPageIter(s.store.ListSIPTrunk, req)
	defer it.Close()

	items, err := iters.AllPages(ctx, it)
	if err != nil {
		return nil, err
	}
	return &samvaad.ListSIPTrunkResponse{Items: items}, nil
}

func ListSIPInboundTrunk(ctx context.Context, s SIPStore, req *samvaad.ListSIPInboundTrunkRequest, add ...*samvaad.SIPInboundTrunkInfo) (iters.Iter[*samvaad.SIPInboundTrunkInfo], error) {
	if s == nil {
		return nil, ErrSIPNotConnected
	}
	pages := samvaad.ListPageIter(s.ListSIPInboundTrunk, req)
	it := iters.PagesAsIter(ctx, pages)
	if len(add) != 0 {
		it = iters.MultiIter(true, it, iters.Slice(add))
	}
	return it, nil
}

func ListSIPOutboundTrunk(ctx context.Context, s SIPStore, req *samvaad.ListSIPOutboundTrunkRequest, add ...*samvaad.SIPOutboundTrunkInfo) (iters.Iter[*samvaad.SIPOutboundTrunkInfo], error) {
	if s == nil {
		return nil, ErrSIPNotConnected
	}
	pages := samvaad.ListPageIter(s.ListSIPOutboundTrunk, req)
	it := iters.PagesAsIter(ctx, pages)
	if len(add) != 0 {
		it = iters.MultiIter(true, it, iters.Slice(add))
	}
	return it, nil
}

func (s *SIPService) ListSIPInboundTrunk(ctx context.Context, req *samvaad.ListSIPInboundTrunkRequest) (*samvaad.ListSIPInboundTrunkResponse, error) {
	if err := EnsureSIPAdminPermission(ctx); err != nil {
		return nil, twirpAuthError(err)
	}
	if s.store == nil {
		return nil, ErrSIPNotConnected
	}
	it, err := ListSIPInboundTrunk(ctx, s.store, req)
	if err != nil {
		return nil, err
	}
	defer it.Close()

	items, err := iters.All(it)
	if err != nil {
		return nil, err
	}
	return &samvaad.ListSIPInboundTrunkResponse{Items: items}, nil
}

func (s *SIPService) ListSIPOutboundTrunk(ctx context.Context, req *samvaad.ListSIPOutboundTrunkRequest) (*samvaad.ListSIPOutboundTrunkResponse, error) {
	if err := EnsureSIPAdminPermission(ctx); err != nil {
		return nil, twirpAuthError(err)
	}
	if s.store == nil {
		return nil, ErrSIPNotConnected
	}
	it, err := ListSIPOutboundTrunk(ctx, s.store, req)
	if err != nil {
		return nil, err
	}
	defer it.Close()

	items, err := iters.All(it)
	if err != nil {
		return nil, err
	}
	return &samvaad.ListSIPOutboundTrunkResponse{Items: items}, nil
}

func (s *SIPService) DeleteSIPTrunk(ctx context.Context, req *samvaad.DeleteSIPTrunkRequest) (*samvaad.SIPTrunkInfo, error) {
	if err := EnsureSIPAdminPermission(ctx); err != nil {
		return nil, twirpAuthError(err)
	}
	if s.store == nil {
		return nil, ErrSIPNotConnected
	}
	if req.SipTrunkId == "" {
		return nil, twirp.NewError(twirp.InvalidArgument, "trunk ID is required")
	}

	AppendLogFields(ctx, "trunkID", req.SipTrunkId)
	if err := s.store.DeleteSIPTrunk(ctx, req.SipTrunkId); err != nil {
		return nil, err
	}

	return &samvaad.SIPTrunkInfo{SipTrunkId: req.SipTrunkId}, nil
}

func (s *SIPService) CreateSIPDispatchRule(ctx context.Context, req *samvaad.CreateSIPDispatchRuleRequest) (*samvaad.SIPDispatchRuleInfo, error) {
	if err := EnsureSIPAdminPermission(ctx); err != nil {
		return nil, twirpAuthError(err)
	}
	if s.store == nil {
		return nil, ErrSIPNotConnected
	}
	req.DispatchRule.Upgrade()
	if err := req.Validate(); err != nil {
		return nil, twirp.WrapError(twirp.NewError(twirp.InvalidArgument, err.Error()), err)
	}

	AppendLogFields(ctx,
		"request", logger.Proto(req),
		"trunkID", req.TrunkIds,
	)
	// Keep ID empty, so that validation can print "<new>" instead of a non-existent ID in the error.
	info := req.DispatchRuleInfo()
	info.SipDispatchRuleId = ""

	// Validate all rules including the new one first.
	it, err := ListSIPDispatchRule(ctx, s.store, &samvaad.ListSIPDispatchRuleRequest{
		TrunkIds: req.TrunkIds,
	}, info)
	if err != nil {
		return nil, err
	}
	defer it.Close()
	if _, err = sip.ValidateDispatchRulesIter(it); err != nil {
		return nil, err
	}

	// Now we can generate ID and store.
	info.SipDispatchRuleId = guid.New(utils.SIPDispatchRulePrefix)
	if err := s.store.StoreSIPDispatchRule(ctx, info); err != nil {
		return nil, err
	}
	return info, nil
}

func (s *SIPService) UpdateSIPDispatchRule(ctx context.Context, req *samvaad.UpdateSIPDispatchRuleRequest) (*samvaad.SIPDispatchRuleInfo, error) {
	if err := EnsureSIPAdminPermission(ctx); err != nil {
		return nil, twirpAuthError(err)
	}
	if s.store == nil {
		return nil, ErrSIPNotConnected
	}
	if err := req.Validate(); err != nil {
		return nil, twirp.WrapError(twirp.NewError(twirp.InvalidArgument, err.Error()), err)
	}

	AppendLogFields(ctx,
		"request", logger.Proto(req),
		"ruleID", req.SipDispatchRuleId,
	)

	// Validate all trunks including the new one first.
	info, err := s.store.LoadSIPDispatchRule(ctx, req.SipDispatchRuleId)
	if err != nil {
		if errors.Is(err, ErrSIPDispatchRuleNotFound) {
			return nil, twirp.NewError(twirp.NotFound, err.Error())
		}
		return nil, err
	}
	switch a := req.Action.(type) {
	default:
		return nil, errors.New("missing or unsupported action")
	case samvaad.UpdateSIPDispatchRuleRequestAction:
		info, err = a.Apply(info)
		if err != nil {
			return nil, err
		}
	}

	it, err := ListSIPDispatchRule(ctx, s.store, &samvaad.ListSIPDispatchRuleRequest{
		TrunkIds: info.TrunkIds,
	})
	if err != nil {
		return nil, err
	}
	defer it.Close()
	if _, err = sip.ValidateDispatchRulesIter(it, sip.WithDispatchRuleReplace(func(t *samvaad.SIPDispatchRuleInfo) *samvaad.SIPDispatchRuleInfo {
		if req.SipDispatchRuleId == t.SipDispatchRuleId {
			return info // updated one
		}
		return t
	})); err != nil {
		return nil, err
	}

	if err := s.store.StoreSIPDispatchRule(ctx, info); err != nil {
		return nil, err
	}
	return info, nil
}

func ListSIPDispatchRule(ctx context.Context, s SIPStore, req *samvaad.ListSIPDispatchRuleRequest, add ...*samvaad.SIPDispatchRuleInfo) (iters.Iter[*samvaad.SIPDispatchRuleInfo], error) {
	if s == nil {
		return nil, ErrSIPNotConnected
	}
	pages := samvaad.ListPageIter(s.ListSIPDispatchRule, req)
	it := iters.PagesAsIter(ctx, pages)
	if len(add) != 0 {
		it = iters.MultiIter(true, it, iters.Slice(add))
	}
	return it, nil
}

func (s *SIPService) ListSIPDispatchRule(ctx context.Context, req *samvaad.ListSIPDispatchRuleRequest) (*samvaad.ListSIPDispatchRuleResponse, error) {
	if err := EnsureSIPAdminPermission(ctx); err != nil {
		return nil, twirpAuthError(err)
	}
	if s.store == nil {
		return nil, ErrSIPNotConnected
	}
	it, err := ListSIPDispatchRule(ctx, s.store, req)
	if err != nil {
		return nil, err
	}
	defer it.Close()

	items, err := iters.All(it)
	if err != nil {
		return nil, err
	}
	return &samvaad.ListSIPDispatchRuleResponse{Items: items}, nil
}

func (s *SIPService) DeleteSIPDispatchRule(ctx context.Context, req *samvaad.DeleteSIPDispatchRuleRequest) (*samvaad.SIPDispatchRuleInfo, error) {
	if err := EnsureSIPAdminPermission(ctx); err != nil {
		return nil, twirpAuthError(err)
	}
	if s.store == nil {
		return nil, ErrSIPNotConnected
	}
	if req.SipDispatchRuleId == "" {
		return nil, twirp.NewError(twirp.InvalidArgument, "dispatch rule ID is required")
	}

	info, err := s.store.LoadSIPDispatchRule(ctx, req.SipDispatchRuleId)
	if err != nil {
		if errors.Is(err, ErrSIPDispatchRuleNotFound) {
			return nil, twirp.NewError(twirp.NotFound, err.Error())
		}
		return nil, err
	}

	if err = s.store.DeleteSIPDispatchRule(ctx, info.SipDispatchRuleId); err != nil {
		return nil, err
	}

	return info, nil
}

func (s *SIPService) CreateSIPParticipant(ctx context.Context, req *samvaad.CreateSIPParticipantRequest) (*samvaad.SIPParticipantInfo, error) {
	unlikelyLogger := logger.GetLogger().WithUnlikelyValues(
		"room", req.RoomName,
		"sipTrunk", req.SipTrunkId,
		"toUser", req.SipCallTo,
		"participant", req.ParticipantIdentity,
	)
	AppendLogFields(ctx,
		"room", req.RoomName,
		"participant", req.ParticipantIdentity,
		"toUser", req.SipCallTo,
		"trunkID", req.SipTrunkId,
	)
	ireq, err := s.CreateSIPParticipantRequest(ctx, req, "", "", "", "")
	if err != nil {
		unlikelyLogger.Errorw("cannot create sip participant request", err)
		return nil, wrapSIPContextError(err)
	}
	unlikelyLogger = unlikelyLogger.WithValues(
		"callID", ireq.SipCallId,
		"fromUser", ireq.Number,
		"toHost", ireq.Address,
	)
	AppendLogFields(ctx,
		"callID", ireq.SipCallId,
		"fromUser", ireq.Number,
		"toHost", ireq.Address,
	)

	// CreateSIPParticipant will wait for Samvaad Participant to be created and that can take some time.
	// Thus, we must set a higher deadline for it, if it's not set already.
	timeout := 30 * time.Second
	if req.WaitUntilAnswered {
		timeout = 80 * time.Second
	}
	if deadline, ok := ctx.Deadline(); ok {
		timeout = time.Until(deadline)
	} else {
		var cancel func()
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}
	resp, err := s.psrpcClient.CreateSIPParticipant(ctx, "", ireq, psrpc.WithRequestTimeout(timeout))
	if err != nil {
		unlikelyLogger.Errorw("cannot create sip participant", err)
		return nil, wrapSIPContextError(err)
	}
	return &samvaad.SIPParticipantInfo{
		ParticipantId:       resp.ParticipantId,
		ParticipantIdentity: resp.ParticipantIdentity,
		RoomName:            req.RoomName,
		SipCallId:           ireq.SipCallId,
	}, nil
}

func (s *SIPService) CreateSIPParticipantRequest(ctx context.Context, req *samvaad.CreateSIPParticipantRequest, projectID, host, wsUrl, token string) (*rpc.InternalCreateSIPParticipantRequest, error) {
	if err := EnsureSIPCallPermission(ctx); err != nil {
		return nil, twirpAuthError(err)
	}
	if s.store == nil {
		return nil, ErrSIPNotConnected
	}
	req.Upgrade()
	if err := req.Validate(); err != nil {
		return nil, twirp.WrapError(twirp.NewError(twirp.InvalidArgument, err.Error()), err)
	}
	callID := sip.NewCallID()
	log := logger.GetLogger().WithUnlikelyValues(
		"callID", callID,
		"room", req.RoomName,
		"sipTrunk", req.SipTrunkId,
		"toUser", req.SipCallTo,
	)
	if projectID != "" {
		log = log.WithValues("projectID", projectID)
	}

	var trunk *samvaad.SIPOutboundTrunkInfo
	if req.SipTrunkId != "" {
		var err error
		trunk, err = s.store.LoadSIPOutboundTrunk(ctx, req.SipTrunkId)
		if err != nil {
			log.Errorw("cannot get trunk to update sip participant", err)
			if errors.Is(err, ErrSIPTrunkNotFound) {
				return nil, twirp.NewError(twirp.NotFound, err.Error())
			}
			return nil, err
		}
	}
	if trunk != nil && trunk.FromHost != "" {
		host = trunk.FromHost
	} else if t := req.Trunk; t != nil && t.FromHost != "" {
		host = t.FromHost
	}
	return rpc.NewCreateSIPParticipantRequest(projectID, callID, host, wsUrl, token, req, trunk)
}

func (s *SIPService) TransferSIPParticipant(ctx context.Context, req *samvaad.TransferSIPParticipantRequest) (*emptypb.Empty, error) {
	log := logger.GetLogger().WithUnlikelyValues(
		"room", req.RoomName,
		"participant", req.ParticipantIdentity,
		"transferTo", req.TransferTo,
		"playDialtone", req.PlayDialtone,
	)
	AppendLogFields(ctx,
		"room", req.RoomName,
		"participant", req.ParticipantIdentity,
		"transferTo", req.TransferTo,
		"playDialtone", req.PlayDialtone,
	)

	ireq, err := s.transferSIPParticipantRequest(ctx, req)
	if err != nil {
		log.Errorw("cannot create transfer sip participant request", err)
		return nil, wrapSIPContextError(err)
	}

	// by default we set the timeout to be 30 seconds.
	// this timeout covers:
	//  - a network failure between this process and the Samvaad SIP bridge
	//  - the SIP transfer target not returning 200 OK fast enough.
	// WARN: any timeout/cancellation of a SIP transfer risks leaving
	// either the SIP bridge, or the SIP REFER exchange, in a "unknown" state.
	timeout := 30 * time.Second
	if req.RingingTimeout != nil {
		timeout = req.RingingTimeout.AsDuration()
	}

	// it's also possible the ctx has a Deadline.
	// in that case we want to use that deadline,
	// or our timeout, whichover is soonest.
	if deadline, ok := ctx.Deadline(); ok {
		timeout = min(timeout, time.Until(deadline))
	} else {
		var cancel func()
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}
	_, err = s.psrpcClient.TransferSIPParticipant(ctx, ireq.SipCallId, ireq, psrpc.WithRequestTimeout(timeout))
	if err != nil {
		log.Errorw("cannot transfer sip participant", err)
		return nil, wrapSIPContextError(err)
	}

	return &emptypb.Empty{}, nil
}

func (s *SIPService) transferSIPParticipantRequest(ctx context.Context, req *samvaad.TransferSIPParticipantRequest) (*rpc.InternalTransferSIPParticipantRequest, error) {
	if req.RoomName == "" {
		return nil, psrpc.NewErrorf(psrpc.InvalidArgument, "Missing room name")
	}

	if req.ParticipantIdentity == "" {
		return nil, psrpc.NewErrorf(psrpc.InvalidArgument, "Missing participant identity")
	}

	if err := EnsureSIPCallPermission(ctx); err != nil {
		return nil, twirpAuthError(err)
	}
	if err := EnsureAdminPermission(ctx, samvaad.RoomName(req.RoomName)); err != nil {
		return nil, twirpAuthError(err)
	}
	if err := req.Validate(); err != nil {
		return nil, twirp.WrapError(twirp.NewError(twirp.InvalidArgument, err.Error()), err)
	}

	resp, err := s.roomService.GetParticipant(ctx, &samvaad.RoomParticipantIdentity{
		Room:     req.RoomName,
		Identity: req.ParticipantIdentity,
	})

	if err != nil {
		return nil, err
	}

	callID, ok := resp.Attributes[samvaad.AttrSIPCallID]
	if !ok {
		return nil, psrpc.NewErrorf(psrpc.InvalidArgument, "no SIP session associated with participant")
	}

	return &rpc.InternalTransferSIPParticipantRequest{
		SipCallId:      callID,
		TransferTo:     req.TransferTo,
		PlayDialtone:   req.PlayDialtone,
		Headers:        req.Headers,
		RingingTimeout: req.RingingTimeout,
	}, nil
}

// wrapSIPContextError converts raw context.DeadlineExceeded / context.Canceled
// into psrpc-coded errors so they aren't surfaced as @code:unknown / HTTP 500
// at the Twirp boundary. psrpc errors and any error that already carries a
// gRPC status are passed through unchanged.
func wrapSIPContextError(err error) error {
	if err == nil {
		return nil
	}
	var psErr psrpc.Error
	if errors.As(err, &psErr) {
		return err
	}
	switch {
	case errors.Is(err, context.DeadlineExceeded):
		return psrpc.NewError(psrpc.DeadlineExceeded, err)
	case errors.Is(err, context.Canceled):
		return psrpc.NewError(psrpc.Canceled, err)
	}
	return err
}


