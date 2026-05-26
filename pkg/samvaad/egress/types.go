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

package egress

import "github.com/msmclass/samvaad/pkg/proto/samvaad"

const (
	EgressTypeTemplate = "template"
	EgressTypeWeb      = "web"
	EgressTypeMedia    = "media"

	EgressTypeRoomComposite  = "room_composite"
	EgressTypeParticipant    = "participant"
	EgressTypeTrackComposite = "track_composite"
	EgressTypeTrack          = "track"

	OutputTypeFile     = "file"
	OutputTypeStream   = "stream"
	OutputTypeSegments = "segments"
	OutputTypeImages   = "images"
	OutputTypeMultiple = "multiple"

	Unknown = "unknown"
)

// Outputs that can be used in egress that are started automatically on room creation
type AutoEncodedOutput interface {
	GetFileOutputs() []*samvaad.EncodedFileOutput
	GetSegmentOutputs() []*samvaad.SegmentedFileOutput
}

type EncodedOutput interface {
	AutoEncodedOutput
	GetStreamOutputs() []*samvaad.StreamOutput
	GetImageOutputs() []*samvaad.ImageOutput
}

type EncodedOutputDeprecated interface {
	GetFile() *samvaad.EncodedFileOutput
	GetStream() *samvaad.StreamOutput
	GetSegments() *samvaad.SegmentedFileOutput
}

type DirectOutput interface {
	GetFile() *samvaad.DirectFileOutput
	GetWebsocketUrl() string
}

type EgressRequest interface {
	GetMedia() *samvaad.MediaSource
	GetTemplate() *samvaad.TemplateSource
	GetWeb() *samvaad.WebSource
	GetOutputs() []*samvaad.Output
	GetStorage() *samvaad.StorageConfig
}

type UploadRequest interface {
	GetS3() *samvaad.S3Upload
	GetGcp() *samvaad.GCPUpload
	GetAzure() *samvaad.AzureBlobUpload
	GetAliOSS() *samvaad.AliOSSUpload
}

func GetTypes(request any) (string, string) {
	switch req := request.(type) {
	// case *samvaad.EgressInfo_Egress:
	// 	return getSourceTypeV2(req.Egress), GetOutputTypeV2(req.Egress.Outputs)

	case *samvaad.EgressInfo_Replay:
		return getSourceTypeV2(req.Replay), GetOutputTypeV2(req.Replay.Outputs)
	
	case *samvaad.EgressInfo_RoomComposite:
		return EgressTypeRoomComposite, GetOutputType(req.RoomComposite)

	case *samvaad.EgressInfo_Web:
		return EgressTypeWeb, GetOutputType(req.Web)

	case *samvaad.EgressInfo_Participant:
		return EgressTypeParticipant, GetOutputType(req.Participant)

	case *samvaad.EgressInfo_TrackComposite:
		return EgressTypeTrackComposite, GetOutputType(req.TrackComposite)

	case *samvaad.EgressInfo_Track:
		return EgressTypeTrack, GetOutputType(req.Track)
	}

	return Unknown, Unknown
}

func getSourceTypeV2(req EgressRequest) string {
	if req.GetMedia() != nil {
		return EgressTypeMedia
	} else if req.GetTemplate() != nil {
		return EgressTypeTemplate
	} else if req.GetWeb() != nil {
		return EgressTypeWeb
	}
	return Unknown
}

func GetOutputTypeV2(outputs []*samvaad.Output) string {
	if len(outputs) == 0 {
		return Unknown
	}
	if len(outputs) > 1 {
		return OutputTypeMultiple
	}
	switch outputs[0].Config.(type) {
	case *samvaad.Output_File:
		return OutputTypeFile
	case *samvaad.Output_Stream:
		return OutputTypeStream
	case *samvaad.Output_Segments:
		return OutputTypeSegments
	case *samvaad.Output_Images:
		return OutputTypeImages
	default:
		return Unknown
	}
}

func GetOutputType(req interface{}) string {
	if r, ok := req.(EncodedOutput); ok {
		outputs := make([]string, 0)
		if len(r.GetFileOutputs()) > 0 {
			outputs = append(outputs, OutputTypeFile)
		}
		if len(r.GetStreamOutputs()) > 0 {
			outputs = append(outputs, OutputTypeStream)
		}
		if len(r.GetSegmentOutputs()) > 0 {
			outputs = append(outputs, OutputTypeSegments)
		}
		if len(r.GetImageOutputs()) > 0 {
			outputs = append(outputs, OutputTypeImages)
		}

		switch len(outputs) {
		default:
			return OutputTypeMultiple
		case 1:
			return outputs[0]
		case 0:
			if r, ok := req.(EncodedOutputDeprecated); ok {
				if r.GetFile() != nil {
					return OutputTypeFile
				}
				if r.GetStream() != nil {
					return OutputTypeStream
				}
				if r.GetSegments() != nil {
					return OutputTypeSegments
				}
			}
		}
	}

	if r, ok := req.(DirectOutput); ok {
		if r.GetFile() != nil {
			return OutputTypeFile
		}
		if r.GetWebsocketUrl() != "" {
			return OutputTypeStream
		}
	}

	return Unknown
}
