package roomobs

import (
	"fmt"
	"strings"
	"unsafe"

	"github.com/msmclass/samvaad/pkg/proto/samvaad"
	"github.com/msmclass/samvaad/pkg/samvaad/utils"
)

const tagDelimiter = "\x1e"

type Tag string

func ToTag(key, value string) Tag {
	return Tag(key + tagDelimiter + value)
}

func (t Tag) KeyValue() (string, string) {
	key, value, ok := strings.Cut(string(t), tagDelimiter)
	if !ok {
		return string(t), ""
	}
	return key, value
}

type Tags []Tag

func (t Tags) Strings() []string {
	return utils.CastStringSlice[string](t)
}

func ToTags(m map[string]string) Tags {
	t := make(Tags, 0, len(m))
	for k, v := range m {
		t = append(t, ToTag(k, v))
	}
	return t
}

func PackTrackLayer(x, y uint32) uint32 {
	return uint32(x<<16 | y)
}

func UnpackTrackLayer(layer uint32) (x, y int) {
	return int(layer >> 16), int(layer & 0xffff)
}

func PackCountryCode(isoAlpha2 string) uint16 {
	if len(isoAlpha2) != 2 {
		return PackCountryCode("??")
	}
	return uint16(isoAlpha2[0])<<8 | uint16(isoAlpha2[1])
}

func UnpackCountryCode(code uint16) (isoAlpha2 string) {
	b := [2]byte{byte(code >> 8), byte(code)}
	return unsafe.String(&b[0], 2)
}

func ToClientOS(os string) ClientOS {
	switch strings.ToLower(os) {
	case "":
		return ClientOSUndefined
	case "ios":
		return ClientOSIos
	case "android":
		return ClientOSAndroid
	case "windows":
		return ClientOSWindows
	case "mac", "mac os x", "darwin", "macos":
		return ClientOSMac
	case "linux", "chrome os":
		return ClientOSLinux
	default:
		return ClientOSUndefined
	}
}

func FormatBrowser(clientInfo *samvaad.ClientInfo) string {
	return strings.TrimSpace(fmt.Sprintf("%s %s", clientInfo.GetBrowser(), clientInfo.GetBrowserVersion()))
}

func FormatSDKVersion(clientInfo *samvaad.ClientInfo) string {
	return strings.TrimSpace(fmt.Sprintf("%s %s", clientInfo.GetSdk(), clientInfo.GetVersion()))
}

func TrackKindFromProto(p samvaad.StreamType) TrackKind {
	switch p {
	case samvaad.StreamType_UPSTREAM:
		return TrackKindSub
	case samvaad.StreamType_DOWNSTREAM:
		return TrackKindPub
	default:
		return TrackKindUndefined
	}
}

func TrackTypeFromProto(p samvaad.TrackType) TrackType {
	switch p {
	case samvaad.TrackType_AUDIO:
		return TrackTypeAudio
	case samvaad.TrackType_VIDEO:
		return TrackTypeVideo
	case samvaad.TrackType_DATA:
		return TrackTypeData
	default:
		return TrackTypeUndefined
	}
}

func TrackSourceFromProto(p samvaad.TrackSource) TrackSource {
	switch p {
	case samvaad.TrackSource_UNKNOWN:
		return TrackSourceUndefined
	case samvaad.TrackSource_CAMERA:
		return TrackSourceCamera
	case samvaad.TrackSource_MICROPHONE:
		return TrackSourceMicrophone
	case samvaad.TrackSource_SCREEN_SHARE:
		return TrackSourceScreenShare
	case samvaad.TrackSource_SCREEN_SHARE_AUDIO:
		return TrackSourceScreenShareAudio
	default:
		return TrackSourceUndefined
	}
}

type RoomFeature uint16

func (f RoomFeature) HasIngress() bool   { return f&IngressRoomFeature != 0 }
func (f RoomFeature) HasEgress() bool    { return f&EgressRoomFeature != 0 }
func (f RoomFeature) HasSIP() bool       { return f&SIPRoomFeature != 0 }
func (f RoomFeature) HasAgent() bool     { return f&AgentRoomFeature != 0 }
func (f RoomFeature) HasConnector() bool { return f&ConnectorRoomFeature != 0 }

const (
	IngressRoomFeature RoomFeature = 1 << iota
	EgressRoomFeature
	SIPRoomFeature
	AgentRoomFeature
	ConnectorRoomFeature
)

func RoomFeatureFromParticipantKind(k samvaad.ParticipantInfo_Kind) RoomFeature {
	switch k {
	case samvaad.ParticipantInfo_INGRESS:
		return IngressRoomFeature
	case samvaad.ParticipantInfo_EGRESS:
		return EgressRoomFeature
	case samvaad.ParticipantInfo_SIP:
		return SIPRoomFeature
	case samvaad.ParticipantInfo_AGENT:
		return AgentRoomFeature
	case samvaad.ParticipantInfo_CONNECTOR:
		return ConnectorRoomFeature
	default:
		return 0
	}
}

func ParticipantKindCode(k samvaad.ParticipantInfo_Kind) int32 {
	return int32(k)
}

func ParticipantKindDetailsCodes(d []samvaad.ParticipantInfo_KindDetail) []int32 {
	return *(*[]int32)(unsafe.Pointer(&d))
}
