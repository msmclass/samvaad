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
	"time"

	"github.com/pion/webrtc/v4"

	"github.com/msmclass/samvaad/pkg/proto/samvaad"
	"github.com/msmclass/samvaad/pkg/samvaad/logger"
	protosignalling "github.com/msmclass/samvaad/pkg/samvaad/signalling"

	"github.com/msmclass/samvaad/pkg/routing"
	"github.com/msmclass/samvaad/pkg/rtc/types"
)

func (p *ParticipantImpl) SwapResponseSink(sink routing.MessageSink, reason types.SignallingCloseReason) {
	p.signaller.SwapResponseSink(sink, reason)
}

func (p *ParticipantImpl) GetResponseSink() routing.MessageSink {
	return p.signaller.GetResponseSink()
}

func (p *ParticipantImpl) CloseSignalConnection(reason types.SignallingCloseReason) {
	p.signaller.CloseSignalConnection(reason)
}

func (p *ParticipantImpl) SendJoinResponse(joinResponse *samvaad.JoinResponse) error {
	// keep track of participant updates and versions
	p.updateLock.Lock()
	for _, op := range joinResponse.OtherParticipants {
		p.updateCache.Add(samvaad.ParticipantID(op.Sid), participantUpdateInfo{
			identity:  samvaad.ParticipantIdentity(op.Identity),
			version:   op.Version,
			state:     op.State,
			updatedAt: time.Now(),
		})
	}
	p.updateLock.Unlock()

	// send Join response
	err := p.signaller.WriteMessage(p.signalling.SignalJoinResponse(joinResponse))
	if err != nil {
		return err
	}

	// update state after sending message, so that no participant updates could slip through before JoinResponse is sent
	p.updateLock.Lock()
	if p.State() == samvaad.ParticipantInfo_JOINING {
		p.updateState(samvaad.ParticipantInfo_JOINED)
	}
	queuedUpdates := p.queuedUpdates
	p.queuedUpdates = nil
	p.updateLock.Unlock()

	if len(queuedUpdates) > 0 {
		return p.SendParticipantUpdate(queuedUpdates)
	}

	return nil
}

func (p *ParticipantImpl) SendParticipantUpdate(participantsToUpdate []*samvaad.ParticipantInfo) error {
	p.updateLock.Lock()
	if p.IsDisconnected() {
		p.updateLock.Unlock()
		return nil
	}

	if !p.IsReady() {
		// queue up updates
		p.queuedUpdates = append(p.queuedUpdates, participantsToUpdate...)
		p.updateLock.Unlock()
		return nil
	}
	validUpdates := make([]*samvaad.ParticipantInfo, 0, len(participantsToUpdate))
	for _, pi := range participantsToUpdate {
		isValid := true
		pID := samvaad.ParticipantID(pi.Sid)
		if lastVersion, ok := p.updateCache.Get(pID); ok {
			// this is a message delivered out of order, a more recent version of the message had already been
			// sent.
			if pi.Version < lastVersion.version {
				p.params.Logger.Debugw(
					"skipping outdated participant update",
					"otherParticipant", pi.Identity,
					"otherPID", pi.Sid,
					"version", pi.Version,
					"lastVersion", lastVersion,
				)
				isValid = false
			}
		}
		if pi.Permission != nil && pi.Permission.Hidden && pi.Sid != string(p.ID()) {
			p.params.Logger.Debugw("skipping hidden participant update", "otherParticipant", pi.Identity)
			isValid = false
		}
		if isValid {
			p.updateCache.Add(pID, participantUpdateInfo{
				identity:  samvaad.ParticipantIdentity(pi.Identity),
				version:   pi.Version,
				state:     pi.State,
				updatedAt: time.Now(),
			})
			validUpdates = append(validUpdates, pi)
		}
	}
	p.updateLock.Unlock()

	return p.signaller.WriteMessage(p.signalling.SignalParticipantUpdate(validUpdates))
}

// SendSpeakerUpdate notifies participant changes to speakers. only send members that have changed since last update
func (p *ParticipantImpl) SendSpeakerUpdate(speakers []*samvaad.SpeakerInfo, force bool) error {
	if !p.IsReady() {
		return nil
	}

	var scopedSpeakers []*samvaad.SpeakerInfo
	if force {
		scopedSpeakers = speakers
	} else {
		for _, s := range speakers {
			participantID := samvaad.ParticipantID(s.Sid)
			if p.IsSubscribedTo(participantID) || participantID == p.ID() {
				scopedSpeakers = append(scopedSpeakers, s)
			}
		}
	}

	return p.signaller.WriteMessage(p.signalling.SignalSpeakerUpdate(scopedSpeakers))
}

func (p *ParticipantImpl) SendRoomUpdate(room *samvaad.Room) error {
	return p.signaller.WriteMessage(p.signalling.SignalRoomUpdate(room))
}

func (p *ParticipantImpl) SendConnectionQualityUpdate(update *samvaad.ConnectionQualityUpdate) error {
	return p.signaller.WriteMessage(p.signalling.SignalConnectionQualityUpdate(update))
}

