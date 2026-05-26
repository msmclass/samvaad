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

package dynacast

import (
	"maps"
	"time"

	"github.com/bep/debounce"

	"github.com/msmclass/samvaad/pkg/samvaad/codecs/mime"
	"github.com/msmclass/samvaad/pkg/proto/samvaad"
	"github.com/msmclass/samvaad/pkg/samvaad/logger"

	"github.com/msmclass/samvaad/pkg/rtc/types"
)

var _ DynacastManager = (*dynacastManagerVideo)(nil)
var _ dynacastQualityListener = (*dynacastManagerVideo)(nil)

type DynacastManagerVideoParams struct {
	DynacastPauseDelay time.Duration
	Listener           DynacastManagerListener
	Logger             logger.Logger
}

type dynacastManagerVideo struct {
	params DynacastManagerVideoParams

	maxSubscribedQuality          map[mime.MimeType]samvaad.VideoQuality
	committedMaxSubscribedQuality map[mime.MimeType]samvaad.VideoQuality

	maxSubscribedQualityDebounce        func(func())
	maxSubscribedQualityDebouncePending bool

	isClosed bool

	*dynacastManagerBase
}

func NewDynacastManagerVideo(params DynacastManagerVideoParams) DynacastManager {
	if params.Logger == nil {
		params.Logger = logger.GetLogger()
	}
	d := &dynacastManagerVideo{
		params:                        params,
		maxSubscribedQuality:          make(map[mime.MimeType]samvaad.VideoQuality),
		committedMaxSubscribedQuality: make(map[mime.MimeType]samvaad.VideoQuality),
	}
	if params.DynacastPauseDelay > 0 {
		d.maxSubscribedQualityDebounce = debounce.New(params.DynacastPauseDelay)
	}
	d.dynacastManagerBase = newDynacastManagerBase(dynacastManagerBaseParams{
		Logger:        params.Logger,
		OpsQueueDepth: 64,
		OnRestart: func() {
			d.committedMaxSubscribedQuality = make(map[mime.MimeType]samvaad.VideoQuality)
		},
		OnDynacastQualityCreate: func(mimeType mime.MimeType) dynacastQuality {
			dq := newDynacastQualityVideo(dynacastQualityVideoParams{
				MimeType: mimeType,
				Listener: d,
				Logger:   d.params.Logger,
			})
			return dq
		},
		OnRegressCodec: func(fromMime, toMime mime.MimeType) {
			d.maxSubscribedQuality[fromMime] = samvaad.VideoQuality_OFF

			// if the new codec is not added, notify the publisher to start publishing
			if _, ok := d.maxSubscribedQuality[toMime]; !ok {
				d.maxSubscribedQuality[toMime] = samvaad.VideoQuality_HIGH
			}
		},
		OnUpdateNeeded: d.update,
	})
	return d
}

// It is possible for tracks to be in pending close state. When track
// is waiting to be closed, a node is not streaming a track. This can
// be used to force an update announcing that subscribed quality is OFF,
// i.e. indicating not pulling track any more.
func (d *dynacastManagerVideo) ForceQuality(quality samvaad.VideoQuality) {
	d.lock.Lock()
	defer d.lock.Unlock()

	for mime := range d.committedMaxSubscribedQuality {
		d.committedMaxSubscribedQuality[mime] = quality
	}

	d.enqueueSubscribedQualityChange()
}

func (d *dynacastManagerVideo) NotifySubscriberMaxQuality(
	subscriberID samvaad.ParticipantID,
	mime mime.MimeType,
	quality samvaad.VideoQuality,
) {
	dq := d.getOrCreateDynacastQuality(mime)
	if dq != nil {
		dq.NotifySubscriberMaxQuality(subscriberID, quality)
	}
}

func (d *dynacastManagerVideo) NotifySubscriberNodeMaxQuality(
	nodeID samvaad.NodeID,
	qualities []types.SubscribedCodecQuality,
) {
	for _, quality := range qualities {
		dq := d.getOrCreateDynacastQuality(quality.CodecMime)
		if dq != nil {
			dq.NotifySubscriberNodeMaxQuality(nodeID, quality.Quality)
		}
	}
}

func (d *dynacastManagerVideo) OnUpdateMaxQualityForMime(
	mime mime.MimeType,
	maxQuality samvaad.VideoQuality,
) {
	d.lock.Lock()
	if _, ok := d.regressedCodec[mime]; !ok {
		d.maxSubscribedQuality[mime] = maxQuality
	}
	d.lock.Unlock()

	d.update(false)
}

