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

package rtc

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/frostbyte73/core"
	"go.uber.org/atomic"

	"github.com/msmclass/samvaad/pkg/rtc/types"
	"github.com/msmclass/samvaad/pkg/telemetry"
	"github.com/msmclass/samvaad/pkg/proto/samvaad"
	"github.com/msmclass/samvaad/pkg/samvaad/observability/roomobs"
	"github.com/msmclass/samvaad/pkg/samvaad/utils"
)

type BytesTrackType string

const (
	BytesTrackTypeData   BytesTrackType = "DT"
	BytesTrackTypeSignal BytesTrackType = "SG"
)

// -------------------------------

type TrafficTotals struct {
	At           time.Time
	SendBytes    uint64
	SendMessages uint32
	RecvBytes    uint64
	RecvMessages uint32
}

// --------------------------------

// stats for signal and data channel
type BytesTrackStats struct {
	country                              string
	pID                                  samvaad.ParticipantID
	kind                                 samvaad.ParticipantInfo_Kind
	kindDetails                          []samvaad.ParticipantInfo_KindDetail
	trackID                              samvaad.TrackID
	send, recv                           atomic.Uint64
	sendMessages, recvMessages           atomic.Uint32
	totalSendBytes, totalRecvBytes       atomic.Uint64
	totalSendMessages, totalRecvMessages atomic.Uint32
	telemetryListener                    types.ParticipantTelemetryListener
	reporter                             roomobs.TrackReporter
	done                                 core.Fuse
}

func NewBytesTrackStats(
	country string,
	trackID samvaad.TrackID,
	pID samvaad.ParticipantID,
	kind samvaad.ParticipantInfo_Kind,
	kindDetails []samvaad.ParticipantInfo_KindDetail,
	telemetryListener types.ParticipantTelemetryListener,
	participantReporter roomobs.ParticipantSessionReporter,
) *BytesTrackStats {
	s := &BytesTrackStats{
		country:           country,
		pID:               pID,
		kind:              kind,
		kindDetails:       kindDetails,
		trackID:           trackID,
		telemetryListener: telemetryListener,
		reporter:          participantReporter.WithTrack(trackID.String()),
	}
	go s.worker()
	return s
}

func (s *BytesTrackStats) AddBytes(bytes uint64, isSend bool) {
	if isSend {
		s.send.Add(bytes)
		s.sendMessages.Inc()
		s.totalSendBytes.Add(bytes)
		s.totalSendMessages.Inc()

		s.reporter.Tx(func(tx roomobs.TrackTx) {
			tx.ParticipantSession().ReportKindCode(roomobs.ParticipantKindCode(s.kind))
			tx.ParticipantSession().ReportKindDetailsCodes(roomobs.ParticipantKindDetailsCodes(s.kindDetails))
			tx.ReportType(roomobs.TrackTypeData)
			tx.ReportSendBytes(uint32(bytes))
			tx.ReportSendPackets(1)
		})
	} else {
		s.recv.Add(bytes)
		s.recvMessages.Inc()
		s.totalRecvBytes.Add(bytes)
		s.totalRecvMessages.Inc()

		s.reporter.Tx(func(tx roomobs.TrackTx) {
			tx.ParticipantSession().ReportKindCode(roomobs.ParticipantKindCode(s.kind))
			tx.ParticipantSession().ReportKindDetailsCodes(roomobs.ParticipantKindDetailsCodes(s.kindDetails))
			tx.ReportType(roomobs.TrackTypeData)
			tx.ReportRecvBytes(uint32(bytes))
			tx.ReportRecvPackets(1)
		})
	}
}

func (s *BytesTrackStats) GetTrafficTotals() *TrafficTotals {
	return &TrafficTotals{
		At:           time.Now(),
		SendBytes:    s.totalSendBytes.Load(),
		SendMessages: s.totalSendMessages.Load(),
		RecvBytes:    s.totalRecvBytes.Load(),
		RecvMessages: s.totalRecvMessages.Load(),
	}
}

func (s *BytesTrackStats) Stop() {
	s.done.Break()
}

func (s *BytesTrackStats) report() {
	if recv := s.recv.Swap(0); recv > 0 {
		packets := s.recvMessages.Swap(0)
		s.telemetryListener.OnTrackStats(
			telemetry.StatsKeyForData(s.country, samvaad.StreamType_UPSTREAM, s.pID, s.trackID),
			&samvaad.AnalyticsStat{
				Streams: []*samvaad.AnalyticsStream{
					{
						PrimaryBytes:   recv,
						PrimaryPackets: packets,
					},
				},
			},
		)
	}

	if send := s.send.Swap(0); send > 0 {
		packets := s.sendMessages.Swap(0)
		s.telemetryListener.OnTrackStats(
			telemetry.StatsKeyForData(s.country, samvaad.StreamType_DOWNSTREAM, s.pID, s.trackID),
			&samvaad.AnalyticsStat{
				Streams: []*samvaad.AnalyticsStream{
					{
						PrimaryBytes:   send,
						PrimaryPackets: packets,
					},
				},
			},
		)
	}
}

