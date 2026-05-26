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

package buffer

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/msmclass/samvaad/pkg/samvaad/codecs/mime"
	"github.com/msmclass/samvaad/pkg/proto/samvaad"
)

func TestRidConversion(t *testing.T) {
	type RidAndLayer struct {
		rid   string
		layer int32
	}
	tests := []struct {
		name       string
		trackInfo  *samvaad.TrackInfo
		mimeType   mime.MimeType
		ridToLayer map[string]RidAndLayer
	}{
		{
			"no track info",
			nil,
			mime.MimeTypeVP8,
			map[string]RidAndLayer{
				"":                 {rid: quarterResolutionQ, layer: 0},
				quarterResolutionQ: {rid: quarterResolutionQ, layer: 0},
				halfResolutionH:    {rid: halfResolutionH, layer: 1},
				fullResolutionF:    {rid: fullResolutionF, layer: 2},
			},
		},
		{
			"no layers",
			&samvaad.TrackInfo{},
			mime.MimeTypeVP8,
			map[string]RidAndLayer{
				"":                 {rid: quarterResolutionQ, layer: 0},
				quarterResolutionQ: {rid: quarterResolutionQ, layer: 0},
				halfResolutionH:    {rid: halfResolutionH, layer: 1},
				fullResolutionF:    {rid: fullResolutionF, layer: 2},
			},
		},
		{
			"single layer, low",
			&samvaad.TrackInfo{
				Codecs: []*samvaad.SimulcastCodecInfo{
					{
						MimeType: mime.MimeTypeVP8.String(),
						Layers: []*samvaad.VideoLayer{
							{Quality: samvaad.VideoQuality_LOW},
						},
					},
				},
			},
			mime.MimeTypeVP8,
			map[string]RidAndLayer{
				"":                 {rid: quarterResolutionQ, layer: 0},
				quarterResolutionQ: {rid: quarterResolutionQ, layer: 0},
				halfResolutionH:    {rid: quarterResolutionQ, layer: 0},
				fullResolutionF:    {rid: quarterResolutionQ, layer: 0},
			},
		},
		{
			"single layer, medium",
			&samvaad.TrackInfo{
				Codecs: []*samvaad.SimulcastCodecInfo{
					{
						MimeType: mime.MimeTypeVP8.String(),
						Layers: []*samvaad.VideoLayer{
							{Quality: samvaad.VideoQuality_MEDIUM},
						},
					},
				},
			},
			mime.MimeTypeVP8,
			map[string]RidAndLayer{
				"":                 {rid: quarterResolutionQ, layer: 0},
				quarterResolutionQ: {rid: quarterResolutionQ, layer: 0},
				halfResolutionH:    {rid: quarterResolutionQ, layer: 0},
				fullResolutionF:    {rid: quarterResolutionQ, layer: 0},
			},
		},
		{
			"single layer, high",
			&samvaad.TrackInfo{
				Codecs: []*samvaad.SimulcastCodecInfo{
					{
						MimeType: mime.MimeTypeVP8.String(),
						Layers: []*samvaad.VideoLayer{
							{Quality: samvaad.VideoQuality_MEDIUM},
						},
					},
				},
			},
			mime.MimeTypeVP8,
			map[string]RidAndLayer{
				"":                 {rid: quarterResolutionQ, layer: 0},
				quarterResolutionQ: {rid: quarterResolutionQ, layer: 0},
				halfResolutionH:    {rid: quarterResolutionQ, layer: 0},
				fullResolutionF:    {rid: quarterResolutionQ, layer: 0},
			},
		},
		{
			"two layers, low and medium",
			&samvaad.TrackInfo{
				Codecs: []*samvaad.SimulcastCodecInfo{
					{
						MimeType: mime.MimeTypeVP8.String(),
						Layers: []*samvaad.VideoLayer{
							{Quality: samvaad.VideoQuality_LOW},
							{Quality: samvaad.VideoQuality_MEDIUM},
						},
					},
				},
			},
			mime.MimeTypeVP8,
			map[string]RidAndLayer{
				"":                 {rid: quarterResolutionQ, layer: 0},
				quarterResolutionQ: {rid: quarterResolutionQ, layer: 0},
				halfResolutionH:    {rid: halfResolutionH, layer: 1},
				fullResolutionF:    {rid: halfResolutionH, layer: 1},
			},
		},
		{
			"two layers, low and high",
			&samvaad.TrackInfo{
				Codecs: []*samvaad.SimulcastCodecInfo{
					{
						MimeType: mime.MimeTypeVP8.String(),
						Layers: []*samvaad.VideoLayer{
							{Quality: samvaad.VideoQuality_LOW},
							{Quality: samvaad.VideoQuality_HIGH},
						},
					},
				},
			},
			mime.MimeTypeVP8,
			map[string]RidAndLayer{
				"":                 {rid: quarterResolutionQ, layer: 0},
				quarterResolutionQ: {rid: quarterResolutionQ, layer: 0},
				halfResolutionH:    {rid: halfResolutionH, layer: 1},
				fullResolutionF:    {rid: halfResolutionH, layer: 1},
			},
		},
		{
			"two layers, medium and high",
			&samvaad.TrackInfo{
				Codecs: []*samvaad.SimulcastCodecInfo{
					{
						MimeType: mime.MimeTypeVP8.String(),
						Layers: []*samvaad.VideoLayer{
							{Quality: samvaad.VideoQuality_MEDIUM},
							{Quality: samvaad.VideoQuality_HIGH},
						},
					},
				},
			},
			mime.MimeTypeVP8,
			map[string]RidAndLayer{
				"":                 {rid: quarterResolutionQ, layer: 0},
				quarterResolutionQ: {rid: quarterResolutionQ, layer: 0},
				halfResolutionH:    {rid: halfResolutionH, layer: 1},
				fullResolutionF:    {rid: halfResolutionH, layer: 1},
			},
		},
		{
			"three layers",
			&samvaad.TrackInfo{
				Codecs: []*samvaad.SimulcastCodecInfo{
					{
						MimeType: mime.MimeTypeVP8.String(),
						Layers: []*samvaad.VideoLayer{
							{Quality: samvaad.VideoQuality_LOW},
							{Quality: samvaad.VideoQuality_MEDIUM},
							{Quality: samvaad.VideoQuality_HIGH},
						},
					},
				},
			},
			mime.MimeTypeVP8,
			map[string]RidAndLayer{
				"":                 {rid: quarterResolutionQ, layer: 0},
				quarterResolutionQ: {rid: quarterResolutionQ, layer: 0},
				halfResolutionH:    {rid: halfResolutionH, layer: 1},
				fullResolutionF:    {rid: fullResolutionF, layer: 2},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			for testRid, expectedResult := range test.ridToLayer {
				actualLayer := RidToSpatialLayer(test.mimeType, testRid, test.trackInfo, DefaultVideoLayersRid)
				require.Equal(t, expectedResult.layer, actualLayer)

				actualRid := SpatialLayerToRid(test.mimeType, actualLayer, test.trackInfo, DefaultVideoLayersRid)
				require.Equal(t, expectedResult.rid, actualRid)
			}
		})
	}
}

