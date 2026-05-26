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
	"errors"
	"sync"

	"github.com/frostbyte73/core"
	"github.com/msmclass/samvaad/pkg/rtc/datatrack"
	"github.com/msmclass/samvaad/pkg/rtc/types"
	sfuutils "github.com/msmclass/samvaad/pkg/sfu/utils"
	"github.com/msmclass/samvaad/pkg/proto/samvaad"
	"github.com/msmclass/samvaad/pkg/samvaad/logger"
	"github.com/msmclass/samvaad/pkg/samvaad/utils"
)

var (
	errReceiverClosed = errors.New("datatrack is closed")
)

var _ types.DataTrack = (*DataTrack)(nil)

type DataTrackParams struct {
	Logger              logger.Logger
	ParticipantID       func() samvaad.ParticipantID
	ParticipantIdentity samvaad.ParticipantIdentity
	BytesTrackStats     *BytesTrackStats
}

type DataTrack struct {
	params DataTrackParams

	logger logger.Logger

	lock             sync.Mutex
	dti              *samvaad.DataTrackInfo
	subscribedTracks map[samvaad.ParticipantID]types.DataDownTrack

	downTrackSpreader *sfuutils.DownTrackSpreader[types.DataTrackSender]

	stats *dataTrackStats

	closed core.Fuse
}

func NewDataTrack(params DataTrackParams, dti *samvaad.DataTrackInfo) *DataTrack {
	d := &DataTrack{
		params:           params,
		dti:              dti,
		subscribedTracks: make(map[samvaad.ParticipantID]types.DataDownTrack),
	}
	d.logger = params.Logger.WithValues("name", d.Name(), "handle", dti.PubHandle)
	d.downTrackSpreader = sfuutils.NewDownTrackSpreader[types.DataTrackSender](sfuutils.DownTrackSpreaderParams{
		Threshold: 20,
		Logger:    d.logger,
	})
	d.stats = newDataTrackStats(dataTrackStatsParams{Logger: d.logger})
	d.logger.Infow("created data track", "dataTrackInfo", logger.Proto(d.dti))
	return d
}

func (d *DataTrack) Close() {
	d.logger.Infow("closing data track")
	d.closed.Break()

	d.stats.Close()
	if d.params.BytesTrackStats != nil {
		d.params.BytesTrackStats.Stop()
	}
}

func (d *DataTrack) PublisherID() samvaad.ParticipantID {
	return d.params.ParticipantID()
}

func (d *DataTrack) PublisherIdentity() samvaad.ParticipantIdentity {
	return d.params.ParticipantIdentity
}

func (d *DataTrack) ToProto() *samvaad.DataTrackInfo {
	return utils.CloneProto(d.dti)
}

func (d *DataTrack) PubHandle() uint16 {
	return uint16(d.dti.PubHandle)
}

func (d *DataTrack) ID() samvaad.TrackID {
	return samvaad.TrackID(d.dti.Sid)
}

func (d *DataTrack) Name() string {
	return d.dti.Name
}

func (d *DataTrack) AddSubscriber(sub types.LocalParticipant) (types.DataDownTrack, error) {
	d.lock.Lock()
	defer d.lock.Unlock()

	if _, ok := d.subscribedTracks[sub.ID()]; ok {
		return nil, errAlreadySubscribed
	}

	bytesStats := NewBytesTrackStats(
		sub.GetCountry(),
		d.ID(),
		sub.ID(),
		sub.Kind(),
		sub.KindDetails(),
		sub.GetTelemetryListener(),
		sub.GetReporter(),
	)
	dataDownTrack, err := NewDataDownTrack(DataDownTrackParams{
		Logger:           sub.GetLogger().WithValues("trackID", d.ID()),
		SubscriberID:     sub.ID(),
		PublishDataTrack: d,
		Handle:           sub.GetNextSubscribedDataTrackHandle(),
		Transport:        sub.GetDataTrackTransport(),
		BytesTrackStats:  bytesStats,
	})
	if err != nil {
		bytesStats.Stop()
		return nil, err
	}

	d.subscribedTracks[sub.ID()] = dataDownTrack
	return dataDownTrack, nil
}

func (d *DataTrack) RemoveSubscriber(subID samvaad.ParticipantID) {
	d.lock.Lock()
	dataDownTrack, ok := d.subscribedTracks[subID]
	delete(d.subscribedTracks, subID)
	d.lock.Unlock()

	if ok {
		dataDownTrack.Close()
	}
}

func (d *DataTrack) IsSubscriber(subID samvaad.ParticipantID) bool {
	d.lock.Lock()
	defer d.lock.Unlock()

	_, ok := d.subscribedTracks[subID]
	return ok
}

func (d *DataTrack) AddDataDownTrack(dts types.DataTrackSender) error {
	if d.closed.IsBroken() {
		return errReceiverClosed
	}

	if d.downTrackSpreader.HasDownTrack(dts.SubscriberID()) {
		d.logger.Infow("subscriberID already exists, replacing data downtrack", "subscriberID", dts.SubscriberID())
	}

	d.downTrackSpreader.Store(dts)
	d.logger.Infow("data downtrack added", "subscriberID", dts.SubscriberID())
	return nil
}

func (d *DataTrack) DeleteDataDownTrack(subscriberID samvaad.ParticipantID) {
	d.downTrackSpreader.Free(subscriberID)
	d.logger.Infow("data downtrack deleted", "subscriberID", subscriberID)
}

func (d *DataTrack) HandlePacket(data []byte, packet *datatrack.Packet, arrivalTime int64) {
	d.stats.Update(packet, arrivalTime, len(data))
	if d.params.BytesTrackStats != nil {
		d.params.BytesTrackStats.AddBytes(uint64(len(data)), false)
	}

	d.downTrackSpreader.Broadcast(func(dts types.DataTrackSender) {
		dts.WritePacket(data, packet, arrivalTime)
	})
}


