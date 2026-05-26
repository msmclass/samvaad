package egressobs

import (
	"testing"

	"github.com/msmclass/samvaad/pkg/proto/samvaad"
	"github.com/stretchr/testify/require"
)

func TestGetSourceType(t *testing.T) {
	tests := []struct {
		sourceType samvaad.EgressSourceType
		expected   string
	}{
		{samvaad.EgressSourceType_EGRESS_SOURCE_TYPE_WEB, "web"},
		{samvaad.EgressSourceType_EGRESS_SOURCE_TYPE_SDK, "sdk"},
		{samvaad.EgressSourceType(99), ""}, // Unknown value falls back to undefined (empty string)
	}

	for _, tt := range tests {
		t.Run(tt.sourceType.String(), func(t *testing.T) {
			info := &samvaad.EgressInfo{SourceType: tt.sourceType}
			result := GetSourceType(info)
			require.Equal(t, tt.expected, string(result))
		})
	}
}

func TestGetRequestType(t *testing.T) {
	tests := []struct {
		name     string
		info     *samvaad.EgressInfo
		expected string
	}{
		{
			name: "RoomComposite",
			info: &samvaad.EgressInfo{
				Request: &samvaad.EgressInfo_RoomComposite{
					RoomComposite: &samvaad.RoomCompositeEgressRequest{},
				},
			},
			expected: "room_composite",
		},
		{
			name: "Web",
			info: &samvaad.EgressInfo{
				Request: &samvaad.EgressInfo_Web{
					Web: &samvaad.WebEgressRequest{},
				},
			},
			expected: "web",
		},
		{
			name: "Participant",
			info: &samvaad.EgressInfo{
				Request: &samvaad.EgressInfo_Participant{
					Participant: &samvaad.ParticipantEgressRequest{},
				},
			},
			expected: "participant",
		},
		{
			name: "TrackComposite",
			info: &samvaad.EgressInfo{
				Request: &samvaad.EgressInfo_TrackComposite{
					TrackComposite: &samvaad.TrackCompositeEgressRequest{},
				},
			},
			expected: "track_composite",
		},
		{
			name: "Track",
			info: &samvaad.EgressInfo{
				Request: &samvaad.EgressInfo_Track{
					Track: &samvaad.TrackEgressRequest{},
				},
			},
			expected: "track",
		},
		{
			name:     "Undefined",
			info:     &samvaad.EgressInfo{},
			expected: "", // Undefined is an empty string
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetRequestType(tt.info)
			require.Equal(t, tt.expected, string(result))
		})
	}
}

func TestGetAudioOnly(t *testing.T) {
	tests := []struct {
		name      string
		info      *samvaad.EgressInfo
		audioOnly bool
	}{
		{
			name: "RoomComposite audio only",
			info: &samvaad.EgressInfo{
				Request: &samvaad.EgressInfo_RoomComposite{
					RoomComposite: &samvaad.RoomCompositeEgressRequest{AudioOnly: true},
				},
			},
			audioOnly: true,
		},
		{
			name: "RoomComposite not audio only",
			info: &samvaad.EgressInfo{
				Request: &samvaad.EgressInfo_RoomComposite{
					RoomComposite: &samvaad.RoomCompositeEgressRequest{AudioOnly: false},
				},
			},
			audioOnly: false,
		},
		{
			name: "Web audio only",
			info: &samvaad.EgressInfo{
				Request: &samvaad.EgressInfo_Web{
					Web: &samvaad.WebEgressRequest{AudioOnly: true},
				},
			},
			audioOnly: true,
		},
		{
			name: "Track request returns false",
			info: &samvaad.EgressInfo{
				Request: &samvaad.EgressInfo_Track{
					Track: &samvaad.TrackEgressRequest{},
				},
			},
			audioOnly: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.audioOnly, GetAudioOnly(tt.info))
		})
	}
}