func TestQualityConversion(t *testing.T) {
	type QualityAndLayer struct {
		quality samvaad.VideoQuality
		layer   int32
	}
	tests := []struct {
		name           string
		trackInfo      *samvaad.TrackInfo
		mimeType       mime.MimeType
		qualityToLayer map[samvaad.VideoQuality]QualityAndLayer
	}{
		{
			"no track info",
			nil,
			mime.MimeTypeVP8,
			map[samvaad.VideoQuality]QualityAndLayer{
				samvaad.VideoQuality_LOW:    {quality: samvaad.VideoQuality_LOW, layer: 0},
				samvaad.VideoQuality_MEDIUM: {quality: samvaad.VideoQuality_MEDIUM, layer: 1},
				samvaad.VideoQuality_HIGH:   {quality: samvaad.VideoQuality_HIGH, layer: 2},
			},
		},
		{
			"no layers",
			&samvaad.TrackInfo{},
			mime.MimeTypeVP8,
			map[samvaad.VideoQuality]QualityAndLayer{
				samvaad.VideoQuality_LOW:    {quality: samvaad.VideoQuality_LOW, layer: 0},
				samvaad.VideoQuality_MEDIUM: {quality: samvaad.VideoQuality_MEDIUM, layer: 1},
				samvaad.VideoQuality_HIGH:   {quality: samvaad.VideoQuality_HIGH, layer: 2},
			},
		},
		{
			"single layer, low",
			&samvaad.TrackInfo{
				Codecs: []*samvaad.SimulcastCodecInfo{
					{
						MimeType: mime.MimeTypeVP8.String(),
						Layers: []*samvaad.VideoLayer{
							{Quality: samvaad.VideoQuality_LOW},
						},
					},
				},
			},
			mime.MimeTypeVP8,
			map[samvaad.VideoQuality]QualityAndLayer{
				samvaad.VideoQuality_LOW:    {quality: samvaad.VideoQuality_LOW, layer: 0},
				samvaad.VideoQuality_MEDIUM: {quality: samvaad.VideoQuality_LOW, layer: 0},
				samvaad.VideoQuality_HIGH:   {quality: samvaad.VideoQuality_LOW, layer: 0},
			},
		},
		{
			"single layer, medium",
			&samvaad.TrackInfo{
				Codecs: []*samvaad.SimulcastCodecInfo{
					{
						MimeType: mime.MimeTypeVP8.String(),
						Layers: []*samvaad.VideoLayer{
							{Quality: samvaad.VideoQuality_MEDIUM},
						},
					},
				},
			},
			mime.MimeTypeVP8,
			map[samvaad.VideoQuality]QualityAndLayer{
				samvaad.VideoQuality_LOW:    {quality: samvaad.VideoQuality_MEDIUM, layer: 0},
				samvaad.VideoQuality_MEDIUM: {quality: samvaad.VideoQuality_MEDIUM, layer: 0},
				samvaad.VideoQuality_HIGH:   {quality: samvaad.VideoQuality_MEDIUM, layer: 0},
			},
		},
		{
			"single layer, high",
			&samvaad.TrackInfo{
				Codecs: []*samvaad.SimulcastCodecInfo{
					{
						MimeType: mime.MimeTypeVP8.String(),
						Layers: []*samvaad.VideoLayer{
							{Quality: samvaad.VideoQuality_HIGH},
						},
					},
				},
			},
			mime.MimeTypeVP8,
			map[samvaad.VideoQuality]QualityAndLayer{
				samvaad.VideoQuality_LOW:    {quality: samvaad.VideoQuality_HIGH, layer: 0},
				samvaad.VideoQuality_MEDIUM: {quality: samvaad.VideoQuality_HIGH, layer: 0},
				samvaad.VideoQuality_HIGH:   {quality: samvaad.VideoQuality_HIGH, layer: 0},
			},
		},
		{
			"two layers, low and medium",
			&samvaad.TrackInfo{
				Codecs: []*samvaad.SimulcastCodecInfo{
					{
						MimeType: mime.MimeTypeVP8.String(),
						Layers: []*samvaad.VideoLayer{
							{Quality: samvaad.VideoQuality_LOW},
							{Quality: samvaad.VideoQuality_MEDIUM},
						},
					},
				},
			},
			mime.MimeTypeVP8,
			map[samvaad.VideoQuality]QualityAndLayer{
				samvaad.VideoQuality_LOW:    {quality: samvaad.VideoQuality_LOW, layer: 0},
				samvaad.VideoQuality_MEDIUM: {quality: samvaad.VideoQuality_MEDIUM, layer: 1},
				samvaad.VideoQuality_HIGH:   {quality: samvaad.VideoQuality_MEDIUM, layer: 1},
			},
		},
		{
			"two layers, low and high",
			&samvaad.TrackInfo{
				Codecs: []*samvaad.SimulcastCodecInfo{
					{
						MimeType: mime.MimeTypeVP8.String(),
						Layers: []*samvaad.VideoLayer{
							{Quality: samvaad.VideoQuality_LOW},
							{Quality: samvaad.VideoQuality_HIGH},
						},
					},
				},
			},
			mime.MimeTypeVP8,
			map[samvaad.VideoQuality]QualityAndLayer{
				samvaad.VideoQuality_LOW:    {quality: samvaad.VideoQuality_LOW, layer: 0},
				samvaad.VideoQuality_MEDIUM: {quality: samvaad.VideoQuality_HIGH, layer: 1},
				samvaad.VideoQuality_HIGH:   {quality: samvaad.VideoQuality_HIGH, layer: 1},
			},
		},
		{
			"two layers, medium and high",
			&samvaad.TrackInfo{
				Codecs: []*samvaad.SimulcastCodecInfo{
					{
						MimeType: mime.MimeTypeVP8.String(),
						Layers: []*samvaad.VideoLayer{
							{Quality: samvaad.VideoQuality_MEDIUM},
							{Quality: samvaad.VideoQuality_HIGH},
						},
					},
				},
			},
			mime.MimeTypeVP8,
			map[samvaad.VideoQuality]QualityAndLayer{
				samvaad.VideoQuality_LOW:    {quality: samvaad.VideoQuality_MEDIUM, layer: 0},
				samvaad.VideoQuality_MEDIUM: {quality: samvaad.VideoQuality_MEDIUM, layer: 0},
				samvaad.VideoQuality_HIGH:   {quality: samvaad.VideoQuality_HIGH, layer: 1},
			},
		},
		{
			"three layers",
			&samvaad.TrackInfo{
				Codecs: []*samvaad.SimulcastCodecInfo{
					{
						MimeType: mime.MimeTypeVP8.String(),
						Layers: []*samvaad.VideoLayer{
							{Quality: samvaad.VideoQuality_LOW},
							{Quality: samvaad.VideoQuality_MEDIUM},
							{Quality: samvaad.VideoQuality_HIGH},
						},
					},
				},
			},
			mime.MimeTypeVP8,
			map[samvaad.VideoQuality]QualityAndLayer{
				samvaad.VideoQuality_LOW:    {quality: samvaad.VideoQuality_LOW, layer: 0},
				samvaad.VideoQuality_MEDIUM: {quality: samvaad.VideoQuality_MEDIUM, layer: 1},
				samvaad.VideoQuality_HIGH:   {quality: samvaad.VideoQuality_HIGH, layer: 2},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			for testQuality, expectedResult := range test.qualityToLayer {
				actualLayer := VideoQualityToSpatialLayer(test.mimeType, testQuality, test.trackInfo)
				require.Equal(t, expectedResult.layer, actualLayer)

				actualQuality := SpatialLayerToVideoQuality(test.mimeType, actualLayer, test.trackInfo)
				require.Equal(t, expectedResult.quality, actualQuality)
			}
		})
	}
}

