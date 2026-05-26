package utils

import (
	"testing"

	"github.com/msmclass/samvaad/pkg/proto/samvaad"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestProtoRedact(t *testing.T) {
	msg := &samvaad.ParticipantInfo{
		Identity: "testIdentity",
		Name:     "testName",
		Metadata: "testMetadata",
		Attributes: map[string]string{
			"key1": "value1",
			"key2": "value2",
		},
		Tracks: []*samvaad.TrackInfo{
			{
				Sid:  "TR_AUDIO",
				Name: "Microphone",
			},
			{
				Sid:  "TR_VIDEO",
				Name: "Camera",
			},
		},
	}
	cloned := CloneProtoRedacted(msg)
	expectedCloned := &samvaad.ParticipantInfo{
		Identity: "testIdentity",
		Tracks: []*samvaad.TrackInfo{
			{
				Sid: "TR_AUDIO",
			},
			{
				Sid: "TR_VIDEO",
			},
		},
	}
	require.True(t, proto.Equal(expectedCloned, cloned))
}
