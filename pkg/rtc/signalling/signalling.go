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
	"github.com/msmclass/samvaad/pkg/samvaad/logger"

	"google.golang.org/protobuf/proto"
)

var _ ParticipantSignalling = (*signalling)(nil)

type SignallingParams struct {
	Logger logger.Logger
}

type signalling struct {
	params SignallingParams
}

func NewSignalling(params SignallingParams) ParticipantSignalling {
	return &signalling{
		params: params,
	}
}

func (s *signalling) SignalJoinResponse(join *samvaad.JoinResponse) proto.Message {
	return &samvaad.SignalResponse{
		Message: &samvaad.SignalResponse_Join{
			Join: join,
		},
	}
}

func (s *signalling) SignalParticipantUpdate(participants []*samvaad.ParticipantInfo) proto.Message {
	if len(participants) == 0 {
		return nil
	}

	return &samvaad.SignalResponse{
		Message: &samvaad.SignalResponse_Update{
			Update: &samvaad.ParticipantUpdate{
				Participants: participants,
			},
		},
	}
}

func (s *signalling) SignalSpeakerUpdate(speakers []*samvaad.SpeakerInfo) proto.Message {
	if len(speakers) == 0 {
		return nil
	}

	return &samvaad.SignalResponse{
		Message: &samvaad.SignalResponse_SpeakersChanged{
			SpeakersChanged: &samvaad.SpeakersChanged{
				Speakers: speakers,
			},
		},
	}
}

func (s *signalling) SignalRoomUpdate(room *samvaad.Room) proto.Message {
	return &samvaad.SignalResponse{
		Message: &samvaad.SignalResponse_RoomUpdate{
			RoomUpdate: &samvaad.RoomUpdate{
				Room: room,
			},
		},
	}
}

func (s *signalling) SignalConnectionQualityUpdate(connectionQuality *samvaad.ConnectionQualityUpdate) proto.Message {
	return &samvaad.SignalResponse{
		Message: &samvaad.SignalResponse_ConnectionQuality{
			ConnectionQuality: connectionQuality,
		},
	}
}

func (s *signalling) SignalRefreshToken(token string) proto.Message {
	return &samvaad.SignalResponse{
		Message: &samvaad.SignalResponse_RefreshToken{
			RefreshToken: token,
		},
	}
}

func (s *signalling) SignalRequestResponse(requestResponse *samvaad.RequestResponse) proto.Message {
	return &samvaad.SignalResponse{
		Message: &samvaad.SignalResponse_RequestResponse{
			RequestResponse: requestResponse,
		},
	}
}

func (s *signalling) SignalRoomMovedResponse(roomMoved *samvaad.RoomMovedResponse) proto.Message {
	return &samvaad.SignalResponse{
		Message: &samvaad.SignalResponse_RoomMoved{
			RoomMoved: roomMoved,
		},
	}
}

func (s *signalling) SignalReconnectResponse(reconnect *samvaad.ReconnectResponse) proto.Message {
	return &samvaad.SignalResponse{
		Message: &samvaad.SignalResponse_Reconnect{
			Reconnect: reconnect,
		},
	}
}

func (s *signalling) SignalICECandidate(trickle *samvaad.TrickleRequest) proto.Message {
	return &samvaad.SignalResponse{
		Message: &samvaad.SignalResponse_Trickle{
			Trickle: trickle,
		},
	}
}

func (s *signalling) SignalTrackMuted(mute *samvaad.MuteTrackRequest) proto.Message {
	return &samvaad.SignalResponse{
		Message: &samvaad.SignalResponse_Mute{
			Mute: mute,
		},
	}
}

func (s *signalling) SignalTrackPublished(trackPublished *samvaad.TrackPublishedResponse) proto.Message {
	return &samvaad.SignalResponse{
		Message: &samvaad.SignalResponse_TrackPublished{
			TrackPublished: trackPublished,
		},
	}
}

func (s *signalling) SignalTrackUnpublished(trackUnpublished *samvaad.TrackUnpublishedResponse) proto.Message {
	return &samvaad.SignalResponse{
		Message: &samvaad.SignalResponse_TrackUnpublished{
			TrackUnpublished: trackUnpublished,
		},
	}
}

