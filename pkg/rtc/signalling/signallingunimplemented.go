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

	"google.golang.org/protobuf/proto"
)

var _ ParticipantSignalling = (*signallingUnimplemented)(nil)

type signallingUnimplemented struct{}

func (u *signallingUnimplemented) SignalJoinResponse(join *samvaad.JoinResponse) proto.Message {
	return nil
}

func (u *signallingUnimplemented) SignalParticipantUpdate(participants []*samvaad.ParticipantInfo) proto.Message {
	return nil
}

func (u *signallingUnimplemented) SignalSpeakerUpdate(speakers []*samvaad.SpeakerInfo) proto.Message {
	return nil
}

func (u *signallingUnimplemented) SignalRoomUpdate(room *samvaad.Room) proto.Message {
	return nil
}

func (u *signallingUnimplemented) SignalConnectionQualityUpdate(connectionQuality *samvaad.ConnectionQualityUpdate) proto.Message {
	return nil
}

func (u *signallingUnimplemented) SignalRefreshToken(token string) proto.Message {
	return nil
}

func (u *signallingUnimplemented) SignalRequestResponse(requestResponse *samvaad.RequestResponse) proto.Message {
	return nil
}

func (u *signallingUnimplemented) SignalRoomMovedResponse(roomMoved *samvaad.RoomMovedResponse) proto.Message {
	return nil
}

func (u *signallingUnimplemented) SignalReconnectResponse(reconnect *samvaad.ReconnectResponse) proto.Message {
	return nil
}

func (u *signallingUnimplemented) SignalICECandidate(trickle *samvaad.TrickleRequest) proto.Message {
	return nil
}

func (u *signallingUnimplemented) SignalTrackMuted(mute *samvaad.MuteTrackRequest) proto.Message {
	return nil
}

func (u *signallingUnimplemented) SignalTrackPublished(trackPublished *samvaad.TrackPublishedResponse) proto.Message {
	return nil
}

func (u *signallingUnimplemented) SignalTrackUnpublished(trackUnpublished *samvaad.TrackUnpublishedResponse) proto.Message {
	return nil
}

func (u *signallingUnimplemented) SignalTrackSubscribed(trackSubscribed *samvaad.TrackSubscribed) proto.Message {
	return nil
}

func (u *signallingUnimplemented) SignalLeaveRequest(leave *samvaad.LeaveRequest) proto.Message {
	return nil
}

func (u *signallingUnimplemented) SignalSdpAnswer(answer *samvaad.SessionDescription) proto.Message {
	return nil
}

func (u *signallingUnimplemented) SignalSdpOffer(offer *samvaad.SessionDescription) proto.Message {
	return nil
}

func (u *signallingUnimplemented) SignalStreamStateUpdate(streamStateUpdate *samvaad.StreamStateUpdate) proto.Message {
	return nil
}

func (u *signallingUnimplemented) SignalSubscribedQualityUpdate(subscribedQualityUpdate *samvaad.SubscribedQualityUpdate) proto.Message {
	return nil
}

func (u *signallingUnimplemented) SignalSubscriptionResponse(subscriptionResponse *samvaad.SubscriptionResponse) proto.Message {
	return nil
}

func (u *signallingUnimplemented) SignalSubscriptionPermissionUpdate(subscriptionPermissionUpdate *samvaad.SubscriptionPermissionUpdate) proto.Message {
	return nil
}

func (u *signallingUnimplemented) SignalMediaSectionsRequirement(mediaSectionsRequirement *samvaad.MediaSectionsRequirement) proto.Message {
	return nil
}

func (u *signallingUnimplemented) SignalSubscribedAudioCodecUpdate(subscribedAudioCodecUpdate *samvaad.SubscribedAudioCodecUpdate) proto.Message {
	return nil
}

func (u *signallingUnimplemented) SignalPublishDataTrackResponse(publishDataTrackResponse *samvaad.PublishDataTrackResponse) proto.Message {
	return nil
}

func (u *signallingUnimplemented) SignalUnpublishDataTrackResponse(unpublishDataTrackResponse *samvaad.UnpublishDataTrackResponse) proto.Message {
	return nil
}

func (u *signallingUnimplemented) SignalDataTrackSubscriberHandles(dataTrackSubscriberHandles *samvaad.DataTrackSubscriberHandles) proto.Message {
	return nil
}


