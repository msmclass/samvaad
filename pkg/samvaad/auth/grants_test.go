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

package auth

import (
	"reflect"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/msmclass/samvaad/pkg/proto/samvaad"
)

func TestGrants(t *testing.T) {
	t.Parallel()

	t.Run("clone default grant", func(t *testing.T) {
		grants := &ClaimGrants{}
		clone := grants.Clone()
		require.NotSame(t, grants, clone)
		require.Same(t, grants.Video, clone.Video)
		require.Same(t, grants.Agent, clone.Agent)
		require.Same(t, grants.Inference, clone.Inference)
		require.Same(t, grants.SIP, clone.SIP)
		require.True(t, reflect.DeepEqual(grants, clone))
		require.True(t, reflect.DeepEqual(grants.Video, clone.Video))
	})

	t.Run("clone nil video", func(t *testing.T) {
		grants := &ClaimGrants{
			Identity: "identity",
			Name:     "name",
			Sha256:   "sha256",
			Metadata: "metadata",
		}
		clone := grants.Clone()
		require.NotSame(t, grants, clone)
		require.Same(t, grants.Video, clone.Video)
		require.True(t, reflect.DeepEqual(grants, clone))
		require.True(t, reflect.DeepEqual(grants.Video, clone.Video))

		// require SIP
		require.Same(t, grants.SIP, clone.SIP)
		require.True(t, reflect.DeepEqual(grants.SIP, clone.SIP))
		// require Agent
		require.Same(t, grants.Agent, clone.Agent)
		require.True(t, reflect.DeepEqual(grants.Agent, clone.Agent))
		// require Inference
		require.Same(t, grants.Inference, clone.Inference)
		require.True(t, reflect.DeepEqual(grants.Inference, clone.Inference))
	})

	t.Run("clone with video", func(t *testing.T) {
		tr := true
		fa := false
		video := &VideoGrant{
			RoomCreate:          true,
			RoomList:            false,
			RoomRecord:          true,
			RoomAdmin:           false,
			RoomJoin:            true,
			Room:                "room",
			CanPublish:          &tr,
			CanSubscribe:        &fa,
			CanPublishData:      nil,
			Hidden:              true,
			Recorder:            false,
			CanSubscribeMetrics: &tr,
		}
		grants := &ClaimGrants{
			Identity: "identity",
			Name:     "name",
			Kind:     "kind",
			Video:    video,
			Sha256:   "sha256",
			Metadata: "metadata",
		}
		clone := grants.Clone()
		require.NotSame(t, grants, clone)
		require.NotSame(t, grants.Video, clone.Video)
		require.NotSame(t, grants.Video.CanPublish, clone.Video.CanPublish)
		require.NotSame(t, grants.Video.CanSubscribe, clone.Video.CanSubscribe)
		require.Same(t, grants.Video.CanPublishData, clone.Video.CanPublishData)
		require.True(t, reflect.DeepEqual(grants, clone))
		require.True(t, reflect.DeepEqual(grants.Video, clone.Video))
	})

	t.Run("clone with SIP", func(t *testing.T) {
		sip := &SIPGrant{
			Admin: true,
		}
		grants := &ClaimGrants{
			Identity: "identity",
			Name:     "name",
			Kind:     "kind",
			SIP:      sip,
			Sha256:   "sha256",
			Metadata: "metadata",
		}
		clone := grants.Clone()
		require.NotSame(t, grants, clone)
		require.NotSame(t, grants.SIP, clone.SIP)
		require.Equal(t, grants.SIP.Admin, clone.SIP.Admin)
		require.True(t, reflect.DeepEqual(grants, clone))
		require.True(t, reflect.DeepEqual(grants.SIP, clone.SIP))
	})

	t.Run("clone with Agent", func(t *testing.T) {
		agent := &AgentGrant{
			Admin: true,
		}
		grants := &ClaimGrants{
			Identity: "identity",
			Name:     "name",
			Kind:     "kind",
			Agent:    agent,
			Sha256:   "sha256",
			Metadata: "metadata",
		}
		clone := grants.Clone()
		require.NotSame(t, grants, clone)
		require.NotSame(t, grants.Agent, clone.Agent)
		require.Equal(t, grants.Agent.Admin, clone.Agent.Admin)
		require.True(t, reflect.DeepEqual(grants, clone))
		require.True(t, reflect.DeepEqual(grants.Agent, clone.Agent))
	})

	t.Run("clone with Inference", func(t *testing.T) {
		inference := &InferenceGrant{
			Perform: true,
		}
		grants := &ClaimGrants{
			Identity:  "identity",
			Name:      "name",
			Kind:      "kind",
			Inference: inference,
			Sha256:    "sha256",
			Metadata:  "metadata",
		}
		clone := grants.Clone()
		require.NotSame(t, grants, clone)
		require.NotSame(t, grants.Inference, clone.Inference)
		require.Equal(t, grants.Inference.Perform, clone.Inference.Perform)
		require.True(t, reflect.DeepEqual(grants, clone))
		require.True(t, reflect.DeepEqual(grants.Inference, clone.Inference))
	})
}