func (p *ParticipantImpl) SendRefreshToken(token string) error {
	return p.signaller.WriteMessage(p.signalling.SignalRefreshToken(token))
}

func (p *ParticipantImpl) sendRequestResponse(requestResponse *samvaad.RequestResponse) error {
	if !p.params.ClientInfo.SupportsRequestResponse() {
		return nil
	}

	if requestResponse.Reason == samvaad.RequestResponse_OK && !p.ProtocolVersion().SupportsNonErrorSignalResponse() {
		return nil
	}

	return p.signaller.WriteMessage(p.signalling.SignalRequestResponse(requestResponse))
}

func (p *ParticipantImpl) SendRoomMovedResponse(roomMovedResponse *samvaad.RoomMovedResponse) error {
	return p.signaller.WriteMessage(p.signalling.SignalRoomMovedResponse(roomMovedResponse))
}

func (p *ParticipantImpl) HandleReconnectAndSendResponse(reconnectReason samvaad.ReconnectReason, reconnectResponse *samvaad.ReconnectResponse) error {
	p.TransportManager.HandleClientReconnect(reconnectReason)

	if !p.params.ClientInfo.CanHandleReconnectResponse() {
		return nil
	}
	if err := p.signaller.WriteMessage(p.signalling.SignalReconnectResponse(reconnectResponse)); err != nil {
		return err
	}

	if p.params.ProtocolVersion.SupportsDisconnectedUpdate() {
		return p.sendDisconnectUpdatesForReconnect()
	}

	return nil
}

func (p *ParticipantImpl) sendDisconnectUpdatesForReconnect() error {
	lastSignalAt := p.TransportManager.LastSeenSignalAt()
	var disconnectedParticipants []*samvaad.ParticipantInfo
	p.updateLock.Lock()
	keys := p.updateCache.Keys()
	for i := len(keys) - 1; i >= 0; i-- {
		if info, ok := p.updateCache.Get(keys[i]); ok {
			if info.updatedAt.Before(lastSignalAt) {
				break
			} else if info.state == samvaad.ParticipantInfo_DISCONNECTED {
				disconnectedParticipants = append(disconnectedParticipants, &samvaad.ParticipantInfo{
					Sid:      string(keys[i]),
					Identity: string(info.identity),
					Version:  info.version,
					State:    samvaad.ParticipantInfo_DISCONNECTED,
				})
			}
		}
	}
	p.updateLock.Unlock()

	return p.signaller.WriteMessage(p.signalling.SignalParticipantUpdate(disconnectedParticipants))
}

func (p *ParticipantImpl) sendICECandidate(ic *webrtc.ICECandidate, target samvaad.SignalTarget) error {
	prevIC := p.icQueue[target].Swap(ic)
	if prevIC == nil {
		return nil
	}

	trickle := protosignalling.ToProtoTrickle(prevIC.ToJSON(), target, ic == nil)
	p.params.Logger.Debugw("sending ICE candidate", "transport", target, "trickle", logger.Proto(trickle))

	return p.signaller.WriteMessage(p.signalling.SignalICECandidate(trickle))
}

func (p *ParticipantImpl) sendTrackMuted(trackID samvaad.TrackID, muted bool) {
	_ = p.signaller.WriteMessage(p.signalling.SignalTrackMuted(&samvaad.MuteTrackRequest{
		Sid:   string(trackID),
		Muted: muted,
	}))
}

func (p *ParticipantImpl) sendTrackPublished(cid string, ti *samvaad.TrackInfo) error {
	p.pubLogger.Debugw("sending track published", "cid", cid, "trackInfo", logger.Proto(ti))
	return p.signaller.WriteMessage(p.signalling.SignalTrackPublished(&samvaad.TrackPublishedResponse{
		Cid:   cid,
		Track: ti,
	}))
}

func (p *ParticipantImpl) sendTrackUnpublished(trackID samvaad.TrackID) {
	_ = p.signaller.WriteMessage(p.signalling.SignalTrackUnpublished(&samvaad.TrackUnpublishedResponse{
		TrackSid: string(trackID),
	}))
}

func (p *ParticipantImpl) sendTrackHasBeenSubscribed(trackID samvaad.TrackID) {
	if !p.params.ClientInfo.SupportsTrackSubscribedEvent() {
		return
	}
	_ = p.signaller.WriteMessage(p.signalling.SignalTrackSubscribed(&samvaad.TrackSubscribed{
		TrackSid: string(trackID),
	}))
	p.params.Logger.Debugw("track has been subscribed", "trackID", trackID)
}