func TestVideoQualityToRidConversion(t *testing.T) {
	tests := []struct {
		name         string
		trackInfo    *samvaad.TrackInfo
		mimeTye      mime.MimeType
		qualityToRid map[samvaad.VideoQuality]string
	}{
		{
			"no track info",
			nil,
			mime.MimeTypeVP8,
			map[samvaad.VideoQuality]string{
				samvaad.VideoQuality_LOW:    quarterResolutionQ,
				samvaad.VideoQuality_MEDIUM: halfResolutionH,
				samvaad.VideoQuality_HIGH:   fullResolutionF,
			},
		},
		{
			"no layers",
			&samvaad.TrackInfo{},
			mime.MimeTypeVP8,
			map[samvaad.VideoQuality]string{
				samvaad.VideoQuality_LOW:    quarterResolutionQ,
				samvaad.VideoQuality_MEDIUM: halfResolutionH,
				samvaad.VideoQuality_HIGH:   fullResolutionF,
			},
		},
		{
			"single layer, low",
			&samvaad.TrackInfo{
				Codecs: []*samvaad.SimulcastCodecInfo{
					{
						MimeType: mime.MimeTypeVP8.String(),
						Layers: []*samvaad.VideoLayer{
							{Quality: samvaad.VideoQuality_LOW},
						},
					},
				},
			},
			mime.MimeTypeVP8,
			map[samvaad.VideoQuality]string{
				samvaad.VideoQuality_LOW:    quarterResolutionQ,
				samvaad.VideoQuality_MEDIUM: quarterResolutionQ,
				samvaad.VideoQuality_HIGH:   quarterResolutionQ,
			},
		},
		{
			"single layer, medium",
			&samvaad.TrackInfo{
				Codecs: []*samvaad.SimulcastCodecInfo{
					{
						MimeType: mime.MimeTypeVP8.String(),
						Layers: []*samvaad.VideoLayer{
							{Quality: samvaad.VideoQuality_MEDIUM},
						},
					},
				},
			},
			mime.MimeTypeVP8,
			map[samvaad.VideoQuality]string{
				samvaad.VideoQuality_LOW:    quarterResolutionQ,
				samvaad.VideoQuality_MEDIUM: quarterResolutionQ,
				samvaad.VideoQuality_HIGH:   quarterResolutionQ,
			},
		},
		{
			"single layer, high",
			&samvaad.TrackInfo{
				Codecs: []*samvaad.SimulcastCodecInfo{
					{
						MimeType: mime.MimeTypeVP8.String(),
						Layers: []*samvaad.VideoLayer{
							{Quality: samvaad.VideoQuality_HIGH},
						},
					},
				},
			},
			mime.MimeTypeVP8,
			map[samvaad.VideoQuality]string{
				samvaad.VideoQuality_LOW:    quarterResolutionQ,
				samvaad.VideoQuality_MEDIUM: quarterResolutionQ,
				samvaad.VideoQuality_HIGH:   quarterResolutionQ,
			},
		},
		{
			"two layers, low and medium",
			&samvaad.TrackInfo{
				Codecs: []*samvaad.SimulcastCodecInfo{
					{
						MimeType: mime.MimeTypeVP8.String(),
						Layers: []*samvaad.VideoLayer{
							{Quality: samvaad.VideoQuality_LOW},
							{Quality: samvaad.VideoQuality_MEDIUM},
						},
					},
				},
			},
			mime.MimeTypeVP8,
			map[samvaad.VideoQuality]string{
				samvaad.VideoQuality_LOW:    quarterResolutionQ,
				samvaad.VideoQuality_MEDIUM: halfResolutionH,
				samvaad.VideoQuality_HIGH:   halfResolutionH,
			},
		},
		{
			"two layers, low and high",
			&samvaad.TrackInfo{
				Codecs: []*samvaad.SimulcastCodecInfo{
					{
						MimeType: mime.MimeTypeVP8.String(),
						Layers: []*samvaad.VideoLayer{
							{Quality: samvaad.VideoQuality_LOW},
							{Quality: samvaad.VideoQuality_HIGH},
						},
					},
				},
			},
			mime.MimeTypeVP8,
			map[samvaad.VideoQuality]string{
				samvaad.VideoQuality_LOW:    quarterResolutionQ,
				samvaad.VideoQuality_MEDIUM: halfResolutionH,
				samvaad.VideoQuality_HIGH:   halfResolutionH,
			},
		},
		{
			"two layers, medium and high",
			&samvaad.TrackInfo{
				Codecs: []*samvaad.SimulcastCodecInfo{
					{
						MimeType: mime.MimeTypeVP8.String(),
						Layers: []*samvaad.VideoLayer{
							{Quality: samvaad.VideoQuality_MEDIUM},
							{Quality: samvaad.VideoQuality_HIGH},
						},
					},
				},
			},
			mime.MimeTypeVP8,
			map[samvaad.VideoQuality]string{
				samvaad.VideoQuality_LOW:    quarterResolutionQ,
				samvaad.VideoQuality_MEDIUM: quarterResolutionQ,
				samvaad.VideoQuality_HIGH:   halfResolutionH,
			},
		},
		{
			"three layers",
			&samvaad.TrackInfo{
				Codecs: []*samvaad.SimulcastCodecInfo{
					{
						MimeType: mime.MimeTypeVP8.String(),
						Layers: []*samvaad.VideoLayer{
							{Quality: samvaad.VideoQuality_LOW},
							{Quality: samvaad.VideoQuality_MEDIUM},
							{Quality: samvaad.VideoQuality_HIGH},
						},
					},
				},
			},
			mime.MimeTypeVP8,
			map[samvaad.VideoQuality]string{
				samvaad.VideoQuality_LOW:    quarterResolutionQ,
				samvaad.VideoQuality_MEDIUM: halfResolutionH,
				samvaad.VideoQuality_HIGH:   fullResolutionF,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			for testQuality, expectedRid := range test.qualityToRid {
				actualRid := VideoQualityToRid(test.mimeTye, testQuality, test.trackInfo, DefaultVideoLayersRid)
				require.Equal(t, expectedRid, actualRid)
			}
		})
	}
}

