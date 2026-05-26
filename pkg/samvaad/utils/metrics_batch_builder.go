// Copyright 2023 Samvaad, Inc.
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

package utils

import (
	"errors"
	"time"

	"github.com/msmclass/samvaad/pkg/proto/samvaad"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	MetricsBatchBuilderInvalidTimeSeriesMetricId = -1
)

var (
	ErrInvalidMetricLabel           = errors.New("invalid metric label")
	ErrFilteredMetricLabel          = errors.New("filtered metric label")
	ErrInvalidTimeSeriesMetricIndex = errors.New("invalid time series metric index")
)

type MetricsBatchBuilder struct {
	*samvaad.MetricsBatch

	stringData       map[string]uint32
	restrictedLabels MetricRestrictedLabels
}

func NewMetricsBatchBuilder() *MetricsBatchBuilder {
	return &MetricsBatchBuilder{
		MetricsBatch: &samvaad.MetricsBatch{},
		stringData:   make(map[string]uint32),
	}
}

func (m *MetricsBatchBuilder) ToProto() *samvaad.MetricsBatch {
	return m.MetricsBatch
}

func (m *MetricsBatchBuilder) SetTime(at time.Time, normalizedAt time.Time) {
	m.MetricsBatch.TimestampMs = at.UnixMilli()
	m.MetricsBatch.NormalizedTimestamp = timestamppb.New(normalizedAt)
}

type MetricLabelRange struct {
	StartInclusive samvaad.MetricLabel
	EndInclusive   samvaad.MetricLabel
}

type MetricRestrictedLabels struct {
	LabelRanges         []MetricLabelRange
	ParticipantIdentity samvaad.ParticipantIdentity
}

func (m *MetricsBatchBuilder) SetRestrictedLabels(mrl MetricRestrictedLabels) {
	m.restrictedLabels = mrl
}

type MetricSample struct {
	At           time.Time
	NormalizedAt time.Time
	Value        float32
}

type TimeSeriesMetric struct {
	MetricLabel         samvaad.MetricLabel
	CustomMetricLabel   string
	ParticipantIdentity samvaad.ParticipantIdentity
	TrackID             samvaad.TrackID
	Samples             []MetricSample
	Rid                 string
}

func (m *MetricsBatchBuilder) AddTimeSeriesMetric(tsm TimeSeriesMetric) (int, error) {
	ptsm := &samvaad.TimeSeriesMetric{}

	if tsm.CustomMetricLabel != "" {
		ptsm.Label = m.getStrDataIndex(tsm.CustomMetricLabel)
	} else {
		if tsm.MetricLabel >= samvaad.MetricLabel_METRIC_LABEL_PREDEFINED_MAX_VALUE {
			return MetricsBatchBuilderInvalidTimeSeriesMetricId, ErrInvalidMetricLabel
		}

		if m.isLabelFiltered(tsm.MetricLabel, tsm.ParticipantIdentity) {
			return MetricsBatchBuilderInvalidTimeSeriesMetricId, ErrFilteredMetricLabel
		}

		ptsm.Label = uint32(tsm.MetricLabel)
	}

	if tsm.ParticipantIdentity != "" {
		ptsm.ParticipantIdentity = m.getStrDataIndex(string(tsm.ParticipantIdentity))
	}

	if tsm.TrackID != "" {
		ptsm.TrackSid = m.getStrDataIndex(string(tsm.TrackID))
	}

	for _, sample := range tsm.Samples {
		ptsm.Samples = append(ptsm.Samples, &samvaad.MetricSample{
			TimestampMs:         sample.At.UnixMilli(),
			NormalizedTimestamp: timestamppb.New(sample.NormalizedAt),
			Value:               sample.Value,
		})
	}

	if tsm.Rid != "" {
		ptsm.Rid = m.getStrDataIndex(tsm.Rid)
	}

	m.MetricsBatch.TimeSeries = append(m.MetricsBatch.TimeSeries, ptsm)
	return len(m.MetricsBatch.TimeSeries) - 1, nil
}

