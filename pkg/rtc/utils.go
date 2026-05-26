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

package rtc

import (
	"errors"
	"io"
	"net"
	"strings"

	"github.com/pion/webrtc/v4"
	"google.golang.org/protobuf/proto"

	"github.com/msmclass/samvaad/pkg/samvaad/codecs/mime"
	"github.com/msmclass/samvaad/pkg/proto/samvaad"
	"github.com/msmclass/samvaad/pkg/samvaad/logger"
)

const (
	trackIdSeparator = "|"

	cMinIPTruncateLen = 8
)

func UnpackStreamID(packed string) (participantID samvaad.ParticipantID, trackID samvaad.TrackID) {
	parts := strings.Split(packed, trackIdSeparator)
	if len(parts) > 1 {
		return samvaad.ParticipantID(parts[0]), samvaad.TrackID(packed[len(parts[0])+1:])
	}
	return samvaad.ParticipantID(packed), ""
}

func PackStreamID(participantID samvaad.ParticipantID, trackID samvaad.TrackID) string {
	return string(participantID) + trackIdSeparator + string(trackID)
}

func PackSyncStreamID(participantID samvaad.ParticipantID, stream string) string {
	return string(participantID) + trackIdSeparator + stream
}

func StreamFromTrackSource(source samvaad.TrackSource) string {
	// group camera/mic, screenshare/audio together
	switch source {
	case samvaad.TrackSource_SCREEN_SHARE:
		return "screen"
	case samvaad.TrackSource_SCREEN_SHARE_AUDIO:
		return "screen"
	case samvaad.TrackSource_CAMERA:
		return "camera"
	case samvaad.TrackSource_MICROPHONE:
		return "camera"
	}
	return "unknown"
}

func PackDataTrackLabel(participantID samvaad.ParticipantID, trackID samvaad.TrackID, label string) string {
	return string(participantID) + trackIdSeparator + string(trackID) + trackIdSeparator + label
}

func UnpackDataTrackLabel(packed string) (participantID samvaad.ParticipantID, trackID samvaad.TrackID, label string) {
	parts := strings.Split(packed, trackIdSeparator)
	if len(parts) != 3 {
		return "", samvaad.TrackID(packed), ""
	}
	participantID = samvaad.ParticipantID(parts[0])
	trackID = samvaad.TrackID(parts[1])
	label = parts[2]
	return
}

func ToProtoTrackKind(kind webrtc.RTPCodecType) samvaad.TrackType {
	switch kind {
	case webrtc.RTPCodecTypeVideo:
		return samvaad.TrackType_VIDEO
	case webrtc.RTPCodecTypeAudio:
		return samvaad.TrackType_AUDIO
	}
	panic("unsupported track direction")
}

func IsEOF(err error) bool {
	return err == io.ErrClosedPipe || err == io.EOF
}

func Recover(l logger.Logger) any {
	if l == nil {
		l = logger.GetLogger()
	}
	r := recover()
	if r != nil {
		var err error
		switch e := r.(type) {
		case string:
			err = errors.New(e)
		case error:
			err = e
		default:
			err = errors.New("unknown panic")
		}
		l.Errorw("recovered panic", err, "panic", r)
	}

	return r
}

// logger helpers
func LoggerWithParticipant(l logger.Logger, identity samvaad.ParticipantIdentity, sid samvaad.ParticipantID, isRemote bool) logger.Logger {
	values := make([]any, 0, 4)
	if identity != "" {
		values = append(values, "participant", identity)
	}
	if sid != "" {
		values = append(values, "participantID", sid)
	}
	values = append(values, "remote", isRemote)
	// enable sampling per participant
	return l.WithValues(values...)
}

func LoggerWithRoom(l logger.Logger, name samvaad.RoomName, roomID samvaad.RoomID) logger.Logger {
	values := make([]any, 0, 2)
	if name != "" {
		values = append(values, "room", name)
	}
	if roomID != "" {
		values = append(values, "roomID", roomID)
	}
	// also sample for the room
	return l.WithItemSampler().WithValues(values...)
}

func LoggerWithTrack(l logger.Logger, trackID samvaad.TrackID, isRelayed bool) logger.Logger {
	// sampling not required because caller already passing in participant's logger
	if trackID != "" {
		return l.WithValues("trackID", trackID, "relayed", isRelayed)
	}
	return l
}

func LoggerWithPCTarget(l logger.Logger, target samvaad.SignalTarget) logger.Logger {
	return l.WithValues("transport", target)
}

func LoggerWithCodecMime(l logger.Logger, mimeType mime.MimeType) logger.Logger {
	if mimeType != mime.MimeTypeUnknown {
		return l.WithValues("mime", mimeType.String())
	}
	return l
}

func MaybeTruncateIP(addr string) string {
	ipAddr := net.ParseIP(addr)
	if ipAddr == nil {
		return ""
	}

	if ipAddr.IsPrivate() || len(addr) <= cMinIPTruncateLen {
		return addr
	}

	return addr[:len(addr)-3] + "..."
}

func ChunkProtoBatch[T proto.Message](batch []T, target int) [][]T {
	var chunks [][]T
	var start, size int
	for i, m := range batch {
		s := proto.Size(m)
		if size+s > target {
			if start < i {
				chunks = append(chunks, batch[start:i])
			}
			start = i
			size = 0
		}
		size += s
	}
	if start < len(batch) {
		chunks = append(chunks, batch[start:])
	}
	return chunks
}

func IsRedEnabled(ti *samvaad.TrackInfo) bool {
	if len(ti.Codecs) != 0 && ti.Codecs[0].MimeType != "" {
		return mime.IsMimeTypeStringRED(ti.Codecs[0].MimeType)
	}

	return !ti.GetDisableRed()
}