func TestGetSpatialLayerForRid(t *testing.T) {
	tests := []struct {
		name              string
		trackInfo         *samvaad.TrackInfo
		mimeType          mime.MimeType
		ridToSpatialLayer map[string]int32
	}{
		{
			"no track info",
			nil,
			mime.MimeTypeVP8,
			map[string]int32{
				quarterResolutionQ: InvalidLayerSpatial,
				halfResolutionH:    InvalidLayerSpatial,
				fullResolutionF:    InvalidLayerSpatial,
			},
		},
		{
			"no layers",
			&samvaad.TrackInfo{},
			mime.MimeTypeVP8,
			map[string]int32{
				// SIMULCAST-CODEC-TODO
				// quarterResolutionQ: InvalidLayerSpatial,
				// halfResolutionH:    InvalidLayerSpatial,
				// fullResolutionF:    InvalidLayerSpatial,
				quarterResolutionQ: 0,
				halfResolutionH:    0,
				fullResolutionF:    0,
			},
		},
		{
			"no rid",
			&samvaad.TrackInfo{},
			mime.MimeTypeVP8,
			map[string]int32{
				"": 0,
			},
		},
		{
			"single layer",
			&samvaad.TrackInfo{
				Codecs: []*samvaad.SimulcastCodecInfo{
					{
						MimeType: mime.MimeTypeVP8.String(),
						Layers: []*samvaad.VideoLayer{
							{Quality: samvaad.VideoQuality_LOW, SpatialLayer: 0},
						},
					},
				},
			},
			mime.MimeTypeVP8,
			map[string]int32{
				quarterResolutionQ: 0,
				halfResolutionH:    0,
				fullResolutionF:    0,
			},
		},
		{
			"layers",
			&samvaad.TrackInfo{
				Codecs: []*samvaad.SimulcastCodecInfo{
					{
						MimeType: mime.MimeTypeVP8.String(),
						Layers: []*samvaad.VideoLayer{
							{Quality: samvaad.VideoQuality_LOW, SpatialLayer: 0, Rid: quarterResolutionQ},
							{Quality: samvaad.VideoQuality_MEDIUM, SpatialLayer: 1, Rid: halfResolutionH},
						},
					},
				},
			},
			mime.MimeTypeVP8,
			map[string]int32{
				quarterResolutionQ: 0,
				halfResolutionH:    1,
				// SIMULCAST-CODEC-TODO
				// fullResolutionF:    InvalidLayerSpatial,
				fullResolutionF: 0,
			},
		},
		{
			"layers - no rid",
			&samvaad.TrackInfo{
				Codecs: []*samvaad.SimulcastCodecInfo{
					{
						MimeType: mime.MimeTypeVP8.String(),
						Layers: []*samvaad.VideoLayer{
							{Quality: samvaad.VideoQuality_LOW, SpatialLayer: 0},
							{Quality: samvaad.VideoQuality_MEDIUM, SpatialLayer: 1},
						},
					},
				},
			},
			mime.MimeTypeVP8,
			map[string]int32{
				quarterResolutionQ: 0,
				halfResolutionH:    0,
				fullResolutionF:    0,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			for testRid, expectedSpatialLayer := range test.ridToSpatialLayer {
				actualSpatialLayer := GetSpatialLayerForRid(test.mimeType, testRid, test.trackInfo)
				require.Equal(t, expectedSpatialLayer, actualSpatialLayer)
			}
		})
	}
}