func (m *MetricsBatchBuilder) AddMetricSamplesToTimeSeriesMetric(timeSeriesMetricIdx int, samples []MetricSample) error {
	if timeSeriesMetricIdx < 0 || timeSeriesMetricIdx >= len(m.MetricsBatch.TimeSeries) {
		return ErrInvalidTimeSeriesMetricIndex
	}

	ptsm := m.MetricsBatch.TimeSeries[timeSeriesMetricIdx]
	for _, sample := range samples {
		ptsm.Samples = append(ptsm.Samples, &samvaad.MetricSample{
			TimestampMs:         sample.At.UnixMilli(),
			NormalizedTimestamp: timestamppb.New(sample.NormalizedAt),
			Value:               sample.Value,
		})
	}

	return nil
}

type EventMetric struct {
	MetricLabel         samvaad.MetricLabel
	CustomMetricLabel   string
	ParticipantIdentity samvaad.ParticipantIdentity
	TrackID             samvaad.TrackID
	StartedAt           time.Time
	EndedAt             time.Time
	NormalizedStartedAt time.Time
	NormalizedEndedAt   time.Time
	Metadata            string
	Rid                 string
}

func (m *MetricsBatchBuilder) AddEventMetric(em EventMetric) error {
	pem := &samvaad.EventMetric{}

	if em.CustomMetricLabel != "" {
		pem.Label = m.getStrDataIndex(em.CustomMetricLabel)
	} else {
		if em.MetricLabel >= samvaad.MetricLabel_METRIC_LABEL_PREDEFINED_MAX_VALUE {
			return ErrInvalidMetricLabel
		}

		if m.isLabelFiltered(em.MetricLabel, em.ParticipantIdentity) {
			return ErrFilteredMetricLabel
		}

		pem.Label = uint32(em.MetricLabel)
	}

	if em.ParticipantIdentity != "" {
		pem.ParticipantIdentity = m.getStrDataIndex(string(em.ParticipantIdentity))
	}

	if em.TrackID != "" {
		pem.TrackSid = m.getStrDataIndex(string(em.TrackID))
	}

	pem.StartTimestampMs = em.StartedAt.UnixMilli()
	if !em.EndedAt.IsZero() {
		endTimestampMs := em.EndedAt.UnixMilli()
		pem.EndTimestampMs = &endTimestampMs
	}

	pem.NormalizedStartTimestamp = timestamppb.New(em.NormalizedStartedAt)
	if !em.NormalizedEndedAt.IsZero() {
		pem.NormalizedEndTimestamp = timestamppb.New(em.NormalizedEndedAt)
	}

	pem.Metadata = em.Metadata

	if em.Rid != "" {
		pem.Rid = m.getStrDataIndex(em.Rid)
	}

	m.MetricsBatch.Events = append(m.MetricsBatch.Events, pem)
	return nil
}

