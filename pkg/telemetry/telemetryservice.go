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
	"sync"
	"time"

	"github.com/msmclass/samvaad/pkg/utils"
	"github.com/msmclass/samvaad/pkg/samvaad/codecs/mime"
	"github.com/msmclass/samvaad/pkg/proto/samvaad"
	"github.com/msmclass/samvaad/pkg/samvaad/logger"
	"github.com/msmclass/samvaad/pkg/samvaad/webhook"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

//counterfeiter:generate . TelemetryService
type TelemetryService interface {
	// TrackStats is called periodically for each track in both directions (published/subscribed)
	TrackStats(roomID samvaad.RoomID, roomName samvaad.RoomName, key StatsKey, stat *samvaad.AnalyticsStat)

	// events
	RoomStarted(ctx context.Context, room *samvaad.Room)
	RoomEnded(ctx context.Context, room *samvaad.Room)

	// ParticipantJoined - a participant establishes signal connection to a room
	ParticipantJoined(ctx context.Context, room *samvaad.Room, participant *samvaad.ParticipantInfo, clientInfo *samvaad.ClientInfo, clientMeta *samvaad.AnalyticsClientMeta, shouldSendEvent bool, guard *ReferenceGuard)
	// ParticipantActive - a participant establishes media connection
	ParticipantActive(ctx context.Context, room *samvaad.Room, participant *samvaad.ParticipantInfo, clientMeta *samvaad.AnalyticsClientMeta, isMigration bool, guard *ReferenceGuard)
	// ParticipantResumed - there has been an ICE restart or connection resume attempt, and we've received their signal connection
	ParticipantResumed(ctx context.Context, room *samvaad.Room, participant *samvaad.ParticipantInfo, nodeID samvaad.NodeID, reason samvaad.ReconnectReason)
	// ParticipantLeft - the participant leaves the room, only sent if ParticipantActive has been called before
	ParticipantLeft(ctx context.Context, room *samvaad.Room, participant *samvaad.ParticipantInfo, shouldSendEvent bool, guard *ReferenceGuard)
	// TrackPublishRequested - a publication attempt has been received
	TrackPublishRequested(ctx context.Context, roomID samvaad.RoomID, roomName samvaad.RoomName, participantID samvaad.ParticipantID, identity samvaad.ParticipantIdentity, track *samvaad.TrackInfo)
	// TrackPublished - a publication attempt has been successful
	TrackPublished(ctx context.Context, roomID samvaad.RoomID, roomName samvaad.RoomName, participantID samvaad.ParticipantID, identity samvaad.ParticipantIdentity, track *samvaad.TrackInfo, shouldSendEvent bool)
	// TrackUnpublished - a participant unpublished a track
	TrackUnpublished(ctx context.Context, roomID samvaad.RoomID, roomName samvaad.RoomName, participantID samvaad.ParticipantID, identity samvaad.ParticipantIdentity, track *samvaad.TrackInfo, shouldSendEvent bool)
	// TrackSubscribeRequested - a participant requested to subscribe to a track
	TrackSubscribeRequested(ctx context.Context, roomID samvaad.RoomID, roomName samvaad.RoomName, participantID samvaad.ParticipantID, track *samvaad.TrackInfo)
	// TrackSubscribed - a participant subscribed to a track successfully
	TrackSubscribed(ctx context.Context, roomID samvaad.RoomID, roomName samvaad.RoomName, participantID samvaad.ParticipantID, track *samvaad.TrackInfo, publisher *samvaad.ParticipantInfo, shouldSendEvent bool)
	// TrackUnsubscribed - a participant unsubscribed from a track successfully
	TrackUnsubscribed(ctx context.Context, roomID samvaad.RoomID, roomName samvaad.RoomName, participantID samvaad.ParticipantID, track *samvaad.TrackInfo, shouldSendEvent bool)
	// TrackSubscribeFailed - failure to subscribe to a track
	TrackSubscribeFailed(ctx context.Context, roomID samvaad.RoomID, roomName samvaad.RoomName, participantID samvaad.ParticipantID, trackID samvaad.TrackID, err error, isUserError bool)
	// TrackMuted - the publisher has muted the Track
	TrackMuted(ctx context.Context, roomID samvaad.RoomID, roomName samvaad.RoomName, participantID samvaad.ParticipantID, track *samvaad.TrackInfo)
	// TrackUnmuted - the publisher has muted the Track
	TrackUnmuted(ctx context.Context, roomID samvaad.RoomID, roomName samvaad.RoomName, participantID samvaad.ParticipantID, track *samvaad.TrackInfo)
	// TrackPublishedUpdate - track metadata has been updated
	TrackPublishedUpdate(ctx context.Context, roomID samvaad.RoomID, roomName samvaad.RoomName, participantID samvaad.ParticipantID, track *samvaad.TrackInfo)
	// TrackMaxSubscribedVideoQuality - publisher is notified of the max quality subscribers desire
	TrackMaxSubscribedVideoQuality(ctx context.Context, roomID samvaad.RoomID, roomName samvaad.RoomName, participantID samvaad.ParticipantID, track *samvaad.TrackInfo, mime mime.MimeType, maxQuality samvaad.VideoQuality)
	TrackPublishRTPStats(ctx context.Context, roomID samvaad.RoomID, roomName samvaad.RoomName, participantID samvaad.ParticipantID, trackID samvaad.TrackID, mimeType mime.MimeType, layer int, stats *samvaad.RTPStats)
	TrackSubscribeRTPStats(ctx context.Context, roomID samvaad.RoomID, roomName samvaad.RoomName, participantID samvaad.ParticipantID, trackID samvaad.TrackID, mimeType mime.MimeType, stats *samvaad.RTPStats)

	EgressStarted(ctx context.Context, info *samvaad.EgressInfo)
	EgressUpdated(ctx context.Context, info *samvaad.EgressInfo)
	EgressEnded(ctx context.Context, info *samvaad.EgressInfo)

	IngressCreated(ctx context.Context, info *samvaad.IngressInfo)
	IngressDeleted(ctx context.Context, info *samvaad.IngressInfo)
	IngressStarted(ctx context.Context, info *samvaad.IngressInfo)
	IngressUpdated(ctx context.Context, info *samvaad.IngressInfo)
	IngressEnded(ctx context.Context, info *samvaad.IngressInfo)

	LocalRoomState(ctx context.Context, info *samvaad.AnalyticsNodeRooms)

	Report(ctx context.Context, reportInfo *samvaad.ReportInfo)

	APICall(ctx context.Context, apiCallInfo *samvaad.APICallInfo)

	Webhook(ctx context.Context, webhookInfo *samvaad.WebhookInfo)

	// helpers
	AnalyticsService
	NotifyEgressEvent(ctx context.Context, event string, info *samvaad.EgressInfo)
	FlushStats()
}

