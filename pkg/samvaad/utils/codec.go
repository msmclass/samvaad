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
	"github.com/msmclass/samvaad/pkg/proto/samvaad"
	"github.com/msmclass/samvaad/pkg/samvaad/logger"
	"github.com/pion/webrtc/v4"
)

func GetMimeTypeForVideoCodec(codec samvaad.VideoCodec) string {
	switch codec {
	case samvaad.VideoCodec_H264_BASELINE,
		samvaad.VideoCodec_H264_MAIN,
		samvaad.VideoCodec_H264_HIGH:
		return webrtc.MimeTypeH264
	case samvaad.VideoCodec_VP8:
		return webrtc.MimeTypeVP8
	default:
		logger.Errorw("GetMimeTypeForVideoCodec unimplemented", nil, "codec", codec.String())
		return ""
	}
}

func GetMimeTypeForAudioCodec(codec samvaad.AudioCodec) string {
	// AAC is not supported in webrtc

	switch codec {
	case samvaad.AudioCodec_OPUS:
		return webrtc.MimeTypeOpus
	default:
		logger.Errorw("GetMimeTypeForAudioCodec unimplemented", nil, "codec", codec.String())
		return ""
	}
}