func TestParticipantKind(t *testing.T) {
	const kindMin, kindMax = samvaad.ParticipantInfo_STANDARD, samvaad.ParticipantInfo_AGENT
	for k := kindMin; k <= kindMax; k++ {
		k := k
		t.Run(k.String(), func(t *testing.T) {
			require.Equal(t, k, kindToProto(kindFromProto(k)))
		})
	}
	const kindNext = kindMax + 1
	if _, err := strconv.Atoi(kindNext.String()); err != nil {
		t.Errorf("Please update kindMax to match protobuf. Missing value: %s", kindNext)
	}
}

func TestParticipantKindDetail(t *testing.T) {
	const detailMin, detailMax = samvaad.ParticipantInfo_CLOUD_AGENT, samvaad.ParticipantInfo_CONNECTOR_TWILIO
	var details []samvaad.ParticipantInfo_KindDetail
	for k := detailMin; k <= detailMax; k++ {
		details = append(details, k)
	}

	require.EqualValues(t, details, kindDetailsToProto(kindDetailsFromProto(details)))
}

func TestRoomConfiguration_CheckCredentials(t *testing.T) {
	t.Parallel()

	t.Run("nil egress returns nil", func(t *testing.T) {
		config := &RoomConfiguration{}
		require.NoError(t, config.CheckCredentials())
	})

	t.Run("empty egress returns nil", func(t *testing.T) {
		config := &RoomConfiguration{
			Egress: &samvaad.RoomEgress{},
		}
		require.NoError(t, config.CheckCredentials())
	})

	t.Run("participant file output with S3 secret fails", func(t *testing.T) {
		config := &RoomConfiguration{
			Egress: &samvaad.RoomEgress{
				Participant: &samvaad.AutoParticipantEgress{
					FileOutputs: []*samvaad.EncodedFileOutput{
						{
							Output: &samvaad.EncodedFileOutput_S3{
								S3: &samvaad.S3Upload{
									AccessKey: "access",
									Secret:    "secret", // This should trigger error
									Bucket:    "bucket",
								},
							},
						},
					},
				},
			},
		}
		require.ErrorIs(t, config.CheckCredentials(), ErrSensitiveCredentials)
	})

	t.Run("participant file output with S3 but no secret passes", func(t *testing.T) {
		config := &RoomConfiguration{
			Egress: &samvaad.RoomEgress{
				Participant: &samvaad.AutoParticipantEgress{
					FileOutputs: []*samvaad.EncodedFileOutput{
						{
							Output: &samvaad.EncodedFileOutput_S3{
								S3: &samvaad.S3Upload{
									AccessKey: "access",
									Secret:    "", // No secret
									Bucket:    "bucket",
									Region:    "us-west-2",
								},
							},
						},
					},
				},
			},
		}
		require.NoError(t, config.CheckCredentials())
	})

	t.Run("participant segment output with GCP credentials fails", func(t *testing.T) {
		config := &RoomConfiguration{
			Egress: &samvaad.RoomEgress{
				Participant: &samvaad.AutoParticipantEgress{
					SegmentOutputs: []*samvaad.SegmentedFileOutput{
						{
							Output: &samvaad.SegmentedFileOutput_Gcp{
								Gcp: &samvaad.GCPUpload{
									Credentials: "credentials", // This should trigger error
									Bucket:      "bucket",
								},
							},
						},
					},
				},
			},
		}
		require.ErrorIs(t, config.CheckCredentials(), ErrSensitiveCredentials)
	})

	t.Run("participant segment output with GCP but no credentials passes", func(t *testing.T) {
		config := &RoomConfiguration{
			Egress: &samvaad.RoomEgress{
				Participant: &samvaad.AutoParticipantEgress{
					SegmentOutputs: []*samvaad.SegmentedFileOutput{
						{
							Output: &samvaad.SegmentedFileOutput_Gcp{
								Gcp: &samvaad.GCPUpload{
									Credentials: "", // No credentials
									Bucket:      "bucket",
								},
							},
						},
					},
				},
			},
		}
		require.NoError(t, config.CheckCredentials())
	})

	t.Run("room file output with Azure account key fails", func(t *testing.T) {
		config := &RoomConfiguration{
			Egress: &samvaad.RoomEgress{
				Room: &samvaad.RoomCompositeEgressRequest{
					FileOutputs: []*samvaad.EncodedFileOutput{
						{
							Output: &samvaad.EncodedFileOutput_Azure{
								Azure: &samvaad.AzureBlobUpload{
									AccountName:   "account",
									AccountKey:    "key", // This should trigger error
									ContainerName: "container",
								},
							},
						},
					},
				},
			},
		}
		require.ErrorIs(t, config.CheckCredentials(), ErrSensitiveCredentials)
	})

	t.Run("room segment output with AliOSS secret fails", func(t *testing.T) {
		config := &RoomConfiguration{
			Egress: &samvaad.RoomEgress{
				Room: &samvaad.RoomCompositeEgressRequest{
					SegmentOutputs: []*samvaad.SegmentedFileOutput{
						{
							Output: &samvaad.SegmentedFileOutput_AliOSS{
								AliOSS: &samvaad.AliOSSUpload{
									AccessKey: "access",
									Secret:    "secret", // This should trigger error
									Bucket:    "bucket",
								},
							},
						},
					},
				},
			},
		}
		require.ErrorIs(t, config.CheckCredentials(), ErrSensitiveCredentials)
	})

	t.Run("room image output with valid config passes", func(t *testing.T) {
		config := &RoomConfiguration{
			Egress: &samvaad.RoomEgress{
				Room: &samvaad.RoomCompositeEgressRequest{
					ImageOutputs: []*samvaad.ImageOutput{
						{
							CaptureInterval: 5,
							Width:           1920,
							Height:          1080,
							Output: &samvaad.ImageOutput_S3{
								S3: &samvaad.S3Upload{
									AccessKey: "access",
									Secret:    "", // No secret
									Bucket:    "bucket",
								},
							},
						},
					},
				},
			},
		}
		require.NoError(t, config.CheckCredentials())
	})

	t.Run("room stream outputs always fail", func(t *testing.T) {
		config := &RoomConfiguration{
			Egress: &samvaad.RoomEgress{
				Room: &samvaad.RoomCompositeEgressRequest{
					StreamOutputs: []*samvaad.StreamOutput{
						{
							Protocol: samvaad.StreamProtocol_RTMP,
							Urls:     []string{"rtmp://example.com/live"},
						},
					},
				},
			},
		}
		require.ErrorIs(t, config.CheckCredentials(), ErrSensitiveCredentials)
	})

	t.Run("tracks output with S3 secret fails", func(t *testing.T) {
		config := &RoomConfiguration{
			Egress: &samvaad.RoomEgress{
				Tracks: &samvaad.AutoTrackEgress{
					Filepath: "output.mp4",
					Output: &samvaad.AutoTrackEgress_S3{
						S3: &samvaad.S3Upload{
							AccessKey: "access",
							Secret:    "secret", // This should trigger error
							Bucket:    "bucket",
						},
					},
				},
			},
		}
		require.ErrorIs(t, config.CheckCredentials(), ErrSensitiveCredentials)
	})

	t.Run("tracks output without credentials passes", func(t *testing.T) {
		config := &RoomConfiguration{
			Egress: &samvaad.RoomEgress{
				Tracks: &samvaad.AutoTrackEgress{
					Filepath: "output.mp4",
					Output: &samvaad.AutoTrackEgress_Gcp{
						Gcp: &samvaad.GCPUpload{
							Credentials: "", // No credentials
							Bucket:      "bucket",
						},
					},
				},
			},
		}
		require.NoError(t, config.CheckCredentials())
	})

	t.Run("multiple outputs with mixed credentials", func(t *testing.T) {
		config := &RoomConfiguration{
			Egress: &samvaad.RoomEgress{
				Participant: &samvaad.AutoParticipantEgress{
					FileOutputs: []*samvaad.EncodedFileOutput{
						{
							Output: &samvaad.EncodedFileOutput_S3{
								S3: &samvaad.S3Upload{
									AccessKey: "access",
									Secret:    "", // No secret - OK
									Bucket:    "bucket1",
								},
							},
						},
						{
							Output: &samvaad.EncodedFileOutput_Gcp{
								Gcp: &samvaad.GCPUpload{
									Credentials: "credentials", // Has credentials - should fail
									Bucket:      "bucket2",
								},
							},
						},
					},
				},
			},
		}
		require.ErrorIs(t, config.CheckCredentials(), ErrSensitiveCredentials)
	})

	t.Run("all cloud providers without credentials pass", func(t *testing.T) {
		config := &RoomConfiguration{
			Egress: &samvaad.RoomEgress{
				Room: &samvaad.RoomCompositeEgressRequest{
					FileOutputs: []*samvaad.EncodedFileOutput{
						{
							Output: &samvaad.EncodedFileOutput_S3{
								S3: &samvaad.S3Upload{
									AccessKey: "access",
									Secret:    "", // No secret
									Bucket:    "s3bucket",
								},
							},
						},
						{
							Output: &samvaad.EncodedFileOutput_Gcp{
								Gcp: &samvaad.GCPUpload{
									Credentials: "", // No credentials
									Bucket:      "gcpbucket",
								},
							},
						},
						{
							Output: &samvaad.EncodedFileOutput_Azure{
								Azure: &samvaad.AzureBlobUpload{
									AccountName:   "account",
									AccountKey:    "", // No key
									ContainerName: "container",
								},
							},
						},
						{
							Output: &samvaad.EncodedFileOutput_AliOSS{
								AliOSS: &samvaad.AliOSSUpload{
									AccessKey: "access",
									Secret:    "", // No secret
									Bucket:    "alibucket",
								},
							},
						},
					},
				},
			},
		}
		require.NoError(t, config.CheckCredentials())
	})
}
