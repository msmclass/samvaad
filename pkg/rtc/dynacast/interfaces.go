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
	"github.com/msmclass/samvaad/pkg/samvaad/codecs/mime"
	"github.com/msmclass/samvaad/pkg/proto/samvaad"

	"github.com/msmclass/samvaad/pkg/rtc/types"
)

type DynacastManagerListener interface {
	OnDynacastSubscribedMaxQualityChange(
		subscribedQualities []*samvaad.SubscribedCodec,
		maxSubscribedQualities []types.SubscribedCodecQuality,
	)

	OnDynacastSubscribedAudioCodecChange(codecs []*samvaad.SubscribedAudioCodec)
}

var _ DynacastManagerListener = (*DynacastManagerListenerNull)(nil)

type DynacastManagerListenerNull struct {
}

func (d *DynacastManagerListenerNull) OnDynacastSubscribedMaxQualityChange(
	subscribedQualities []*samvaad.SubscribedCodec,
	maxSubscribedQualities []types.SubscribedCodecQuality,
) {
}
func (d *DynacastManagerListenerNull) OnDynacastSubscribedAudioCodecChange(
	codecs []*samvaad.SubscribedAudioCodec,
) {
}

// -----------------------------------------

type DynacastManager interface {
	AddCodec(mime mime.MimeType)
	HandleCodecRegression(fromMime, toMime mime.MimeType)
	Restart()
	Close()
	ForceUpdate()
	ForceQuality(quality samvaad.VideoQuality)
	ForceEnable(enabled bool)

	NotifySubscriberMaxQuality(
		subscriberID samvaad.ParticipantID,
		mime mime.MimeType,
		quality samvaad.VideoQuality,
	)
	NotifySubscription(
		subscriberID samvaad.ParticipantID,
		mime mime.MimeType,
		enabled bool,
	)

	NotifySubscriberNodeMaxQuality(
		nodeID samvaad.NodeID,
		qualities []types.SubscribedCodecQuality,
	)
	NotifySubscriptionNode(
		nodeID samvaad.NodeID,
		codecs []*samvaad.SubscribedAudioCodec,
	)
	ClearSubscriberNodes()
}

var _ DynacastManager = (*dynacastManagerNull)(nil)

type dynacastManagerNull struct {
}

func (d *dynacastManagerNull) AddCodec(mime mime.MimeType)                          {}
func (d *dynacastManagerNull) HandleCodecRegression(fromMime, toMime mime.MimeType) {}
func (d *dynacastManagerNull) Restart()                                             {}
func (d *dynacastManagerNull) Close()                                               {}
func (d *dynacastManagerNull) ForceUpdate()                                         {}
func (d *dynacastManagerNull) ForceQuality(quality samvaad.VideoQuality)            {}
func (d *dynacastManagerNull) ForceEnable(enabled bool)                             {}
func (d *dynacastManagerNull) NotifySubscriberMaxQuality(
	subscriberID samvaad.ParticipantID,
	mime mime.MimeType,
	quality samvaad.VideoQuality,
) {
}
func (d *dynacastManagerNull) NotifySubscription(
	subscriberID samvaad.ParticipantID,
	mime mime.MimeType,
	enabled bool,
) {
}
func (d *dynacastManagerNull) NotifySubscriberNodeMaxQuality(
	nodeID samvaad.NodeID,
	qualities []types.SubscribedCodecQuality,
) {
}
func (d *dynacastManagerNull) NotifySubscriptionNode(
	nodeID samvaad.NodeID,
	codecs []*samvaad.SubscribedAudioCodec,
) {
}
func (d *dynacastManagerNull) ClearSubscriberNodes() {}

// ------------------------------------------------

type dynacastQualityListener interface {
	OnUpdateMaxQualityForMime(mimeType mime.MimeType, maxQuality samvaad.VideoQuality)
	OnUpdateAudioCodecForMime(mimeType mime.MimeType, enabled bool)
}

var _ dynacastQualityListener = (*dynacastQualityListenerNull)(nil)

type dynacastQualityListenerNull struct {
}

func (d *dynacastQualityListenerNull) OnUpdateMaxQualityForMime(
	mimeType mime.MimeType,
	maxQuality samvaad.VideoQuality,
) {
}

func (d *dynacastQualityListenerNull) OnUpdateAudioCodecForMime(
	mimeType mime.MimeType,
	enabled bool,
) {
}

// ------------------------------------------------

type dynacastQuality interface {
	Start()
	Restart()
	Stop()

	NotifySubscriberMaxQuality(subscriberID samvaad.ParticipantID, quality samvaad.VideoQuality)
	NotifySubscription(subscriberID samvaad.ParticipantID, enabled bool)

	NotifySubscriberNodeMaxQuality(nodeID samvaad.NodeID, quality samvaad.VideoQuality)
	NotifySubscriptionNode(nodeID samvaad.NodeID, enabled bool)
	ClearSubscriberNodes()

	Replace(
		maxSubscriberQuality map[samvaad.ParticipantID]samvaad.VideoQuality,
		maxSubscriberNodeQuality map[samvaad.NodeID]samvaad.VideoQuality,
	)

	Mime() mime.MimeType
	RegressTo(other dynacastQuality)
}

var _ dynacastQuality = (*dynacastQualityNull)(nil)

type dynacastQualityNull struct {
}

func (d *dynacastQualityNull) Start()   {}
func (d *dynacastQualityNull) Restart() {}
func (d *dynacastQualityNull) Stop()    {}
func (d *dynacastQualityNull) NotifySubscriberMaxQuality(subscriberID samvaad.ParticipantID, quality samvaad.VideoQuality) {
}
func (d *dynacastQualityNull) NotifySubscription(subscriberID samvaad.ParticipantID, enabled bool) {}
func (d *dynacastQualityNull) NotifySubscriberNodeMaxQuality(nodeID samvaad.NodeID, quality samvaad.VideoQuality) {
}
func (d *dynacastQualityNull) NotifySubscriptionNode(nodeID samvaad.NodeID, enabled bool) {}
func (d *dynacastQualityNull) ClearSubscriberNodes()                                      {}
func (d *dynacastQualityNull) Replace(
	maxSubscriberQuality map[samvaad.ParticipantID]samvaad.VideoQuality,
	maxSubscriberNodeQuality map[samvaad.NodeID]samvaad.VideoQuality,
) {
}
func (d *dynacastQualityNull) Mime() mime.MimeType             { return mime.MimeTypeUnknown }
func (d *dynacastQualityNull) RegressTo(other dynacastQuality) {}


