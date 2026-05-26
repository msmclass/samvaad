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
	"github.com/msmclass/samvaad/pkg/proto/samvaad"
	"github.com/msmclass/samvaad/pkg/samvaad/logger"
	"github.com/msmclass/samvaad/pkg/samvaad/observability/roomobs"
	"github.com/msmclass/samvaad/pkg/samvaad/utils"
	"github.com/msmclass/samvaad/pkg/samvaad/utils/guid"

	"github.com/msmclass/samvaad/pkg/rtc/types"
	"github.com/msmclass/samvaad/pkg/rtc/types/typesfakes"
)

func NewMockParticipant(
	identity samvaad.ParticipantIdentity,
	protocol types.ProtocolVersion,
	hidden bool,
	publisher bool,
	participantListener types.LocalParticipantListener,
) *typesfakes.FakeLocalParticipant {
	p := &typesfakes.FakeLocalParticipant{}
	sid := guid.New(utils.ParticipantPrefix)
	p.IDReturns(samvaad.ParticipantID(sid))
	p.IdentityReturns(identity)
	p.StateReturns(samvaad.ParticipantInfo_JOINED)
	p.ProtocolVersionReturns(protocol)
	p.CanSubscribeReturns(true)
	p.CanPublishSourceReturns(!hidden)
	p.CanPublishDataReturns(!hidden)
	p.HiddenReturns(hidden)
	p.ToProtoReturns(&samvaad.ParticipantInfo{
		Sid:         sid,
		Identity:    string(identity),
		State:       samvaad.ParticipantInfo_JOINED,
		IsPublisher: publisher,
	})
	p.ToProtoWithVersionReturns(&samvaad.ParticipantInfo{
		Sid:         sid,
		Identity:    string(identity),
		State:       samvaad.ParticipantInfo_JOINED,
		IsPublisher: publisher,
	}, utils.TimedVersion(0))

	p.SetMetadataCalls(func(m string) {
		participantListener.OnParticipantUpdate(p)
	})
	updateTrack := func() {
		participantListener.OnTrackUpdated(p, NewMockTrack(samvaad.TrackType_VIDEO, "testcam"))
	}

	p.SetTrackMutedCalls(func(mute *samvaad.MuteTrackRequest, fromServer bool) *samvaad.TrackInfo {
		updateTrack()
		return nil
	})
	p.AddTrackCalls(func(req *samvaad.AddTrackRequest) {
		updateTrack()
	})
	p.GetLoggerReturns(logger.GetLogger())
	p.GetReporterReturns(roomobs.NewNoopParticipantSessionReporter())

	return p
}

func NewMockTrack(kind samvaad.TrackType, name string) *typesfakes.FakeMediaTrack {
	t := &typesfakes.FakeMediaTrack{}
	t.IDReturns(samvaad.TrackID(guid.New(utils.TrackPrefix)))
	t.KindReturns(kind)
	t.NameReturns(name)
	t.ToProtoReturns(&samvaad.TrackInfo{
		Type: kind,
		Name: name,
	})
	return t
}