func (s *signalling) SignalTrackSubscribed(trackSubscribed *samvaad.TrackSubscribed) proto.Message {
	return &samvaad.SignalResponse{
		Message: &samvaad.SignalResponse_TrackSubscribed{
			TrackSubscribed: trackSubscribed,
		},
	}
}

func (s *signalling) SignalLeaveRequest(leave *samvaad.LeaveRequest) proto.Message {
	return &samvaad.SignalResponse{
		Message: &samvaad.SignalResponse_Leave{
			Leave: leave,
		},
	}
}

func (s *signalling) SignalSdpAnswer(answer *samvaad.SessionDescription) proto.Message {
	return &samvaad.SignalResponse{
		Message: &samvaad.SignalResponse_Answer{
			Answer: answer,
		},
	}
}

func (s *signalling) SignalSdpOffer(offer *samvaad.SessionDescription) proto.Message {
	return &samvaad.SignalResponse{
		Message: &samvaad.SignalResponse_Offer{
			Offer: offer,
		},
	}
}

func (s *signalling) SignalStreamStateUpdate(streamStateUpdate *samvaad.StreamStateUpdate) proto.Message {
	return &samvaad.SignalResponse{
		Message: &samvaad.SignalResponse_StreamStateUpdate{
			StreamStateUpdate: streamStateUpdate,
		},
	}
}

func (s *signalling) SignalSubscribedQualityUpdate(subscribedQualityUpdate *samvaad.SubscribedQualityUpdate) proto.Message {
	return &samvaad.SignalResponse{
		Message: &samvaad.SignalResponse_SubscribedQualityUpdate{
			SubscribedQualityUpdate: subscribedQualityUpdate,
		},
	}
}

func (s *signalling) SignalSubscriptionResponse(subscriptionResponse *samvaad.SubscriptionResponse) proto.Message {
	return &samvaad.SignalResponse{
		Message: &samvaad.SignalResponse_SubscriptionResponse{
			SubscriptionResponse: subscriptionResponse,
		},
	}
}

func (s *signalling) SignalSubscriptionPermissionUpdate(subscriptionPermissionUpdate *samvaad.SubscriptionPermissionUpdate) proto.Message {
	return &samvaad.SignalResponse{
		Message: &samvaad.SignalResponse_SubscriptionPermissionUpdate{
			SubscriptionPermissionUpdate: subscriptionPermissionUpdate,
		},
	}
}

func (s *signalling) SignalMediaSectionsRequirement(mediaSectionsRequirement *samvaad.MediaSectionsRequirement) proto.Message {
	return &samvaad.SignalResponse{
		Message: &samvaad.SignalResponse_MediaSectionsRequirement{
			MediaSectionsRequirement: mediaSectionsRequirement,
		},
	}
}

func (s *signalling) SignalSubscribedAudioCodecUpdate(subscribedAudioCodecUpdate *samvaad.SubscribedAudioCodecUpdate) proto.Message {
	return &samvaad.SignalResponse{
		Message: &samvaad.SignalResponse_SubscribedAudioCodecUpdate{
			SubscribedAudioCodecUpdate: subscribedAudioCodecUpdate,
		},
	}
}

func (s *signalling) SignalPublishDataTrackResponse(publishDataTrackResponse *samvaad.PublishDataTrackResponse) proto.Message {
	return &samvaad.SignalResponse{
		Message: &samvaad.SignalResponse_PublishDataTrackResponse{
			PublishDataTrackResponse: publishDataTrackResponse,
		},
	}
}

func (s *signalling) SignalUnpublishDataTrackResponse(unpublishDataTrackResponse *samvaad.UnpublishDataTrackResponse) proto.Message {
	return &samvaad.SignalResponse{
		Message: &samvaad.SignalResponse_UnpublishDataTrackResponse{
			UnpublishDataTrackResponse: unpublishDataTrackResponse,
		},
	}
}

func (s *signalling) SignalDataTrackSubscriberHandles(dataTrackSubscriberHandles *samvaad.DataTrackSubscriberHandles) proto.Message {
	return &samvaad.SignalResponse{
		Message: &samvaad.SignalResponse_DataTrackSubscriberHandles{
			DataTrackSubscriberHandles: dataTrackSubscriberHandles,
		},
	}
}


