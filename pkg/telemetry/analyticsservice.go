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

package telemetry

import (
	"context"

	"go.uber.org/atomic"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/msmclass/samvaad/pkg/proto/samvaad"
	"github.com/msmclass/samvaad/pkg/samvaad/logger"
	"github.com/msmclass/samvaad/pkg/samvaad/observability/roomobs"
	"github.com/msmclass/samvaad/pkg/samvaad/rpc"
	"github.com/msmclass/samvaad/pkg/samvaad/utils/guid"

	"github.com/msmclass/samvaad/pkg/config"
	"github.com/msmclass/samvaad/pkg/routing"
)

//counterfeiter:generate . AnalyticsService
type AnalyticsService interface {
	SendStats(ctx context.Context, stats []*samvaad.AnalyticsStat)
	SendEvent(ctx context.Context, events *samvaad.AnalyticsEvent)
	SendNodeRoomStates(ctx context.Context, nodeRooms *samvaad.AnalyticsNodeRooms)
	RoomProjectReporter(ctx context.Context) roomobs.ProjectReporter
}

// ----------------------------

var _ AnalyticsService = &NullAnalyticService{}

type NullAnalyticService struct{}

func (n NullAnalyticService) SendStats(_ context.Context, _ []*samvaad.AnalyticsStat)             {}
func (n NullAnalyticService) SendEvent(_ context.Context, _ *samvaad.AnalyticsEvent)              {}
func (n NullAnalyticService) SendNodeRoomStates(_ context.Context, _ *samvaad.AnalyticsNodeRooms) {}
func (n NullAnalyticService) RoomProjectReporter(_ctx context.Context) roomobs.ProjectReporter {
	return nil
}

// ----------------------------

type analyticsService struct {
	analyticsKey   string
	nodeID         string
	sequenceNumber atomic.Uint64

	events    rpc.AnalyticsRecorderService_IngestEventsClient
	stats     rpc.AnalyticsRecorderService_IngestStatsClient
	nodeRooms rpc.AnalyticsRecorderService_IngestNodeRoomStatesClient
}

func NewAnalyticsService(_ *config.Config, currentNode routing.LocalNode) AnalyticsService {
	return &analyticsService{
		analyticsKey: "", // TODO: conf.AnalyticsKey
		nodeID:       string(currentNode.NodeID()),
	}
}

func (a *analyticsService) SendStats(_ context.Context, stats []*samvaad.AnalyticsStat) {
	if a.stats == nil {
		return
	}

	for _, stat := range stats {
		stat.Id = guid.New("AS_")
		stat.AnalyticsKey = a.analyticsKey
		stat.Node = a.nodeID
	}
	if err := a.stats.Send(&samvaad.AnalyticsStats{Stats: stats}); err != nil {
		logger.Errorw("failed to send stats", err)
	}
}

func (a *analyticsService) SendEvent(_ context.Context, event *samvaad.AnalyticsEvent) {
	if a.events == nil {
		return
	}

	event.Id = guid.New("AE_")
	event.NodeId = a.nodeID
	event.AnalyticsKey = a.analyticsKey
	if err := a.events.Send(&samvaad.AnalyticsEvents{
		Events: []*samvaad.AnalyticsEvent{event},
	}); err != nil {
		logger.Errorw("failed to send event", err, "eventType", event.Type.String())
	}
}

func (a *analyticsService) SendNodeRoomStates(_ context.Context, nodeRooms *samvaad.AnalyticsNodeRooms) {
	if a.nodeRooms == nil {
		return
	}

	nodeRooms.NodeId = a.nodeID
	nodeRooms.SequenceNumber = a.sequenceNumber.Add(1)
	nodeRooms.Timestamp = timestamppb.Now()
	if err := a.nodeRooms.Send(nodeRooms); err != nil {
		logger.Errorw("failed to send node room states", err)
	}
}

func (a *analyticsService) RoomProjectReporter(_ context.Context) roomobs.ProjectReporter {
	return roomobs.NewNoopProjectReporter()
}