// -----------------------------

var _ TelemetryService = (*NullTelemetryService)(nil)

type NullTelemetryService struct {
	NullAnalyticService
}

func (n NullTelemetryService) TrackStats(roomID samvaad.RoomID, roomName samvaad.RoomName, key StatsKey, stat *samvaad.AnalyticsStat) {
}
func (n NullTelemetryService) RoomStarted(ctx context.Context, room *samvaad.Room) {}
func (n NullTelemetryService) RoomEnded(ctx context.Context, room *samvaad.Room)   {}
func (n NullTelemetryService) ParticipantJoined(ctx context.Context, room *samvaad.Room, participant *samvaad.ParticipantInfo, clientInfo *samvaad.ClientInfo, clientMeta *samvaad.AnalyticsClientMeta, shouldSendEvent bool, guard *ReferenceGuard) {
}
func (n NullTelemetryService) ParticipantActive(ctx context.Context, room *samvaad.Room, participant *samvaad.ParticipantInfo, clientMeta *samvaad.AnalyticsClientMeta, isMigration bool, guard *ReferenceGuard) {
}
func (n NullTelemetryService) ParticipantResumed(ctx context.Context, room *samvaad.Room, participant *samvaad.ParticipantInfo, nodeID samvaad.NodeID, reason samvaad.ReconnectReason) {
}
func (n NullTelemetryService) ParticipantLeft(ctx context.Context, room *samvaad.Room, participant *samvaad.ParticipantInfo, shouldSendEvent bool, guard *ReferenceGuard) {
}
func (n NullTelemetryService) TrackPublishRequested(ctx context.Context, roomID samvaad.RoomID, roomName samvaad.RoomName, participantID samvaad.ParticipantID, identity samvaad.ParticipantIdentity, track *samvaad.TrackInfo) {
}
func (n NullTelemetryService) TrackPublished(ctx context.Context, roomID samvaad.RoomID, roomName samvaad.RoomName, participantID samvaad.ParticipantID, identity samvaad.ParticipantIdentity, track *samvaad.TrackInfo, shouldSendEvent bool) {
}
func (n NullTelemetryService) TrackUnpublished(ctx context.Context, roomID samvaad.RoomID, roomName samvaad.RoomName, participantID samvaad.ParticipantID, identity samvaad.ParticipantIdentity, track *samvaad.TrackInfo, shouldSendEvent bool) {
}
func (n NullTelemetryService) TrackSubscribeRequested(ctx context.Context, roomID samvaad.RoomID, roomName samvaad.RoomName, participantID samvaad.ParticipantID, track *samvaad.TrackInfo) {
}
func (n NullTelemetryService) TrackSubscribed(ctx context.Context, roomID samvaad.RoomID, roomName samvaad.RoomName, participantID samvaad.ParticipantID, track *samvaad.TrackInfo, publisher *samvaad.ParticipantInfo, shouldSendEvent bool) {
}
func (n NullTelemetryService) TrackUnsubscribed(ctx context.Context, roomID samvaad.RoomID, roomName samvaad.RoomName, participantID samvaad.ParticipantID, track *samvaad.TrackInfo, shouldSendEvent bool) {
}
func (n NullTelemetryService) TrackSubscribeFailed(ctx context.Context, roomID samvaad.RoomID, roomName samvaad.RoomName, participantID samvaad.ParticipantID, trackID samvaad.TrackID, err error, isUserError bool) {
}
func (n NullTelemetryService) TrackMuted(ctx context.Context, roomID samvaad.RoomID, roomName samvaad.RoomName, participantID samvaad.ParticipantID, track *samvaad.TrackInfo) {
}
func (n NullTelemetryService) TrackUnmuted(ctx context.Context, roomID samvaad.RoomID, roomName samvaad.RoomName, participantID samvaad.ParticipantID, track *samvaad.TrackInfo) {
}
func (n NullTelemetryService) TrackPublishedUpdate(ctx context.Context, roomID samvaad.RoomID, roomName samvaad.RoomName, participantID samvaad.ParticipantID, track *samvaad.TrackInfo) {
}
func (n NullTelemetryService) TrackMaxSubscribedVideoQuality(ctx context.Context, roomID samvaad.RoomID, roomName samvaad.RoomName, participantID samvaad.ParticipantID, track *samvaad.TrackInfo, mime mime.MimeType, maxQuality samvaad.VideoQuality) {
}
func (n NullTelemetryService) TrackPublishRTPStats(ctx context.Context, roomID samvaad.RoomID, roomName samvaad.RoomName, participantID samvaad.ParticipantID, trackID samvaad.TrackID, mimeType mime.MimeType, layer int, stats *samvaad.RTPStats) {
}
func (n NullTelemetryService) TrackSubscribeRTPStats(ctx context.Context, roomID samvaad.RoomID, roomName samvaad.RoomName, participantID samvaad.ParticipantID, trackID samvaad.TrackID, mimeType mime.MimeType, stats *samvaad.RTPStats) {
}
func (n NullTelemetryService) EgressStarted(ctx context.Context, info *samvaad.EgressInfo)          {}
func (n NullTelemetryService) EgressUpdated(ctx context.Context, info *samvaad.EgressInfo)          {}
func (n NullTelemetryService) EgressEnded(ctx context.Context, info *samvaad.EgressInfo)            {}
func (n NullTelemetryService) IngressCreated(ctx context.Context, info *samvaad.IngressInfo)        {}
func (n NullTelemetryService) IngressDeleted(ctx context.Context, info *samvaad.IngressInfo)        {}
func (n NullTelemetryService) IngressStarted(ctx context.Context, info *samvaad.IngressInfo)        {}
func (n NullTelemetryService) IngressUpdated(ctx context.Context, info *samvaad.IngressInfo)        {}
func (n NullTelemetryService) IngressEnded(ctx context.Context, info *samvaad.IngressInfo)          {}
func (n NullTelemetryService) LocalRoomState(ctx context.Context, info *samvaad.AnalyticsNodeRooms) {}
func (n NullTelemetryService) Report(ctx context.Context, reportInfo *samvaad.ReportInfo)           {}
func (n NullTelemetryService) APICall(ctx context.Context, apiCallInfo *samvaad.APICallInfo)        {}
func (n NullTelemetryService) Webhook(ctx context.Context, webhookInfo *samvaad.WebhookInfo)        {}
func (n NullTelemetryService) NotifyEgressEvent(ctx context.Context, event string, info *samvaad.EgressInfo) {
}
func (n NullTelemetryService) FlushStats() {}