func TestGetSpatialLayerForVideoQuality(t *testing.T) {
	tests := []struct {
		name                       string
		trackInfo                  *samvaad.TrackInfo
		mimeType                   mime.MimeType
		videoQualityToSpatialLayer map[samvaad.VideoQuality]int32
	}{
		{
			"no track info",
			nil,
			mime.MimeTypeVP8,
			map[samvaad.VideoQuality]int32{
				samvaad.VideoQuality_LOW:    InvalidLayerSpatial,
				samvaad.VideoQuality_MEDIUM: InvalidLayerSpatial,
				samvaad.VideoQuality_HIGH:   InvalidLayerSpatial,
				samvaad.VideoQuality_OFF:    InvalidLayerSpatial,
			},
		},
		{
			"no layers",
			&samvaad.TrackInfo{},
			mime.MimeTypeVP8,
			map[samvaad.VideoQuality]int32{
				samvaad.VideoQuality_LOW:    0,
				samvaad.VideoQuality_MEDIUM: 0,
				samvaad.VideoQuality_HIGH:   0,
				samvaad.VideoQuality_OFF:    InvalidLayerSpatial,
			},
		},
		{
			"not all layers",
			&samvaad.TrackInfo{
				Codecs: []*samvaad.SimulcastCodecInfo{
					{
						MimeType: mime.MimeTypeVP8.String(),
						Layers: []*samvaad.VideoLayer{
							{Quality: samvaad.VideoQuality_LOW, SpatialLayer: 0, Rid: quarterResolutionQ},
							{Quality: samvaad.VideoQuality_MEDIUM, SpatialLayer: 1, Rid: halfResolutionH},
						},
					},
				},
			},
			mime.MimeTypeVP8,
			map[samvaad.VideoQuality]int32{
				samvaad.VideoQuality_LOW:    0,
				samvaad.VideoQuality_MEDIUM: 1,
				samvaad.VideoQuality_HIGH:   1,
				samvaad.VideoQuality_OFF:    InvalidLayerSpatial,
			},
		},
		{
			"all layers",
			&samvaad.TrackInfo{
				Codecs: []*samvaad.SimulcastCodecInfo{
					{
						MimeType: mime.MimeTypeVP8.String(),
						Layers: []*samvaad.VideoLayer{
							{Quality: samvaad.VideoQuality_LOW, SpatialLayer: 0, Rid: quarterResolutionQ},
							{Quality: samvaad.VideoQuality_MEDIUM, SpatialLayer: 1, Rid: halfResolutionH},
							{Quality: samvaad.VideoQuality_HIGH, SpatialLayer: 2, Rid: fullResolutionF},
						},
					},
				},
			},
			mime.MimeTypeVP8,
			map[samvaad.VideoQuality]int32{
				samvaad.VideoQuality_LOW:    0,
				samvaad.VideoQuality_MEDIUM: 1,
				samvaad.VideoQuality_HIGH:   2,
				samvaad.VideoQuality_OFF:    InvalidLayerSpatial,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			for testVideoQuality, expectedSpatialLayer := range test.videoQualityToSpatialLayer {
				actualSpatialLayer := GetSpatialLayerForVideoQuality(test.mimeType, testVideoQuality, test.trackInfo)
				require.Equal(t, expectedSpatialLayer, actualSpatialLayer)
			}
		})
	}
}