func (p *ParticipantImpl) sendLeaveRequest(
	reason types.ParticipantCloseReason,
	isExpectedToResume bool,
	isExpectedToReconnect bool,
	sendOnlyIfSupportingLeaveRequestWithAction bool,
) error {
	var leave *samvaad.LeaveRequest
	if p.ProtocolVersion().SupportsRegionsInLeaveRequest() {
		leave = &samvaad.LeaveRequest{
			Reason: reason.ToDisconnectReason(),
		}
		switch {
		case isExpectedToResume:
			leave.Action = samvaad.LeaveRequest_RESUME
		case isExpectedToReconnect:
			leave.Action = samvaad.LeaveRequest_RECONNECT
		default:
			leave.Action = samvaad.LeaveRequest_DISCONNECT
		}
		if leave.Action != samvaad.LeaveRequest_DISCONNECT {
			// sending region settings even for RESUME just in case client wants to a full reconnect despite server saying RESUME
			leave.Regions = p.helper().GetRegionSettings(p.params.ClientInfo.Address)
		}
	} else {
		if !sendOnlyIfSupportingLeaveRequestWithAction {
			leave = &samvaad.LeaveRequest{
				CanReconnect: isExpectedToReconnect,
				Reason:       reason.ToDisconnectReason(),
			}
		}
	}
	if leave != nil {
		return p.signaller.WriteMessage(p.signalling.SignalLeaveRequest(leave))
	}

	return nil
}

func (p *ParticipantImpl) sendSdpAnswer(answer webrtc.SessionDescription, answerId uint32, midToTrackID map[string]string) error {
	return p.signaller.WriteMessage(p.signalling.SignalSdpAnswer(protosignalling.ToProtoSessionDescription(answer, answerId, midToTrackID)))
}

func (p *ParticipantImpl) sendSdpOffer(offer webrtc.SessionDescription, offerId uint32, midToTrackID map[string]string) error {
	return p.signaller.WriteMessage(p.signalling.SignalSdpOffer(protosignalling.ToProtoSessionDescription(offer, offerId, midToTrackID)))
}

func (p *ParticipantImpl) sendStreamStateUpdate(streamStateUpdate *samvaad.StreamStateUpdate) error {
	return p.signaller.WriteMessage(p.signalling.SignalStreamStateUpdate(streamStateUpdate))
}

func (p *ParticipantImpl) sendSubscribedQualityUpdate(subscribedQualityUpdate *samvaad.SubscribedQualityUpdate) error {
	return p.signaller.WriteMessage(p.signalling.SignalSubscribedQualityUpdate(subscribedQualityUpdate))
}

func (p *ParticipantImpl) sendSubscribedAudioCodecUpdate(subscribedAudioCodecUpdate *samvaad.SubscribedAudioCodecUpdate) error {
	return p.signaller.WriteMessage(p.signalling.SignalSubscribedAudioCodecUpdate(subscribedAudioCodecUpdate))
}

func (p *ParticipantImpl) sendSubscriptionResponse(trackID samvaad.TrackID, subErr samvaad.SubscriptionError) error {
	return p.signaller.WriteMessage(p.signalling.SignalSubscriptionResponse(&samvaad.SubscriptionResponse{
		TrackSid: string(trackID),
		Err:      subErr,
	}))
}

func (p *ParticipantImpl) SendSubscriptionPermissionUpdate(publisherID samvaad.ParticipantID, trackID samvaad.TrackID, allowed bool) error {
	p.subLogger.Debugw("sending subscription permission update", "publisherID", publisherID, "trackID", trackID, "allowed", allowed)
	err := p.signaller.WriteMessage(p.signalling.SignalSubscriptionPermissionUpdate(&samvaad.SubscriptionPermissionUpdate{
		ParticipantSid: string(publisherID),
		TrackSid:       string(trackID),
		Allowed:        allowed,
	}))
	if err != nil {
		p.subLogger.Errorw("could not send subscription permission update", err)
	}
	return err
}

func (p *ParticipantImpl) sendMediaSectionsRequirement(numAudios uint32, numVideos uint32) error {
	p.pubLogger.Debugw(
		"sending media sections requirement",
		"numAudios", numAudios,
		"numVideos", numVideos,
	)
	err := p.signaller.WriteMessage(p.signalling.SignalMediaSectionsRequirement(&samvaad.MediaSectionsRequirement{
		NumAudios: numAudios,
		NumVideos: numVideos,
	}))
	if err != nil {
		p.subLogger.Errorw("could not send media sections requirement", err)
	}
	return err
}

func (p *ParticipantImpl) sendPublishDataTrackResponse(dti *samvaad.DataTrackInfo) error {
	return p.signaller.WriteMessage(p.signalling.SignalPublishDataTrackResponse(&samvaad.PublishDataTrackResponse{
		Info: dti,
	}))
}

func (p *ParticipantImpl) sendUnpublishDataTrackResponse(dti *samvaad.DataTrackInfo) error {
	return p.signaller.WriteMessage(p.signalling.SignalUnpublishDataTrackResponse(&samvaad.UnpublishDataTrackResponse{
		Info: dti,
	}))
}

func (p *ParticipantImpl) SendDataTrackSubscriberHandles(handles map[uint32]*samvaad.DataTrackSubscriberHandles_PublishedDataTrack) error {
	return p.signaller.WriteMessage(p.signalling.SignalDataTrackSubscriberHandles(&samvaad.DataTrackSubscriberHandles{
		SubHandles: handles,
	}))
}