func TestGetRequest(t *testing.T) {
	tests := []struct {
		name string
		info *samvaad.EgressInfo
	}{
		{
			name: "RoomComposite",
			info: &samvaad.EgressInfo{
				Request: &samvaad.EgressInfo_RoomComposite{
					RoomComposite: &samvaad.RoomCompositeEgressRequest{
						RoomName: "test-room",
					},
				},
			},
		},
		{
			name: "Web",
			info: &samvaad.EgressInfo{
				Request: &samvaad.EgressInfo_Web{
					Web: &samvaad.WebEgressRequest{
						Url: "https://example.com",
					},
				},
			},
		},
		{
			name: "Participant",
			info: &samvaad.EgressInfo{
				Request: &samvaad.EgressInfo_Participant{
					Participant: &samvaad.ParticipantEgressRequest{
						RoomName: "test-room",
					},
				},
			},
		},
		{
			name: "TrackComposite",
			info: &samvaad.EgressInfo{
				Request: &samvaad.EgressInfo_TrackComposite{
					TrackComposite: &samvaad.TrackCompositeEgressRequest{
						RoomName: "test-room",
					},
				},
			},
		},
		{
			name: "Track",
			info: &samvaad.EgressInfo{
				Request: &samvaad.EgressInfo_Track{
					Track: &samvaad.TrackEgressRequest{
						RoomName: "test-room",
					},
				},
			},
		},
		{
			name: "Undefined",
			info: &samvaad.EgressInfo{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := GetRequest(tt.info)
			require.NoError(t, err)
			if tt.info.Request == nil {
				require.Empty(t, result)
			} else {
				require.NotEmpty(t, result)
			}
		})
	}
}

func TestGetResult(t *testing.T) {
	tests := []struct {
		name     string
		info     *samvaad.EgressInfo
		expected string
	}{
		{
			name: "FileResult",
			info: &samvaad.EgressInfo{
				Result: &samvaad.EgressInfo_File{
					File: &samvaad.FileInfo{Filename: "test.mp4", Size: 1024},
				},
			},
			expected: `{"file_results":[{"filename":"test.mp4", "size":1024}]}`,
		},
		{
			name: "StreamResult",
			info: &samvaad.EgressInfo{
				Result: &samvaad.EgressInfo_Stream{
					Stream: &samvaad.StreamInfoList{
						Info: []*samvaad.StreamInfo{{Url: "rtmp://example.com/live"}},
					},
				},
			},
			expected: `{"stream_results":[{"url":"rtmp://example.com/live"}]}`,
		},
		{
			name: "SegmentResult",
			info: &samvaad.EgressInfo{
				Result: &samvaad.EgressInfo_Segments{
					Segments: &samvaad.SegmentsInfo{PlaylistName: "playlist.m3u8"},
				},
			},
			expected: `{"segment_results":[{"playlist_name":"playlist.m3u8"}]}`,
		},
		{
			name: "MultipleResults",
			info: &samvaad.EgressInfo{
				FileResults: []*samvaad.FileInfo{
					{Filename: "test.mp4"},
				},
			},
			expected: `{"file_results":[{"filename":"test.mp4"}]}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := GetResult(tt.info)
			require.NoError(t, err)
			require.JSONEq(t, tt.expected, result)
		})
	}
}

func TestGetStatus(t *testing.T) {
	tests := []struct {
		status   samvaad.EgressStatus
		expected string
	}{
		{samvaad.EgressStatus_EGRESS_STARTING, "starting"},
		{samvaad.EgressStatus_EGRESS_ACTIVE, "active"},
		{samvaad.EgressStatus_EGRESS_ENDING, "ending"},
		{samvaad.EgressStatus_EGRESS_COMPLETE, "complete"},
		{samvaad.EgressStatus_EGRESS_ABORTED, "aborted"},
		{samvaad.EgressStatus_EGRESS_LIMIT_REACHED, "limit_reached"},
		{samvaad.EgressStatus_EGRESS_FAILED, "failed"},
	}

	for _, tt := range tests {
		t.Run(tt.status.String(), func(t *testing.T) {
			info := &samvaad.EgressInfo{Status: tt.status}
			result := GetStatus(info)
			require.Equal(t, tt.expected, string(result))
		})
	}
}
