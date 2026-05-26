package agent

import (
	"time"

	"github.com/msmclass/samvaad/pkg/samvaad/auth"
	"github.com/msmclass/samvaad/pkg/proto/samvaad"
)

func BuildAgentToken(
	apiKey, secret, roomName, participantIdentity, participantName, participantMetadata string,
	participantAttributes map[string]string,
	permissions *samvaad.ParticipantPermission,
) (string, error) {
	grant := &auth.VideoGrant{
		RoomJoin:             true,
		Agent:                true,
		Room:                 roomName,
		CanSubscribe:         &permissions.CanSubscribe,
		CanPublish:           &permissions.CanPublish,
		CanPublishData:       &permissions.CanPublishData,
		Hidden:               permissions.Hidden,
		CanUpdateOwnMetadata: &permissions.CanUpdateMetadata,
		CanSubscribeMetrics:  &permissions.CanSubscribeMetrics,
	}

	at := auth.NewAccessToken(apiKey, secret).
		SetVideoGrant(grant).
		SetIdentity(participantIdentity).
		SetName(participantName).
		SetKind(samvaad.ParticipantInfo_AGENT).
		SetValidFor(1 * time.Hour).
		SetMetadata(participantMetadata).
		SetAttributes(participantAttributes)

	return at.ToJWT()
}