func (s *BytesTrackStats) worker() {
	ticker := time.NewTicker(5 * time.Second)
	defer func() {
		ticker.Stop()
		s.report()
	}()

	for {
		select {
		case <-s.done.Watch():
			return
		case <-ticker.C:
			s.report()
		}
	}
}

// -----------------------------------------------------------------------

var _ types.ParticipantTelemetryListener = (*BytesSignalStats)(nil)

type BytesSignalStats struct {
	BytesTrackStats
	ctx context.Context

	telemetry telemetry.TelemetryService
	guard     *telemetry.ReferenceGuard

	participantResolver roomobs.ParticipantReporterResolver
	trackResolver       roomobs.KeyResolver

	mu      sync.Mutex
	ri      *samvaad.Room
	pi      *samvaad.ParticipantInfo
	stopped chan struct{}

	types.NullParticipantTelemetryListener
}

func NewBytesSignalStats(
	ctx context.Context,
	telemetryService telemetry.TelemetryService,
) *BytesSignalStats {
	projectReporter := telemetryService.RoomProjectReporter(ctx)
	participantReporter, participantReporterResolver := roomobs.DeferredParticipantReporter(projectReporter)
	trackReporter, trackReporterResolver := participantReporter.WithDeferredTrack()
	b := &BytesSignalStats{
		ctx:                 ctx,
		telemetry:           telemetryService,
		guard:               &telemetry.ReferenceGuard{},
		participantResolver: participantReporterResolver,
		trackResolver:       trackReporterResolver,
	}
	b.BytesTrackStats = BytesTrackStats{
		telemetryListener: b,
		reporter:          trackReporter,
	}
	return b
}

func (s *BytesSignalStats) ResolveRoom(ri *samvaad.Room) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.ri == nil && ri.GetSid() != "" {
		s.ri = &samvaad.Room{
			Sid:  ri.Sid,
			Name: ri.Name,
		}
		s.maybeStart()
	}
}

func (s *BytesSignalStats) ResolveParticipant(pi *samvaad.ParticipantInfo) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.pi == nil && pi != nil {
		s.pi = &samvaad.ParticipantInfo{
			Sid:         pi.Sid,
			Identity:    pi.Identity,
			Kind:        pi.Kind,
			KindDetails: pi.KindDetails,
		}
		s.kind = pi.Kind
		s.kindDetails = pi.KindDetails
		s.maybeStart()
	}
}

func (s *BytesSignalStats) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.stopped != nil {
		s.done.Break()
		<-s.stopped
		s.stopped = nil
		s.done = core.Fuse{}
		s.guard = &telemetry.ReferenceGuard{}
	}
	s.ri = nil
	s.pi = nil

	s.participantResolver.Reset()
	s.trackResolver.Reset()
}

func (s *BytesSignalStats) maybeStart() {
	if s.ri == nil || s.pi == nil {
		return
	}

	s.pID = samvaad.ParticipantID(s.pi.Sid)
	s.trackID = BytesTrackIDForParticipantID(BytesTrackTypeSignal, s.pID)

	s.participantResolver.Resolve(
		samvaad.RoomName(s.ri.Name),
		samvaad.RoomID(s.ri.Sid),
		samvaad.ParticipantIdentity(s.pi.Identity),
		samvaad.ParticipantID(s.pi.Sid),
	)
	s.trackResolver.Resolve(string(s.trackID))

	s.telemetry.ParticipantJoined(s.ctx, s.ri, s.pi, nil, nil, false, s.guard)
	s.stopped = make(chan struct{})
	go s.worker()
}

func (s *BytesSignalStats) worker() {
	s.BytesTrackStats.worker()
	s.telemetry.ParticipantLeft(s.ctx, s.ri, s.pi, false, s.guard)
	close(s.stopped)
}

func (s *BytesSignalStats) OnTrackStats(key telemetry.StatsKey, stat *samvaad.AnalyticsStat) {
	stat.RoomId, stat.RoomName = s.ri.Sid, s.ri.Name
	s.telemetry.TrackStats(samvaad.RoomID(s.ri.Sid), samvaad.RoomName(s.ri.Name), key, stat)
}

// -----------------------------------------------------------------------

func BytesTrackIDForParticipantID(typ BytesTrackType, participantID samvaad.ParticipantID) samvaad.TrackID {
	return samvaad.TrackID(fmt.Sprintf("%s%s%s", utils.TrackPrefix, typ, participantID))
}


