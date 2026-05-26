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

package ingress

import (
	"testing"

	"github.com/msmclass/samvaad/pkg/proto/samvaad"
	"github.com/stretchr/testify/require"
)

func TestValidate(t *testing.T) {
	info := &samvaad.IngressInfo{}

	err := Validate(info)
	require.Error(t, err)

	info.StreamKey = "stream_key"
	err = Validate(info)
	require.Error(t, err)

	info.RoomName = "room_name"
	err = Validate(info)
	require.Error(t, err)

	info.ParticipantIdentity = "participant_identity"
	err = Validate(info)
	require.NoError(t, err)

	// make sure video parameters are validated. Full validation logic tested in the next test
	info.Video = &samvaad.IngressVideoOptions{}
	err = Validate(info)
	require.NoError(t, err)

	info.Video.Source = samvaad.TrackSource_MICROPHONE
	err = Validate(info)
	require.Error(t, err)

	info.Video.Source = samvaad.TrackSource_CAMERA

	// make sure audio parameters are validated. Full validation logic tested in the next test
	info.Audio = &samvaad.IngressAudioOptions{}
	err = Validate(info)
	require.NoError(t, err)

	info.Audio.Source = samvaad.TrackSource_CAMERA
	err = Validate(info)
	require.Error(t, err)

	info.Audio.Source = samvaad.TrackSource_SCREEN_SHARE_AUDIO
	err = Validate(info)
	require.NoError(t, err)
}

func TestValidateBypassTranscoding(t *testing.T) {
	info := &samvaad.IngressInfo{}

	err := ValidateBypassTranscoding(info)
	require.NoError(t, err)

	info.BypassTranscoding = true
	err = ValidateBypassTranscoding(info)
	require.Error(t, err)

	info.InputType = samvaad.IngressInput_WHIP_INPUT
	err = ValidateBypassTranscoding(info)
	require.NoError(t, err)

	info.Video = &samvaad.IngressVideoOptions{}
	err = ValidateBypassTranscoding(info)
	require.NoError(t, err)

	info.Video.EncodingOptions = &samvaad.IngressVideoOptions_Preset{}
	err = ValidateBypassTranscoding(info)
	require.Error(t, err)

	info.Video = nil

	info.Audio = &samvaad.IngressAudioOptions{}
	err = ValidateBypassTranscoding(info)
	require.NoError(t, err)

	info.Audio.EncodingOptions = &samvaad.IngressAudioOptions_Options{
		Options: &samvaad.IngressAudioEncodingOptions{},
	}
	err = ValidateBypassTranscoding(info)
	require.Error(t, err)

}

func TestValidateEnableTranscoding(t *testing.T) {
	info := &samvaad.IngressInfo{}
	T := true
	F := false

	err := ValidateEnableTranscoding(info)
	require.NoError(t, err)

	info.InputType = samvaad.IngressInput_WHIP_INPUT
	err = ValidateEnableTranscoding(info)
	require.NoError(t, err)

	info.Audio = &samvaad.IngressAudioOptions{}
	info.Audio.EncodingOptions = &samvaad.IngressAudioOptions_Options{}
	err = ValidateEnableTranscoding(info)
	require.Error(t, err)

	info.Audio.EncodingOptions = nil

	info.EnableTranscoding = &T
	err = ValidateEnableTranscoding(info)
	require.NoError(t, err)

	info.EnableTranscoding = &F
	err = ValidateEnableTranscoding(info)
	require.NoError(t, err)

	info.Video = &samvaad.IngressVideoOptions{}
	info.Video.EncodingOptions = &samvaad.IngressVideoOptions_Preset{}
	err = ValidateEnableTranscoding(info)
	require.Error(t, err)

	info.Video.EncodingOptions = nil

	info.InputType = samvaad.IngressInput_RTMP_INPUT
	err = ValidateEnableTranscoding(info)
	require.Error(t, err)

	info.EnableTranscoding = &T
	err = ValidateEnableTranscoding(info)
	require.NoError(t, err)
}

func TestValidateEnabled(t *testing.T) {
	info := &samvaad.IngressInfo{
		StreamKey:           "sk",
		RoomName:            "room_name",
		ParticipantIdentity: "id",
	}
	T := true
	F := false

	err := Validate(info)
	require.NoError(t, err)

	info.Enabled = &T
	err = Validate(info)
	require.NoError(t, err)

	info.Enabled = &F
	err = Validate(info)
	require.NoError(t, err)

	info.InputType = samvaad.IngressInput_URL_INPUT
	info.Url = "url"
	info.Enabled = nil
	err = Validate(info)
	require.NoError(t, err)

	info.Enabled = &T
	err = Validate(info)
	require.NoError(t, err)

	info.Enabled = &F
	err = Validate(info)
	require.Error(t, err)
}

