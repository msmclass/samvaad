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

package signalling

import (
	"github.com/msmclass/samvaad/pkg/proto/samvaad"

	"github.com/msmclass/samvaad/pkg/routing"
	"github.com/msmclass/samvaad/pkg/rtc/types"

	"google.golang.org/protobuf/proto"
)

type ParticipantSignalHandler interface {
	HandleMessage(msg proto.Message) error
}

type ParticipantSignaller interface {
	SwapResponseSink(sink routing.MessageSink, reason types.SignallingCloseReason)
	GetResponseSink() routing.MessageSink
	CloseSignalConnection(reason types.SignallingCloseReason)

	WriteMessage(msg proto.Message) error
}

type ParticipantSignalling interface {
	SignalJoinResponse(join *samvaad.JoinResponse) proto.Message
	SignalParticipantUpdate(participants []*samvaad.ParticipantInfo) proto.Message
	SignalSpeakerUpdate(speakers []*samvaad.SpeakerInfo) proto.Message
	SignalRoomUpdate(room *samvaad.Room) proto.Message
	SignalConnectionQualityUpdate(connectionQuality *samvaad.ConnectionQualityUpdate) proto.Message
	SignalRefreshToken(token string) proto.Message
	SignalRequestResponse(requestResponse *samvaad.RequestResponse) proto.Message
	SignalRoomMovedResponse(roomMoved *samvaad.RoomMovedResponse) proto.Message
	SignalReconnectResponse(reconnect *samvaad.ReconnectResponse) proto.Message
	SignalICECandidate(trickle *samvaad.TrickleRequest) proto.Message
	SignalTrackMuted(mute *samvaad.MuteTrackRequest) proto.Message
	SignalTrackPublished(trackPublished *samvaad.TrackPublishedResponse) proto.Message
	SignalTrackUnpublished(trackUnpublished *samvaad.TrackUnpublishedResponse) proto.Message
	SignalTrackSubscribed(trackSubscribed *samvaad.TrackSubscribed) proto.Message
	SignalLeaveRequest(leave *samvaad.LeaveRequest) proto.Message
	SignalSdpAnswer(answer *samvaad.SessionDescription) proto.Message
	SignalSdpOffer(offer *samvaad.SessionDescription) proto.Message
	SignalStreamStateUpdate(streamStateUpdate *samvaad.StreamStateUpdate) proto.Message
	SignalSubscribedQualityUpdate(subscribedQualityUpdate *samvaad.SubscribedQualityUpdate) proto.Message
	SignalSubscriptionResponse(subscriptionResponse *samvaad.SubscriptionResponse) proto.Message
	SignalSubscriptionPermissionUpdate(subscriptionPermissionUpdate *samvaad.SubscriptionPermissionUpdate) proto.Message
	SignalMediaSectionsRequirement(mediaSectionsRequirement *samvaad.MediaSectionsRequirement) proto.Message
	SignalSubscribedAudioCodecUpdate(subscribedAudioCodecUpdate *samvaad.SubscribedAudioCodecUpdate) proto.Message
	SignalPublishDataTrackResponse(publishDataTrackResponse *samvaad.PublishDataTrackResponse) proto.Message
	SignalUnpublishDataTrackResponse(unpublishDataTrackResponse *samvaad.UnpublishDataTrackResponse) proto.Message
	SignalDataTrackSubscriberHandles(dataTrackSubscriberHandles *samvaad.DataTrackSubscriberHandles) proto.Message
}


