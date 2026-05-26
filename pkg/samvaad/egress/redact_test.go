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

	"google.golang.org/protobuf/proto"

	"github.com/msmclass/samvaad/pkg/proto/samvaad"
	"github.com/stretchr/testify/require"
)

var (
	file = &samvaad.EncodedFileOutput{
		Output: &samvaad.EncodedFileOutput_S3{
			S3: &samvaad.S3Upload{
				AccessKey:            "ACCESS_KEY",
				Secret:               "LONG_SECRET_STRING",
				AssumeRoleExternalId: "EXTERNAL_ID",
				SessionToken:         "SESSION_TOKEN",
			},
		},
	}

	image = &samvaad.ImageOutput{
		Output: &samvaad.ImageOutput_AliOSS{
			AliOSS: &samvaad.AliOSSUpload{
				AccessKey: "ACCESS_KEY",
				Secret:    "LONG_SECRET_STRING",
			},
		},
	}

	segments = &samvaad.SegmentedFileOutput{
		Output: &samvaad.SegmentedFileOutput_Gcp{
			Gcp: &samvaad.GCPUpload{
				Credentials: "CREDENTIALS",
			},
		},
	}

	directFile = &samvaad.DirectFileOutput{
		Output: &samvaad.DirectFileOutput_Azure{
			Azure: &samvaad.AzureBlobUpload{
				AccountName: "ACCOUNT_NAME",
				AccountKey:  "ACCOUNT_KEY",
			},
		},
	}
)

func TestRedactUpload(t *testing.T) {
	cl := proto.Clone(file)
	RedactUpload(cl.(UploadRequest))

	require.Equal(t, "{access_key}", cl.(*samvaad.EncodedFileOutput).Output.(*samvaad.EncodedFileOutput_S3).S3.AccessKey)
	require.Equal(t, "{secret}", cl.(*samvaad.EncodedFileOutput).Output.(*samvaad.EncodedFileOutput_S3).S3.Secret)
	require.Equal(t, "{external_id}", cl.(*samvaad.EncodedFileOutput).Output.(*samvaad.EncodedFileOutput_S3).S3.AssumeRoleExternalId)
	require.Equal(t, "{session_token}", cl.(*samvaad.EncodedFileOutput).Output.(*samvaad.EncodedFileOutput_S3).S3.SessionToken)

	cl = proto.Clone(image)
	RedactUpload(cl.(UploadRequest))

	require.Equal(t, "{access_key}", cl.(*samvaad.ImageOutput).Output.(*samvaad.ImageOutput_AliOSS).AliOSS.AccessKey)
	require.Equal(t, "{secret}", cl.(*samvaad.ImageOutput).Output.(*samvaad.ImageOutput_AliOSS).AliOSS.Secret)

	cl = proto.Clone(segments)
	RedactUpload(cl.(UploadRequest))

	require.Equal(t, "{credentials}", cl.(*samvaad.SegmentedFileOutput).Output.(*samvaad.SegmentedFileOutput_Gcp).Gcp.Credentials)

	cl = proto.Clone(directFile)
	RedactUpload(cl.(UploadRequest))

	require.Equal(t, "{account_name}", cl.(*samvaad.DirectFileOutput).Output.(*samvaad.DirectFileOutput_Azure).Azure.AccountName)
	require.Equal(t, "{account_key}", cl.(*samvaad.DirectFileOutput).Output.(*samvaad.DirectFileOutput_Azure).Azure.AccountKey)
}

func TestRedactStreamOutput(t *testing.T) {
	so := &samvaad.StreamOutput{
		Urls: []string{
			"rtmps://foo.bar.com/app/secret_stream_key",
		},
	}

	RedactStreamKeys(so)
	require.Equal(t, "rtmps://foo.bar.com/app/{sec...key}", so.Urls[0])
}

func TestRedactEncodedOutputs(t *testing.T) {
	trackComposite := &samvaad.TrackCompositeEgressRequest{
		FileOutputs: []*samvaad.EncodedFileOutput{
			file,
		},
	}

	cl := proto.Clone(trackComposite)
	RedactEncodedOutputs(cl.(EncodedOutput))

	require.Equal(t, "{access_key}", cl.(*samvaad.TrackCompositeEgressRequest).FileOutputs[0].Output.(*samvaad.EncodedFileOutput_S3).S3.AccessKey)
	require.Equal(t, "{secret}", cl.(*samvaad.TrackCompositeEgressRequest).FileOutputs[0].Output.(*samvaad.EncodedFileOutput_S3).S3.Secret)

	roomComposite := &samvaad.RoomCompositeEgressRequest{
		ImageOutputs: []*samvaad.ImageOutput{
			image,
		},
	}

	cl = proto.Clone(roomComposite)
	RedactEncodedOutputs(cl.(EncodedOutput))

	require.Equal(t, "{access_key}", cl.(*samvaad.RoomCompositeEgressRequest).ImageOutputs[0].Output.(*samvaad.ImageOutput_AliOSS).AliOSS.AccessKey)
	require.Equal(t, "{secret}", cl.(*samvaad.RoomCompositeEgressRequest).ImageOutputs[0].Output.(*samvaad.ImageOutput_AliOSS).AliOSS.Secret)

	participant := &samvaad.ParticipantEgressRequest{
		SegmentOutputs: []*samvaad.SegmentedFileOutput{
			segments,
		},
	}

	cl = proto.Clone(participant)
	RedactEncodedOutputs(cl.(EncodedOutput))

	require.Equal(t, "{credentials}", cl.(*samvaad.ParticipantEgressRequest).SegmentOutputs[0].Output.(*samvaad.SegmentedFileOutput_Gcp).Gcp.Credentials)
}

func TestRedactDirectOutput(t *testing.T) {
	track := &samvaad.TrackEgressRequest{
		Output: &samvaad.TrackEgressRequest_File{
			File: &samvaad.DirectFileOutput{
				Output: &samvaad.DirectFileOutput_S3{
					S3: &samvaad.S3Upload{
						AccessKey: "ACCESS_KEY",
						Secret:    "SECRET",
					},
				},
			},
		},
	}

	RedactDirectOutputs(track)
	require.Equal(t, "{access_key}", track.Output.(*samvaad.TrackEgressRequest_File).File.Output.(*samvaad.DirectFileOutput_S3).S3.AccessKey)
	require.Equal(t, "{secret}", track.Output.(*samvaad.TrackEgressRequest_File).File.Output.(*samvaad.DirectFileOutput_S3).S3.Secret)
}