func (m *MetricsBatchBuilder) Merge(other *samvaad.MetricsBatch) {
	// Timestamp and NormalizedTimestamp are not merged

	for _, optsm := range other.TimeSeries {
		ptsm := &samvaad.TimeSeriesMetric{
			Samples: optsm.Samples,
		}
		if optsm.Label < uint32(int(samvaad.MetricLabel_METRIC_LABEL_PREDEFINED_MAX_VALUE)) {
			participantIdentity, ok := getStrDataForIndex(other, optsm.ParticipantIdentity)
			if ok && m.isLabelFiltered(samvaad.MetricLabel(optsm.Label), samvaad.ParticipantIdentity(participantIdentity)) {
				continue
			}

			ptsm.Label = optsm.Label
		} else {
			if tidx, ok := m.translateStrDataIndex(other.StrData, optsm.Label); ok {
				ptsm.Label = tidx
			}
		}

		if tidx, ok := m.translateStrDataIndex(other.StrData, optsm.ParticipantIdentity); ok {
			ptsm.ParticipantIdentity = tidx
		}

		if tidx, ok := m.translateStrDataIndex(other.StrData, optsm.TrackSid); ok {
			ptsm.TrackSid = tidx
		}

		if tidx, ok := m.translateStrDataIndex(other.StrData, optsm.Rid); ok {
			ptsm.Rid = tidx
		}

		m.MetricsBatch.TimeSeries = append(m.MetricsBatch.TimeSeries, ptsm)
	}

	for _, opem := range other.Events {
		pem := &samvaad.EventMetric{}
		if opem.Label < uint32(int(samvaad.MetricLabel_METRIC_LABEL_PREDEFINED_MAX_VALUE)) {
			participantIdentity, ok := getStrDataForIndex(other, opem.ParticipantIdentity)
			if ok && m.isLabelFiltered(samvaad.MetricLabel(opem.Label), samvaad.ParticipantIdentity(participantIdentity)) {
				continue
			}

			pem.Label = opem.Label
		} else {
			if tidx, ok := m.translateStrDataIndex(other.StrData, opem.Label); ok {
				pem.Label = tidx
			}
		}

		if tidx, ok := m.translateStrDataIndex(other.StrData, opem.ParticipantIdentity); ok {
			pem.ParticipantIdentity = tidx
		}

		if tidx, ok := m.translateStrDataIndex(other.StrData, opem.TrackSid); ok {
			pem.TrackSid = tidx
		}

		pem.StartTimestampMs = opem.StartTimestampMs
		pem.EndTimestampMs = opem.EndTimestampMs
		pem.NormalizedStartTimestamp = opem.NormalizedStartTimestamp
		pem.NormalizedEndTimestamp = opem.NormalizedEndTimestamp

		pem.Metadata = opem.Metadata

		if tidx, ok := m.translateStrDataIndex(other.StrData, opem.Rid); ok {
			pem.Rid = tidx
		}

		m.MetricsBatch.Events = append(m.MetricsBatch.Events, pem)
	}
}

func (m *MetricsBatchBuilder) IsEmpty() bool {
	return len(m.MetricsBatch.TimeSeries) == 0 && len(m.MetricsBatch.Events) == 0
}

func (m *MetricsBatchBuilder) isLabelFiltered(label samvaad.MetricLabel, participantIdentity samvaad.ParticipantIdentity) bool {
	if participantIdentity == m.restrictedLabels.ParticipantIdentity {
		// all labels allowed for restricted participant
		return false
	}

	for _, mlr := range m.restrictedLabels.LabelRanges {
		if label >= mlr.StartInclusive && label <= mlr.EndInclusive {
			return true
		}
	}

	return false
}

func (m *MetricsBatchBuilder) getStrDataIndex(s string) uint32 {
	idx, ok := m.stringData[s]
	if !ok {
		m.MetricsBatch.StrData = append(m.MetricsBatch.StrData, s)
		idx = uint32(int(samvaad.MetricLabel_METRIC_LABEL_PREDEFINED_MAX_VALUE) + len(m.MetricsBatch.StrData) - 1)
		m.stringData[s] = idx
	}
	return idx
}

func (m *MetricsBatchBuilder) translateStrDataIndex(strData []string, index uint32) (uint32, bool) {
	if index < uint32(samvaad.MetricLabel_METRIC_LABEL_PREDEFINED_MAX_VALUE) {
		return 0, false
	}

	baseIdx := index - uint32(samvaad.MetricLabel_METRIC_LABEL_PREDEFINED_MAX_VALUE)
	if len(strData) <= int(baseIdx) {
		return 0, false
	}

	// add if necessary
	return m.getStrDataIndex(strData[baseIdx]), true
}

// -----------------------------------------------------

func getStrDataForIndex(mb *samvaad.MetricsBatch, index uint32) (string, bool) {
	if index < uint32(samvaad.MetricLabel_METRIC_LABEL_PREDEFINED_MAX_VALUE) {
		return "", false
	}

	baseIdx := index - uint32(samvaad.MetricLabel_METRIC_LABEL_PREDEFINED_MAX_VALUE)
	if len(mb.StrData) <= int(baseIdx) {
		return "", false
	}

	return mb.StrData[baseIdx], true
}