// -----------------------------

const (
	workerCleanupWait = 3 * time.Minute
	jobsQueueMinSize  = 2048

	telemetryStatsUpdateInterval = time.Second * 30
)

type statsWorkerKey struct {
	roomID        samvaad.RoomID
	participantID samvaad.ParticipantID
}

type telemetryService struct {
	AnalyticsService

	notifier  webhook.QueuedNotifier
	jobsQueue *utils.OpsQueue

	workersMu  sync.RWMutex
	workers    map[statsWorkerKey]*StatsWorker
	workerList *StatsWorker

	flushMu sync.Mutex
}

func NewTelemetryService(notifier webhook.QueuedNotifier, analytics AnalyticsService) TelemetryService {
	t := &telemetryService{
		AnalyticsService: analytics,
		notifier:         notifier,
		jobsQueue: utils.NewOpsQueue(utils.OpsQueueParams{
			Name:        "telemetry",
			MinSize:     jobsQueueMinSize,
			FlushOnStop: true,
			Logger:      logger.GetLogger(),
		}),
		workers: make(map[statsWorkerKey]*StatsWorker),
	}
	if t.notifier != nil {
		t.notifier.RegisterProcessedHook(func(ctx context.Context, whi *samvaad.WebhookInfo) {
			t.Webhook(ctx, whi)
		})
	}

	t.jobsQueue.Start()
	go t.run()

	return t
}

