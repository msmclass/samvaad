package client

import (
	"github.com/frostbyte73/core"
	"github.com/msmclass/samvaad/pkg/rtc/datatrack"
	"github.com/msmclass/samvaad/pkg/proto/samvaad"
	"github.com/msmclass/samvaad/pkg/samvaad/logger"
	"go.uber.org/atomic"
)

type DataTrackRemote struct {
	publisherIdentity  samvaad.ParticipantIdentity
	publisherID        samvaad.ParticipantID
	handle             uint16
	trackID            samvaad.TrackID
	logger             logger.Logger
	numReceivedPackets atomic.Uint32

	closed core.Fuse
}

func NewDataTrackRemote(
	publisherIdentity samvaad.ParticipantIdentity,
	publisherID samvaad.ParticipantID,
	handle uint16,
	trackID samvaad.TrackID,
	logger logger.Logger,
) *DataTrackRemote {
	logger.Infow(
		"creating data track remote",
		"publisherIdentity", publisherIdentity,
		"publisherID", publisherID,
		"handle", handle,
		"trackID", trackID,
	)
	return &DataTrackRemote{
		publisherIdentity: publisherIdentity,
		publisherID:       publisherID,
		handle:            handle,
		trackID:           trackID,
		logger:            logger,
	}
}

func (d *DataTrackRemote) Close() {
	d.logger.Infow(
		"closing data track remote",
		"publisherIdentity", d.publisherIdentity,
		"publisherID", d.publisherID,
		"handle", d.handle,
		"trackID", d.trackID,
	)
	d.closed.Break()
}

func (d *DataTrackRemote) Handle() uint16 {
	return d.handle
}

func (d *DataTrackRemote) ID() samvaad.TrackID {
	return d.trackID
}

func (d *DataTrackRemote) PacketReceived(packet *datatrack.Packet) {
	if d.closed.IsBroken() {
		return
	}

	valid := len(packet.Payload) != 0

	for i := range packet.Payload {
		if packet.Payload[i] != byte(255-i) {
			valid = false
			break
		}
	}
	if valid {
		d.numReceivedPackets.Inc()
	}
}

func (d *DataTrackRemote) NumReceivedPackets() uint32 {
	return d.numReceivedPackets.Load()
}


