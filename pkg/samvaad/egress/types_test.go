// Copyright 2025 Samvaad, Inc.
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

import (
	"testing"

	"github.com/msmclass/samvaad/pkg/proto/samvaad"
	"github.com/stretchr/testify/require"
)

func TestGetOutputType(t *testing.T) {
	roomReq := &samvaad.RoomCompositeEgressRequest{
		FileOutputs: []*samvaad.EncodedFileOutput{
			&samvaad.EncodedFileOutput{},
		},
	}

	ot := GetOutputType(roomReq)
	require.Equal(t, OutputTypeFile, ot)

	roomReq = &samvaad.RoomCompositeEgressRequest{
		Output: &samvaad.RoomCompositeEgressRequest_File{
			File: &samvaad.EncodedFileOutput{},
		},
	}

	ot = GetOutputType(roomReq)
	require.Equal(t, OutputTypeFile, ot)

	trackReq := &samvaad.TrackCompositeEgressRequest{
		SegmentOutputs: []*samvaad.SegmentedFileOutput{
			&samvaad.SegmentedFileOutput{},
		},
	}

	ot = GetOutputType(trackReq)
	require.Equal(t, OutputTypeSegments, ot)

	trackReq = &samvaad.TrackCompositeEgressRequest{
		Output: &samvaad.TrackCompositeEgressRequest_Segments{
			Segments: &samvaad.SegmentedFileOutput{},
		},
	}

	ot = GetOutputType(trackReq)
	require.Equal(t, OutputTypeSegments, ot)

	webReq := &samvaad.WebEgressRequest{
		StreamOutputs: []*samvaad.StreamOutput{
			&samvaad.StreamOutput{},
		},
	}

	ot = GetOutputType(webReq)
	require.Equal(t, OutputTypeStream, ot)

	webReq = &samvaad.WebEgressRequest{
		Output: &samvaad.WebEgressRequest_Stream{
			Stream: &samvaad.StreamOutput{},
		},
	}

	ot = GetOutputType(webReq)
	require.Equal(t, OutputTypeStream, ot)

	participantReq := &samvaad.ParticipantEgressRequest{
		ImageOutputs: []*samvaad.ImageOutput{
			&samvaad.ImageOutput{},
		},
	}

	ot = GetOutputType(participantReq)
	require.Equal(t, OutputTypeImages, ot)

	participantReq.SegmentOutputs = []*samvaad.SegmentedFileOutput{
		&samvaad.SegmentedFileOutput{},
	}

	ot = GetOutputType(participantReq)
	require.Equal(t, OutputTypeMultiple, ot)

}