func TestValidateVideoOptionsConsistency(t *testing.T) {
	video := &samvaad.IngressVideoOptions{}
	err := ValidateVideoOptionsConsistency(video)
	require.NoError(t, err)

	video.Source = samvaad.TrackSource_MICROPHONE
	err = ValidateVideoOptionsConsistency(video)
	require.Error(t, err)

	video.Source = samvaad.TrackSource_CAMERA
	video.EncodingOptions = &samvaad.IngressVideoOptions_Preset{
		Preset: samvaad.IngressVideoEncodingPreset(42),
	}
	err = ValidateVideoOptionsConsistency(video)
	require.Error(t, err)

	video.EncodingOptions = &samvaad.IngressVideoOptions_Preset{
		Preset: samvaad.IngressVideoEncodingPreset_H264_1080P_30FPS_1_LAYER,
	}
	err = ValidateVideoOptionsConsistency(video)
	require.NoError(t, err)

	video.EncodingOptions = &samvaad.IngressVideoOptions_Options{
		Options: &samvaad.IngressVideoEncodingOptions{
			VideoCodec: samvaad.VideoCodec_H264_HIGH,
		},
	}
	err = ValidateVideoOptionsConsistency(video)
	require.Error(t, err)

	video.EncodingOptions = &samvaad.IngressVideoOptions_Options{
		Options: &samvaad.IngressVideoEncodingOptions{
			VideoCodec: samvaad.VideoCodec_DEFAULT_VC,
		},
	}
	err = ValidateVideoOptionsConsistency(video)
	require.NoError(t, err)

	video.EncodingOptions.(*samvaad.IngressVideoOptions_Options).Options.Layers = []*samvaad.VideoLayer{
		&samvaad.VideoLayer{},
	}
	err = ValidateVideoOptionsConsistency(video)
	require.Error(t, err)

	video.EncodingOptions.(*samvaad.IngressVideoOptions_Options).Options.Layers = []*samvaad.VideoLayer{
		&samvaad.VideoLayer{
			Width:  640,
			Height: 480,
		},
	}
	err = ValidateVideoOptionsConsistency(video)
	require.NoError(t, err)

	video.EncodingOptions.(*samvaad.IngressVideoOptions_Options).Options.Layers = []*samvaad.VideoLayer{
		&samvaad.VideoLayer{
			Width:  641,
			Height: 480,
		},
	}
	err = ValidateVideoOptionsConsistency(video)
	require.Error(t, err)

	video.EncodingOptions.(*samvaad.IngressVideoOptions_Options).Options.Layers = []*samvaad.VideoLayer{
		&samvaad.VideoLayer{
			Width:   640,
			Height:  480,
			Quality: samvaad.VideoQuality_HIGH,
		},
		&samvaad.VideoLayer{
			Width:   640,
			Height:  480,
			Quality: samvaad.VideoQuality_LOW,
		},
	}
	err = ValidateVideoOptionsConsistency(video)
	require.NoError(t, err)

	video.EncodingOptions.(*samvaad.IngressVideoOptions_Options).Options.Layers = []*samvaad.VideoLayer{
		&samvaad.VideoLayer{
			Width:   640,
			Height:  480,
			Quality: samvaad.VideoQuality_HIGH,
		},
		&samvaad.VideoLayer{
			Width:   1280,
			Height:  720,
			Quality: samvaad.VideoQuality_HIGH,
		},
	}
	err = ValidateVideoOptionsConsistency(video)
	require.Error(t, err)

	video.EncodingOptions.(*samvaad.IngressVideoOptions_Options).Options.Layers = []*samvaad.VideoLayer{
		&samvaad.VideoLayer{
			Width:   640,
			Height:  480,
			Quality: samvaad.VideoQuality_HIGH,
		},
		&samvaad.VideoLayer{
			Width:   1280,
			Height:  720,
			Quality: samvaad.VideoQuality_LOW,
		},
	}
	err = ValidateVideoOptionsConsistency(video)
	require.Error(t, err)

	video.EncodingOptions.(*samvaad.IngressVideoOptions_Options).Options.Layers = []*samvaad.VideoLayer{
		&samvaad.VideoLayer{
			Width:   640,
			Height:  480,
			Quality: samvaad.VideoQuality_LOW,
		},
		&samvaad.VideoLayer{
			Width:   1280,
			Height:  720,
			Quality: samvaad.VideoQuality_HIGH,
		},
	}

	err = ValidateVideoOptionsConsistency(video)
	require.NoError(t, err)
}

func TestValidateAudioOptionsConsistency(t *testing.T) {
	audio := &samvaad.IngressAudioOptions{}
	err := ValidateAudioOptionsConsistency(audio)
	require.NoError(t, err)

	audio.Source = samvaad.TrackSource_CAMERA
	err = ValidateAudioOptionsConsistency(audio)
	require.Error(t, err)

	audio.Source = samvaad.TrackSource_SCREEN_SHARE_AUDIO
	audio.EncodingOptions = &samvaad.IngressAudioOptions_Preset{
		Preset: samvaad.IngressAudioEncodingPreset(42),
	}
	err = ValidateAudioOptionsConsistency(audio)
	require.Error(t, err)

	audio.EncodingOptions = &samvaad.IngressAudioOptions_Preset{
		Preset: samvaad.IngressAudioEncodingPreset_OPUS_MONO_64KBS,
	}
	err = ValidateAudioOptionsConsistency(audio)
	require.NoError(t, err)

	audio.EncodingOptions = &samvaad.IngressAudioOptions_Options{
		Options: &samvaad.IngressAudioEncodingOptions{
			AudioCodec: samvaad.AudioCodec_AAC,
		},
	}
	err = ValidateAudioOptionsConsistency(audio)
	require.Error(t, err)

	audio.EncodingOptions = &samvaad.IngressAudioOptions_Options{
		Options: &samvaad.IngressAudioEncodingOptions{
			AudioCodec: samvaad.AudioCodec_OPUS,
			Channels:   3,
		},
	}
	err = ValidateAudioOptionsConsistency(audio)
	require.Error(t, err)

	audio.EncodingOptions.(*samvaad.IngressAudioOptions_Options).Options.Channels = 2
	err = ValidateAudioOptionsConsistency(audio)
	require.NoError(t, err)
}