func TestGetVideoQualityorSpatialLayer(t *testing.T) {
	tests := []struct {
		name                       string
		trackInfo                  *samvaad.TrackInfo
		mimeType                   mime.MimeType
		spatialLayerToVideoQuality map[int32]samvaad.VideoQuality
	}{
		{
			"no track info",
			nil,
			mime.MimeTypeVP8,
			map[int32]samvaad.VideoQuality{
				InvalidLayerSpatial: samvaad.VideoQuality_OFF,
				0:                   samvaad.VideoQuality_OFF,
				1:                   samvaad.VideoQuality_OFF,
				2:                   samvaad.VideoQuality_OFF,
			},
		},
		{
			"no layers",
			&samvaad.TrackInfo{},
			mime.MimeTypeVP8,
			map[int32]samvaad.VideoQuality{
				InvalidLayerSpatial: samvaad.VideoQuality_OFF,
				0:                   samvaad.VideoQuality_OFF,
				1:                   samvaad.VideoQuality_OFF,
				2:                   samvaad.VideoQuality_OFF,
			},
		},
		{
			"layers",
			&samvaad.TrackInfo{
				Codecs: []*samvaad.SimulcastCodecInfo{
					{
						MimeType: mime.MimeTypeVP8.String(),
						Layers: []*samvaad.VideoLayer{
							{Quality: samvaad.VideoQuality_LOW, SpatialLayer: 0, Rid: quarterResolutionQ},
							{Quality: samvaad.VideoQuality_MEDIUM, SpatialLayer: 1, Rid: halfResolutionH},
						},
					},
				},
			},
			mime.MimeTypeVP8,
			map[int32]samvaad.VideoQuality{
				InvalidLayerSpatial: samvaad.VideoQuality_OFF,
				0:                   samvaad.VideoQuality_LOW,
				1:                   samvaad.VideoQuality_MEDIUM,
				2:                   samvaad.VideoQuality_OFF,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			for testSpatialLayer, expectedVideoQuality := range test.spatialLayerToVideoQuality {
				actualVideoQuality := GetVideoQualityForSpatialLayer(test.mimeType, testSpatialLayer, test.trackInfo)
				require.Equal(t, expectedVideoQuality, actualVideoQuality)
			}
		})
	}
}

func TestNormalizeVideoLayersRid(t *testing.T) {
	tests := []struct {
		name       string
		rids       VideoLayersRid
		normalized VideoLayersRid
	}{
		{
			"empty",
			VideoLayersRid{},
			VideoLayersRid{},
		},
		{
			"unknown pattern",
			VideoLayersRid{"3", "2", "1"},
			VideoLayersRid{"3", "2", "1"},
		},
		{
			"qhf",
			videoLayersRidQHF,
			videoLayersRidQHF,
		},
		{
			"scrambled qhf",
			VideoLayersRid{"f", "h", "q"},
			videoLayersRidQHF,
		},
		{
			"partial qhf",
			VideoLayersRid{"h", "q"},
			VideoLayersRid{"q", "h", ""},
		},
		{
			"210",
			videoLayersRid210,
			videoLayersRid210,
		},
		{
			"scrambled 210",
			VideoLayersRid{"2", "0", "1"},
			videoLayersRid210,
		},
		{
			"partial 210",
			VideoLayersRid{"1", "2"},
			VideoLayersRid{"2", "1", ""},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			normalizedRids := NormalizeVideoLayersRid(test.rids)
			require.Equal(t, test.normalized, normalizedRids)
		})
	}
}