func (t *telemetryService) FlushStats() {
	t.flushMu.Lock()
	defer t.flushMu.Unlock()

	t.workersMu.RLock()
	worker := t.workerList
	t.workersMu.RUnlock()

	now := time.Now()
	var prev, reap *StatsWorker
	for worker != nil {
		next := worker.next
		if closed := worker.Flush(now, workerCleanupWait); closed {
			if prev == nil {
				// this worker was at the head of the list
				t.workersMu.Lock()
				p := &t.workerList
				for *p != worker {
					// new workers have been added. scan until we find the one
					// immediately before this
					prev = *p
					p = &prev.next
				}
				*p = worker.next
				t.workersMu.Unlock()
			} else {
				prev.next = worker.next
			}

			worker.next = reap
			reap = worker
		} else {
			prev = worker
		}
		worker = next
	}

	if reap != nil {
		t.workersMu.Lock()
		for reap != nil {
			key := statsWorkerKey{reap.roomID, reap.participantID}
			if reap == t.workers[key] {
				delete(t.workers, key)
			}
			reap = reap.next
		}
		t.workersMu.Unlock()
	}
}

func (t *telemetryService) run() {
	for range time.Tick(telemetryStatsUpdateInterval) {
		t.FlushStats()
	}
}

func (t *telemetryService) enqueue(op func()) {
	t.jobsQueue.Enqueue(op)
}

func (t *telemetryService) getWorker(roomID samvaad.RoomID, participantID samvaad.ParticipantID) (worker *StatsWorker, ok bool) {
	t.workersMu.RLock()
	defer t.workersMu.RUnlock()

	worker, ok = t.workers[statsWorkerKey{roomID, participantID}]
	return
}

func (t *telemetryService) getOrCreateWorker(
	ctx context.Context,
	roomID samvaad.RoomID,
	roomName samvaad.RoomName,
	participantID samvaad.ParticipantID,
	participantIdentity samvaad.ParticipantIdentity,
	guard *ReferenceGuard,
) (*StatsWorker, bool) {
	t.workersMu.Lock()
	defer t.workersMu.Unlock()

	key := statsWorkerKey{roomID, participantID}
	worker, ok := t.workers[key]
	if ok && !worker.Closed(guard) {
		return worker, true
	}

	existingIsConnected := false
	if ok {
		existingIsConnected = worker.IsConnected()
	}

	worker = newStatsWorker(
		ctx,
		t,
		roomID,
		roomName,
		participantID,
		participantIdentity,
		guard,
	)
	if existingIsConnected {
		worker.SetConnected()
	}

	t.workers[key] = worker

	worker.next = t.workerList
	t.workerList = worker

	return worker, false
}

func (t *telemetryService) LocalRoomState(ctx context.Context, info *samvaad.AnalyticsNodeRooms) {
	t.enqueue(func() {
		t.SendNodeRoomStates(ctx, info)
	})
}