func (d *dynacastManagerVideo) update(force bool) {
	d.lock.Lock()

	d.params.Logger.Debugw(
		"processing quality change",
		"force", force,
		"committedMaxSubscribedQuality", d.committedMaxSubscribedQuality,
		"maxSubscribedQuality", d.maxSubscribedQuality,
	)

	if len(d.maxSubscribedQuality) == 0 {
		// no mime has been added, nothing to update
		d.lock.Unlock()
		return
	}

	// add or remove of a mime triggers an update
	changed := len(d.maxSubscribedQuality) != len(d.committedMaxSubscribedQuality)
	downgradesOnly := !changed
	if !changed {
		for mime, quality := range d.maxSubscribedQuality {
			if cq, ok := d.committedMaxSubscribedQuality[mime]; ok {
				if cq != quality {
					changed = true
				}

				if (cq == samvaad.VideoQuality_OFF && quality != samvaad.VideoQuality_OFF) || (cq != samvaad.VideoQuality_OFF && quality != samvaad.VideoQuality_OFF && cq < quality) {
					downgradesOnly = false
				}
			}
		}
	}

	if !force {
		if !changed {
			d.lock.Unlock()
			return
		}

		if downgradesOnly && d.maxSubscribedQualityDebounce != nil {
			if !d.maxSubscribedQualityDebouncePending {
				d.params.Logger.Debugw(
					"debouncing quality downgrade",
					"committedMaxSubscribedQuality", d.committedMaxSubscribedQuality,
					"maxSubscribedQuality", d.maxSubscribedQuality,
				)
				d.maxSubscribedQualityDebounce(func() {
					d.update(true)
				})
				d.maxSubscribedQualityDebouncePending = true
			} else {
				d.params.Logger.Debugw(
					"quality downgrade waiting for debounce",
					"committedMaxSubscribedQuality", d.committedMaxSubscribedQuality,
					"maxSubscribedQuality", d.maxSubscribedQuality,
				)
			}
			d.lock.Unlock()
			return
		}
	}

	// clear debounce on send
	if d.maxSubscribedQualityDebounce != nil {
		d.maxSubscribedQualityDebounce(func() {})
		d.maxSubscribedQualityDebouncePending = false
	}

	d.params.Logger.Debugw(
		"committing quality change",
		"force", force,
		"committedMaxSubscribedQuality", d.committedMaxSubscribedQuality,
		"maxSubscribedQuality", d.maxSubscribedQuality,
	)

	// commit change
	d.committedMaxSubscribedQuality = make(map[mime.MimeType]samvaad.VideoQuality, len(d.maxSubscribedQuality))
	maps.Copy(d.committedMaxSubscribedQuality, d.maxSubscribedQuality)

	d.enqueueSubscribedQualityChange()
	d.lock.Unlock()
}

func (d *dynacastManagerVideo) enqueueSubscribedQualityChange() {
	if d.isClosed || d.params.Listener == nil {
		return
	}

	subscribedCodecs := make([]*samvaad.SubscribedCodec, 0, len(d.committedMaxSubscribedQuality))
	maxSubscribedQualities := make([]types.SubscribedCodecQuality, 0, len(d.committedMaxSubscribedQuality))
	for mime, quality := range d.committedMaxSubscribedQuality {
		maxSubscribedQualities = append(maxSubscribedQualities, types.SubscribedCodecQuality{
			CodecMime: mime,
			Quality:   quality,
		})

		if quality == samvaad.VideoQuality_OFF {
			subscribedCodecs = append(subscribedCodecs, &samvaad.SubscribedCodec{
				Codec: mime.String(),
				Qualities: []*samvaad.SubscribedQuality{
					{Quality: samvaad.VideoQuality_LOW, Enabled: false},
					{Quality: samvaad.VideoQuality_MEDIUM, Enabled: false},
					{Quality: samvaad.VideoQuality_HIGH, Enabled: false},
				},
			})
		} else {
			var subscribedQualities []*samvaad.SubscribedQuality
			for q := samvaad.VideoQuality_LOW; q <= samvaad.VideoQuality_HIGH; q++ {
				subscribedQualities = append(subscribedQualities, &samvaad.SubscribedQuality{
					Quality: q,
					Enabled: q <= quality,
				})
			}
			subscribedCodecs = append(subscribedCodecs, &samvaad.SubscribedCodec{
				Codec:     mime.String(),
				Qualities: subscribedQualities,
			})
		}
	}

	d.params.Logger.Debugw(
		"subscribedMaxQualityChange",
		"subscribedCodecs", subscribedCodecs,
		"maxSubscribedQualities", maxSubscribedQualities,
	)
	d.notifyOpsQueue.Enqueue(func() {
		d.params.Listener.OnDynacastSubscribedMaxQualityChange(subscribedCodecs, maxSubscribedQualities)
	})
}


