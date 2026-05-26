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

package types

import (
	"fmt"
	"time"

	"github.com/pion/rtcp"
	"github.com/pion/webrtc/v4"

	"github.com/msmclass/samvaad/pkg/samvaad/auth"
	"github.com/msmclass/samvaad/pkg/samvaad/codecs/mime"
	"github.com/msmclass/samvaad/pkg/proto/samvaad"
	"github.com/msmclass/samvaad/pkg/samvaad/logger"
	"github.com/msmclass/samvaad/pkg/samvaad/observability/roomobs"
	"github.com/msmclass/samvaad/pkg/samvaad/utils"

	"github.com/msmclass/samvaad/pkg/routing"
	"github.com/msmclass/samvaad/pkg/rtc/datatrack"
	"github.com/msmclass/samvaad/pkg/sfu"
	"github.com/msmclass/samvaad/pkg/sfu/buffer"
	"github.com/msmclass/samvaad/pkg/sfu/pacer"
	"github.com/msmclass/samvaad/pkg/telemetry"

	"google.golang.org/protobuf/proto"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

//counterfeiter:generate . WebsocketClient
type WebsocketClient interface {
	ReadMessage() (messageType int, p []byte, err error)
	WriteMessage(messageType int, data []byte) error
	WriteControl(messageType int, data []byte, deadline time.Time) error
	SetReadDeadline(deadline time.Time) error
	Close() error
}

type AddSubscriberParams struct {
	AllTracks bool
	TrackIDs  []samvaad.TrackID
}

// ---------------------------------------------

type MigrateState int32

const (
	MigrateStateInit MigrateState = iota
	MigrateStateSync
	MigrateStateComplete
)

func (m MigrateState) String() string {
	switch m {
	case MigrateStateInit:
		return "MIGRATE_STATE_INIT"
	case MigrateStateSync:
		return "MIGRATE_STATE_SYNC"
	case MigrateStateComplete:
		return "MIGRATE_STATE_COMPLETE"
	default:
		return fmt.Sprintf("%d", int(m))
	}
}

// ---------------------------------------------

type SubscribedCodecQuality struct {
	CodecMime mime.MimeType
	Quality   samvaad.VideoQuality
}

// ---------------------------------------------

type ParticipantCloseReason int

const (
	ParticipantCloseReasonNone ParticipantCloseReason = iota
	ParticipantCloseReasonClientRequestLeave
	ParticipantCloseReasonRoomManagerStop
	ParticipantCloseReasonVerifyFailed
	ParticipantCloseReasonJoinFailed
	ParticipantCloseReasonJoinTimeout
	ParticipantCloseReasonMessageBusFailed
	ParticipantCloseReasonPeerConnectionDisconnected
	ParticipantCloseReasonDuplicateIdentity
	ParticipantCloseReasonMigrationComplete
	ParticipantCloseReasonStale
	ParticipantCloseReasonServiceRequestRemoveParticipant
	ParticipantCloseReasonServiceRequestDeleteRoom
	ParticipantCloseReasonSimulateMigration
	ParticipantCloseReasonSimulateNodeFailure
	ParticipantCloseReasonSimulateServerLeave
	ParticipantCloseReasonSimulateLeaveRequest
	ParticipantCloseReasonNegotiateFailed
	ParticipantCloseReasonMigrationRequested
	ParticipantCloseReasonPublicationError
	ParticipantCloseReasonSubscriptionError
	ParticipantCloseReasonDataChannelError
	ParticipantCloseReasonMigrateCodecMismatch
	ParticipantCloseReasonSignalSourceClose
	ParticipantCloseReasonRoomClosed
	ParticipantCloseReasonUserUnavailable
	ParticipantCloseReasonUserRejected
	ParticipantCloseReasonMoveFailed
	ParticipantCloseReasonAgentError
)

func (p ParticipantCloseReason) String() string {
	switch p {
	case ParticipantCloseReasonNone:
		return "NONE"
	case ParticipantCloseReasonClientRequestLeave:
		return "CLIENT_REQUEST_LEAVE"
	case ParticipantCloseReasonRoomManagerStop:
		return "ROOM_MANAGER_STOP"
	case ParticipantCloseReasonVerifyFailed:
		return "VERIFY_FAILED"
	case ParticipantCloseReasonJoinFailed:
		return "JOIN_FAILED"
	case ParticipantCloseReasonJoinTimeout:
		return "JOIN_TIMEOUT"
	case ParticipantCloseReasonMessageBusFailed:
		return "MESSAGE_BUS_FAILED"
	case ParticipantCloseReasonPeerConnectionDisconnected:
		return "PEER_CONNECTION_DISCONNECTED"
	case ParticipantCloseReasonDuplicateIdentity:
		return "DUPLICATE_IDENTITY"
	case ParticipantCloseReasonMigrationComplete:
		return "MIGRATION_COMPLETE"
	case ParticipantCloseReasonStale:
		return "STALE"
	case ParticipantCloseReasonServiceRequestRemoveParticipant:
		return "SERVICE_REQUEST_REMOVE_PARTICIPANT"
	case ParticipantCloseReasonServiceRequestDeleteRoom:
		return "SERVICE_REQUEST_DELETE_ROOM"
	case ParticipantCloseReasonSimulateMigration:
		return "SIMULATE_MIGRATION"
	case ParticipantCloseReasonSimulateNodeFailure:
		return "SIMULATE_NODE_FAILURE"
	case ParticipantCloseReasonSimulateServerLeave:
		return "SIMULATE_SERVER_LEAVE"
	case ParticipantCloseReasonSimulateLeaveRequest:
		return "SIMULATE_LEAVE_REQUEST"
	case ParticipantCloseReasonNegotiateFailed:
		return "NEGOTIATE_FAILED"
	case ParticipantCloseReasonMigrationRequested:
		return "MIGRATION_REQUESTED"
	case ParticipantCloseReasonPublicationError:
		return "PUBLICATION_ERROR"
	case ParticipantCloseReasonSubscriptionError:
		return "SUBSCRIPTION_ERROR"
	case ParticipantCloseReasonDataChannelError:
		return "DATA_CHANNEL_ERROR"
	case ParticipantCloseReasonMigrateCodecMismatch:
		return "MIGRATE_CODEC_MISMATCH"
	case ParticipantCloseReasonSignalSourceClose:
		return "SIGNAL_SOURCE_CLOSE"
	case ParticipantCloseReasonRoomClosed:
		return "ROOM_CLOSED"
	case ParticipantCloseReasonUserUnavailable:
		return "USER_UNAVAILABLE"
	case ParticipantCloseReasonUserRejected:
		return "USER_REJECTED"
	case ParticipantCloseReasonMoveFailed:
		return "MOVE_FAILED"
	case ParticipantCloseReasonAgentError:
		return "AGENT_ERROR"
	default:
		return fmt.Sprintf("%d", int(p))
	}
}

func (p ParticipantCloseReason) ToDisconnectReason() samvaad.DisconnectReason {
	switch p {
	case ParticipantCloseReasonClientRequestLeave, ParticipantCloseReasonSimulateLeaveRequest:
		return samvaad.DisconnectReason_CLIENT_INITIATED
	case ParticipantCloseReasonRoomManagerStop:
		return samvaad.DisconnectReason_SERVER_SHUTDOWN
	case ParticipantCloseReasonVerifyFailed, ParticipantCloseReasonJoinFailed, ParticipantCloseReasonJoinTimeout, ParticipantCloseReasonMessageBusFailed:
		// expected to be connected but is not
		return samvaad.DisconnectReason_JOIN_FAILURE
	case ParticipantCloseReasonPeerConnectionDisconnected:
		return samvaad.DisconnectReason_CONNECTION_TIMEOUT
	case ParticipantCloseReasonDuplicateIdentity, ParticipantCloseReasonStale:
		return samvaad.DisconnectReason_DUPLICATE_IDENTITY
	case ParticipantCloseReasonMigrationRequested, ParticipantCloseReasonMigrationComplete, ParticipantCloseReasonSimulateMigration:
		return samvaad.DisconnectReason_MIGRATION
	case ParticipantCloseReasonServiceRequestRemoveParticipant:
		return samvaad.DisconnectReason_PARTICIPANT_REMOVED
	case ParticipantCloseReasonServiceRequestDeleteRoom:
		return samvaad.DisconnectReason_ROOM_DELETED
	case ParticipantCloseReasonSimulateNodeFailure, ParticipantCloseReasonSimulateServerLeave:
		return samvaad.DisconnectReason_SERVER_SHUTDOWN
	case ParticipantCloseReasonNegotiateFailed, ParticipantCloseReasonPublicationError, ParticipantCloseReasonSubscriptionError,
		ParticipantCloseReasonDataChannelError, ParticipantCloseReasonMigrateCodecMismatch, ParticipantCloseReasonMoveFailed:
		return samvaad.DisconnectReason_STATE_MISMATCH
	case ParticipantCloseReasonSignalSourceClose:
		return samvaad.DisconnectReason_SIGNAL_CLOSE
	case ParticipantCloseReasonRoomClosed:
		return samvaad.DisconnectReason_ROOM_CLOSED
	case ParticipantCloseReasonUserUnavailable:
		return samvaad.DisconnectReason_USER_UNAVAILABLE
	case ParticipantCloseReasonUserRejected:
		return samvaad.DisconnectReason_USER_REJECTED
	case ParticipantCloseReasonAgentError:
		return samvaad.DisconnectReason_AGENT_ERROR
	default:
		// the other types will map to unknown reason
		return samvaad.DisconnectReason_UNKNOWN_REASON
	}
}

// IsIntentionalDisconnect reports whether a disconnect reason represents an
// intentional/expected closure (client leaving, admin action, room teardown,
// migration, etc.) as opposed to a connection failure.
func IsIntentionalDisconnect(reason samvaad.DisconnectReason) bool {
	switch reason {
	case samvaad.DisconnectReason_CLIENT_INITIATED,
		samvaad.DisconnectReason_SERVER_SHUTDOWN,
		samvaad.DisconnectReason_DUPLICATE_IDENTITY,
		samvaad.DisconnectReason_MIGRATION,
		samvaad.DisconnectReason_PARTICIPANT_REMOVED,
		samvaad.DisconnectReason_ROOM_DELETED,
		samvaad.DisconnectReason_ROOM_CLOSED:
		return true
	}
	return false
}

// ---------------------------------------------

type SignallingCloseReason int

const (
	SignallingCloseReasonUnknown SignallingCloseReason = iota
	SignallingCloseReasonMigration
	SignallingCloseReasonResume
	SignallingCloseReasonTransportFailure
	SignallingCloseReasonFullReconnectPublicationError
	SignallingCloseReasonFullReconnectSubscriptionError
	SignallingCloseReasonFullReconnectDataChannelError
	SignallingCloseReasonFullReconnectNegotiateFailed
	SignallingCloseReasonParticipantClose
	SignallingCloseReasonDisconnectOnResume
	SignallingCloseReasonDisconnectOnResumeNoMessages
)

func (s SignallingCloseReason) String() string {
	switch s {
	case SignallingCloseReasonUnknown:
		return "UNKNOWN"
	case SignallingCloseReasonMigration:
		return "MIGRATION"
	case SignallingCloseReasonResume:
		return "RESUME"
	case SignallingCloseReasonTransportFailure:
		return "TRANSPORT_FAILURE"
	case SignallingCloseReasonFullReconnectPublicationError:
		return "FULL_RECONNECT_PUBLICATION_ERROR"
	case SignallingCloseReasonFullReconnectSubscriptionError:
		return "FULL_RECONNECT_SUBSCRIPTION_ERROR"
	case SignallingCloseReasonFullReconnectDataChannelError:
		return "FULL_RECONNECT_DATA_CHANNEL_ERROR"
	case SignallingCloseReasonFullReconnectNegotiateFailed:
		return "FULL_RECONNECT_NEGOTIATE_FAILED"
	case SignallingCloseReasonParticipantClose:
		return "PARTICIPANT_CLOSE"
	case SignallingCloseReasonDisconnectOnResume:
		return "DISCONNECT_ON_RESUME"
	case SignallingCloseReasonDisconnectOnResumeNoMessages:
		return "DISCONNECT_ON_RESUME_NO_MESSAGES"
	default:
		return fmt.Sprintf("%d", int(s))
	}
}

// ---------------------------------------------
const (
	ParticipantCloseKeyNormal = "normal"
	ParticipantCloseKeyWHIP   = "whip"
)

// ---------------------------------------------

//counterfeiter:generate . Participant
type Participant interface {
	ID() samvaad.ParticipantID
	Identity() samvaad.ParticipantIdentity
	State() samvaad.ParticipantInfo_State
	ConnectedAt() time.Time
	CloseReason() ParticipantCloseReason
	Kind() samvaad.ParticipantInfo_Kind
	KindDetails() []samvaad.ParticipantInfo_KindDetail
	IsRecorder() bool
	IsDependent() bool
	IsAgent() bool

	GetLogger() logger.Logger

	CanSkipBroadcast() bool
	Version() utils.TimedVersion
	ToProto() *samvaad.ParticipantInfo
	ToProtoWithVersion() (*samvaad.ParticipantInfo, utils.TimedVersion)

	IsPublisher() bool
	GetPublishedTrack(trackID samvaad.TrackID) MediaTrack
	GetPublishedTracks() []MediaTrack
	RemovePublishedTrack(track MediaTrack, isExpectedToResume bool)

	GetPublishedDataTracks() []DataTrack
	GetPublishedDataTrack(handle uint16) DataTrack
	RemovePublishedDataTrack(track DataTrack)

	GetAudioLevel() (smoothedLevel float64, active bool)

	// HasPermission checks permission of the subscriber by identity. Returns true if subscriber is allowed to subscribe
	// to the track with trackID
	HasPermission(trackID samvaad.TrackID, subIdentity samvaad.ParticipantIdentity) bool

	// permissions
	Hidden() bool

	MigrateState() MigrateState

	Close(sendLeave bool, reason ParticipantCloseReason, isExpectedToResume bool) error
	IsClosed() bool
	IsDisconnected() bool

	SubscriptionPermission() (*samvaad.SubscriptionPermission, utils.TimedVersion)

	// updates from remotes
	UpdateSubscriptionPermission(
		subscriptionPermission *samvaad.SubscriptionPermission,
		timedVersion utils.TimedVersion,
		resolverBySid func(participantID samvaad.ParticipantID) LocalParticipant,
	) error

	DebugInfo() map[string]any

	HandleReceivedDataTrackMessage([]byte, *datatrack.Packet, int64)

	GetParticipantListener() ParticipantListener
}

// -------------------------------------------------------

type AddTrackParams struct {
	Stereo bool
	Red    bool
}

type MoveToRoomParams struct {
	RoomName      samvaad.RoomName
	ParticipantID samvaad.ParticipantID
	Listener      LocalParticipantListener
	Helper        LocalParticipantHelper
}

type DataMessageCache struct {
	Data           []byte
	SenderID       samvaad.ParticipantID
	Seq            uint32
	DestIdentities []samvaad.ParticipantIdentity
}

//counterfeiter:generate . LocalParticipantHelper
type LocalParticipantHelper interface {
	ResolveMediaTrack(LocalParticipant, samvaad.TrackID) MediaResolverResult
	ResolveDataTrack(LocalParticipant, samvaad.TrackID) DataResolverResult
	GetParticipantInfo(pID samvaad.ParticipantID) *samvaad.ParticipantInfo
	GetRegionSettings(ip string) *samvaad.RegionSettings
	GetSubscriberForwarderState(p LocalParticipant) (map[samvaad.TrackID]*samvaad.RTPForwarderState, error)
	ShouldRegressCodec() bool
	GetCachedReliableDataMessage(seqs map[samvaad.ParticipantID]uint32) []*DataMessageCache
}

//counterfeiter:generate . LocalParticipant
type LocalParticipant interface {
	Participant

	TelemetryGuard() *telemetry.ReferenceGuard
	GetTelemetryListener() ParticipantTelemetryListener

	// getters
	GetCountry() string
	GetTrailer() []byte
	GetLoggerResolver() logger.DeferredFieldResolver
	GetReporter() roomobs.ParticipantSessionReporter
	GetReporterResolver() roomobs.ParticipantReporterResolver
	GetAdaptiveStream() bool
	ProtocolVersion() ProtocolVersion
	SupportsSyncStreamID() bool
	SupportsTransceiverReuse(mt MediaTrack) bool
	IsUsingSinglePeerConnection() bool
	IsReady() bool
	ActiveAt() time.Time
	Disconnected() <-chan struct{}
	IsIdle() bool
	SubscriberAsPrimary() bool
	GetClientInfo() *samvaad.ClientInfo
	GetClientConfiguration() *samvaad.ClientConfiguration
	GetBufferFactory() *buffer.Factory
	GetPlayoutDelayConfig() *samvaad.PlayoutDelay
	GetPendingTrack(trackID samvaad.TrackID) *samvaad.TrackInfo
	GetICEConnectionInfo() []*ICEConnectionInfo
	HasConnected() bool
	GetEnabledPublishCodecs() []*samvaad.Codec
	GetPublisherICESessionUfrag() (string, error)
	SupportsMoving() error
	GetLastReliableSequence(migrateOut bool) uint32

	SwapResponseSink(sink routing.MessageSink, reason SignallingCloseReason)
	GetResponseSink() routing.MessageSink
	CloseSignalConnection(reason SignallingCloseReason)
	UpdateLastSeenSignal()
	SetSignalSourceValid(valid bool)
	HandleSignalSourceClose()

	// updates
	UpdateMetadata(update *samvaad.UpdateParticipantMetadata, fromAdmin bool) error
	SetName(name string)
	SetMetadata(metadata string)
	SetAttributes(attributes map[string]string)
	UpdateAudioTrack(update *samvaad.UpdateLocalAudioTrack) error
	UpdateVideoTrack(update *samvaad.UpdateLocalVideoTrack) error

	// permissions
	ClaimGrants() *auth.ClaimGrants
	SetPermission(permission *samvaad.ParticipantPermission) bool
	CanPublish() bool
	CanPublishSource(source samvaad.TrackSource) bool
	CanSubscribe() bool
	CanPublishData() bool

	// PeerConnection
	HandleICETrickle(trickleRequest *samvaad.TrickleRequest)
	HandleOffer(sd *samvaad.SessionDescription) error
	GetAnswer() (webrtc.SessionDescription, uint32, error)
	HandleICETrickleSDPFragment(sdpFragment string) error
	HandleICERestartSDPFragment(sdpFragment string) (string, error)
	AddTrack(req *samvaad.AddTrackRequest)
	SetTrackMuted(mute *samvaad.MuteTrackRequest, fromAdmin bool) *samvaad.TrackInfo

	HandleAnswer(sd *samvaad.SessionDescription)
	Negotiate(force bool)
	ICERestart(iceConfig *samvaad.ICEConfig)
	AddTrackLocal(trackLocal webrtc.TrackLocal, params AddTrackParams) (*webrtc.RTPSender, *webrtc.RTPTransceiver, error)
	AddTransceiverFromTrackLocal(trackLocal webrtc.TrackLocal, params AddTrackParams) (*webrtc.RTPSender, *webrtc.RTPTransceiver, error)
	RemoveTrackLocal(sender *webrtc.RTPSender) error

	WriteSubscriberRTCP(pkts []rtcp.Packet) error

	// subscriptions
	SubscribeToTrack(trackID samvaad.TrackID, isSync bool)
	UnsubscribeFromTrack(trackID samvaad.TrackID)
	UpdateSubscribedTrackSettings(trackID samvaad.TrackID, settings *samvaad.UpdateTrackSettings)
	GetSubscribedTracks() []SubscribedTrack
	IsTrackNameSubscribed(publisherIdentity samvaad.ParticipantIdentity, trackName string) bool
	SubscribeToDataTrack(trackID samvaad.TrackID)
	UnsubscribeFromDataTrack(trackID samvaad.TrackID)
	UpdateDataTrackSubscriptionOptions(trackID samvaad.TrackID, subscriptionOptions *samvaad.DataTrackSubscriptionOptions)
	Verify() bool
	VerifySubscribeParticipantInfo(pID samvaad.ParticipantID, version uint32)
	// WaitUntilSubscribed waits until all subscriptions have been settled, or if the timeout
	// has been reached. If the timeout expires, it will return an error.
	WaitUntilSubscribed(timeout time.Duration) error
	StopAndGetSubscribedTracksForwarderState() map[samvaad.TrackID]*samvaad.RTPForwarderState
	SupportsCodecChange() bool

	// returns list of participant identities that the current participant is subscribed to
	GetSubscribedParticipants() []samvaad.ParticipantID
	IsSubscribedTo(sid samvaad.ParticipantID) bool

	GetConnectionQuality() *samvaad.ConnectionQualityInfo

	// server sent messages
	SendJoinResponse(joinResponse *samvaad.JoinResponse) error
	SendParticipantUpdate(participants []*samvaad.ParticipantInfo) error
	SendSpeakerUpdate(speakers []*samvaad.SpeakerInfo, force bool) error
	SendDataMessage(kind samvaad.DataPacket_Kind, data []byte, senderID samvaad.ParticipantID, seq uint32) error
	SendDataMessageUnlabeled(data []byte, useRaw bool, sender samvaad.ParticipantIdentity) error
	SendRoomUpdate(room *samvaad.Room) error
	SendConnectionQualityUpdate(update *samvaad.ConnectionQualityUpdate) error
	SendSubscriptionPermissionUpdate(publisherID samvaad.ParticipantID, trackID samvaad.TrackID, allowed bool) error
	SendRefreshToken(token string) error
	HandleReconnectAndSendResponse(reconnectReason samvaad.ReconnectReason, reconnectResponse *samvaad.ReconnectResponse) error
	IssueFullReconnect(reason ParticipantCloseReason)
	SendRoomMovedResponse(moved *samvaad.RoomMovedResponse) error
	SendDataTrackSubscriberHandles(handles map[uint32]*samvaad.DataTrackSubscriberHandles_PublishedDataTrack) error

	AddOnClose(key string, callback func(LocalParticipant))
	OnClaimsChanged(callback func(LocalParticipant))

	HandleReceiverReport(dt *sfu.DownTrack, report *rtcp.ReceiverReport)

	// session migration
	MaybeStartMigration(force bool, onStart func()) bool
	NotifyMigration()
	SetMigrateState(s MigrateState)
	SetMigrateInfo(
		previousOffer *webrtc.SessionDescription,
		previousAnswer *webrtc.SessionDescription,
		mediaTracks []*samvaad.TrackPublishedResponse,
		dataChannels []*samvaad.DataChannelInfo,
		dataChannelReceiveState []*samvaad.DataChannelReceiveState,
		dataTracks []*samvaad.PublishDataTrackResponse,
	)
	IsReconnect() bool
	MoveToRoom(params MoveToRoomParams)

	UpdateMediaRTT(rtt uint32)
	UpdateSignalingRTT(rtt uint32)

	CacheDownTrack(trackID samvaad.TrackID, rtpTransceiver *webrtc.RTPTransceiver, downTrackState sfu.DownTrackState)
	UncacheDownTrack(rtpTransceiver *webrtc.RTPTransceiver)
	GetCachedDownTrack(trackID samvaad.TrackID) (*webrtc.RTPTransceiver, sfu.DownTrackState)

	SetICEConfig(iceConfig *samvaad.ICEConfig)
	GetICEConfig() *samvaad.ICEConfig
	OnICEConfigChanged(callback func(participant LocalParticipant, iceConfig *samvaad.ICEConfig))

	UpdateSubscribedQuality(nodeID samvaad.NodeID, trackID samvaad.TrackID, maxQualities []SubscribedCodecQuality) error
	UpdateSubscribedAudioCodecs(nodeID samvaad.NodeID, trackID samvaad.TrackID, codecs []*samvaad.SubscribedAudioCodec) error
	UpdateMediaLoss(nodeID samvaad.NodeID, trackID samvaad.TrackID, fractionalLoss uint32) error

	// down stream bandwidth management
	SetSubscriberAllowPause(allowPause bool)
	SetSubscriberChannelCapacity(channelCapacity int64)

	GetPacer() pacer.Pacer

	GetDisableSenderReportPassThrough() bool

	HandleMetrics(senderParticipantID samvaad.ParticipantID, batch *samvaad.MetricsBatch) error
	HandleUpdateSubscriptions(
		[]samvaad.TrackID,
		[]*samvaad.ParticipantTracks,
		bool,
	)
	HandleUpdateSubscriptionPermission(*samvaad.SubscriptionPermission) error
	HandleSyncState(*samvaad.SyncState) error
	HandleSimulateScenario(*samvaad.SimulateScenario) error
	HandleLeaveRequest(reason ParticipantCloseReason)

	HandlePublishDataTrackRequest(*samvaad.PublishDataTrackRequest)
	HandleUnpublishDataTrackRequest(*samvaad.UnpublishDataTrackRequest)
	HandleUpdateDataSubscription(*samvaad.UpdateDataSubscription)

	HandleSignalMessage(msg proto.Message) error

	PerformRpc(req *samvaad.PerformRpcRequest, resultCh chan string, errorCh chan error)

	GetDataTrackTransport() DataTrackTransport

	ClearParticipantListener()

	GetNextSubscribedDataTrackHandle() uint16
}

// ---------------------------------------------

//counterfeiter:generate . ParticipantListener
type ParticipantListener interface {
	OnParticipantUpdate(Participant)
	OnTrackPublished(Participant, MediaTrack)
	OnTrackUpdated(Participant, MediaTrack)
	OnTrackUnpublished(Participant, MediaTrack)
	OnDataTrackPublished(Participant, DataTrack)
	OnDataTrackUnpublished(Participant, DataTrack)
	OnDataTrackMessage(Participant, []byte, *datatrack.Packet)
	OnMetrics(Participant, *samvaad.DataPacket)
}

var _ ParticipantListener = (*NullParticipantListener)(nil)

type NullParticipantListener struct{}

func (*NullParticipantListener) OnParticipantUpdate(Participant)                           {}
func (*NullParticipantListener) OnTrackPublished(Participant, MediaTrack)                  {}
func (*NullParticipantListener) OnTrackUpdated(Participant, MediaTrack)                    {}
func (*NullParticipantListener) OnTrackUnpublished(Participant, MediaTrack)                {}
func (*NullParticipantListener) OnDataTrackPublished(Participant, DataTrack)               {}
func (*NullParticipantListener) OnDataTrackUnpublished(Participant, DataTrack)             {}
func (*NullParticipantListener) OnDataTrackMessage(Participant, []byte, *datatrack.Packet) {}
func (*NullParticipantListener) OnMetrics(Participant, *samvaad.DataPacket)                {}

// ---------------------------------------------

//counterfeiter:generate . LocalParticipantListener
type LocalParticipantListener interface {
	ParticipantListener

	OnStateChange(LocalParticipant)
	OnSubscriberReady(LocalParticipant)
	OnMigrateStateChange(LocalParticipant, MigrateState)
	OnDataMessage(LocalParticipant, samvaad.DataPacket_Kind, *samvaad.DataPacket)
	OnDataMessageUnlabeled(LocalParticipant, []byte)
	OnSubscribeStatusChanged(LocalParticipant, samvaad.ParticipantID, bool)
	OnUpdateSubscriptions(
		LocalParticipant,
		[]samvaad.TrackID,
		[]*samvaad.ParticipantTracks,
		bool,
	)
	OnUpdateSubscriptionPermission(LocalParticipant, *samvaad.SubscriptionPermission) error
	OnUpdateDataSubscriptions(LocalParticipant, *samvaad.UpdateDataSubscription)
	OnSyncState(LocalParticipant, *samvaad.SyncState) error
	OnSimulateScenario(LocalParticipant, *samvaad.SimulateScenario) error
	OnLeave(LocalParticipant, ParticipantCloseReason)
}

var _ LocalParticipantListener = (*NullLocalParticipantListener)(nil)

type NullLocalParticipantListener struct {
	NullParticipantListener
}

func (*NullLocalParticipantListener) OnStateChange(LocalParticipant)                      {}
func (*NullLocalParticipantListener) OnSubscriberReady(LocalParticipant)                  {}
func (*NullLocalParticipantListener) OnMigrateStateChange(LocalParticipant, MigrateState) {}
func (*NullLocalParticipantListener) OnDataMessage(LocalParticipant, samvaad.DataPacket_Kind, *samvaad.DataPacket) {
}
func (*NullLocalParticipantListener) OnDataMessageUnlabeled(LocalParticipant, []byte) {}
func (*NullLocalParticipantListener) OnSubscribeStatusChanged(LocalParticipant, samvaad.ParticipantID, bool) {
}
func (*NullLocalParticipantListener) OnUpdateSubscriptions(
	LocalParticipant,
	[]samvaad.TrackID,
	[]*samvaad.ParticipantTracks,
	bool,
) {
}
func (*NullLocalParticipantListener) OnUpdateSubscriptionPermission(LocalParticipant, *samvaad.SubscriptionPermission) error {
	return nil
}
func (*NullLocalParticipantListener) OnUpdateDataSubscriptions(LocalParticipant, *samvaad.UpdateDataSubscription) {
}
func (*NullLocalParticipantListener) OnSyncState(LocalParticipant, *samvaad.SyncState) error {
	return nil
}
func (*NullLocalParticipantListener) OnSimulateScenario(LocalParticipant, *samvaad.SimulateScenario) error {
	return nil
}
func (*NullLocalParticipantListener) OnLeave(LocalParticipant, ParticipantCloseReason) {}

// ---------------------------------------------

//counterfeiter:generate . ParticipantTelemetryListener
type ParticipantTelemetryListener interface {
	OnTrackPublishRequested(pID samvaad.ParticipantID, identity samvaad.ParticipantIdentity, ti *samvaad.TrackInfo)
	OnTrackPublished(pID samvaad.ParticipantID, identity samvaad.ParticipantIdentity, ti *samvaad.TrackInfo, shouldSendEvent bool)
	OnTrackUnpublished(pID samvaad.ParticipantID, identity samvaad.ParticipantIdentity, ti *samvaad.TrackInfo, shouldSendEvent bool)
	OnTrackSubscribeRequested(pID samvaad.ParticipantID, ti *samvaad.TrackInfo)
	OnTrackSubscribed(pID samvaad.ParticipantID, ti *samvaad.TrackInfo, publisherInfo *samvaad.ParticipantInfo, shouldSendEvent bool)
	OnTrackUnsubscribed(pID samvaad.ParticipantID, ti *samvaad.TrackInfo, shouldSendEvent bool)
	OnTrackSubscribeFailed(pID samvaad.ParticipantID, trackID samvaad.TrackID, err error, isUserError bool)
	OnTrackSubscribeStreamStarted(pID samvaad.ParticipantID, ti *samvaad.TrackInfo)
	OnTrackMuted(pID samvaad.ParticipantID, ti *samvaad.TrackInfo)
	OnTrackUnmuted(pID samvaad.ParticipantID, ti *samvaad.TrackInfo)
	OnTrackPublishedUpdate(pID samvaad.ParticipantID, ti *samvaad.TrackInfo)
	OnTrackMaxSubscribedVideoQuality(pID samvaad.ParticipantID, ti *samvaad.TrackInfo, mime mime.MimeType, maxQuality samvaad.VideoQuality)
	OnTrackPublishRTPStats(pID samvaad.ParticipantID, trackID samvaad.TrackID, mimeType mime.MimeType, layer int, stats *samvaad.RTPStats)
	OnTrackSubscribeRTPStats(pID samvaad.ParticipantID, trackID samvaad.TrackID, mimeType mime.MimeType, stats *samvaad.RTPStats)

	OnTrackStats(key telemetry.StatsKey, stat *samvaad.AnalyticsStat)
}

var _ ParticipantTelemetryListener = (*NullParticipantTelemetryListener)(nil)

type NullParticipantTelemetryListener struct{}

func (NullParticipantTelemetryListener) OnTrackPublishRequested(pID samvaad.ParticipantID, identity samvaad.ParticipantIdentity, ti *samvaad.TrackInfo) {
}
func (NullParticipantTelemetryListener) OnTrackPublished(pID samvaad.ParticipantID, identity samvaad.ParticipantIdentity, ti *samvaad.TrackInfo, shouldSendEvent bool) {
}
func (NullParticipantTelemetryListener) OnTrackUnpublished(pID samvaad.ParticipantID, identity samvaad.ParticipantIdentity, ti *samvaad.TrackInfo, shouldSendEvent bool) {
}
func (NullParticipantTelemetryListener) OnTrackSubscribeRequested(pID samvaad.ParticipantID, ti *samvaad.TrackInfo) {
}
func (NullParticipantTelemetryListener) OnTrackSubscribed(pID samvaad.ParticipantID, ti *samvaad.TrackInfo, publisherInfo *samvaad.ParticipantInfo, shouldSendEvent bool) {
}
func (NullParticipantTelemetryListener) OnTrackUnsubscribed(pID samvaad.ParticipantID, ti *samvaad.TrackInfo, shouldSendEvent bool) {
}
func (NullParticipantTelemetryListener) OnTrackSubscribeFailed(pID samvaad.ParticipantID, trackID samvaad.TrackID, err error, isUserError bool) {
}
func (NullParticipantTelemetryListener) OnTrackSubscribeStreamStarted(pID samvaad.ParticipantID, ti *samvaad.TrackInfo) {
}
func (NullParticipantTelemetryListener) OnTrackMuted(pID samvaad.ParticipantID, ti *samvaad.TrackInfo) {
}
func (NullParticipantTelemetryListener) OnTrackUnmuted(pID samvaad.ParticipantID, ti *samvaad.TrackInfo) {
}
func (NullParticipantTelemetryListener) OnTrackPublishedUpdate(pID samvaad.ParticipantID, ti *samvaad.TrackInfo) {
}
func (NullParticipantTelemetryListener) OnTrackMaxSubscribedVideoQuality(pID samvaad.ParticipantID, ti *samvaad.TrackInfo, mime mime.MimeType, maxQuality samvaad.VideoQuality) {
}
func (NullParticipantTelemetryListener) OnTrackPublishRTPStats(pID samvaad.ParticipantID, trackID samvaad.TrackID, mimeType mime.MimeType, layer int, stats *samvaad.RTPStats) {
}
func (NullParticipantTelemetryListener) OnTrackSubscribeRTPStats(pID samvaad.ParticipantID, trackID samvaad.TrackID, mimeType mime.MimeType, stats *samvaad.RTPStats) {
}

func (NullParticipantTelemetryListener) OnTrackStats(key telemetry.StatsKey, stat *samvaad.AnalyticsStat) {
}

// ---------------------------------------------

// Room is a container of participants, and can provide room-level actions
//
//counterfeiter:generate . Room
type Room interface {
	Name() samvaad.RoomName
	ID() samvaad.RoomID
	RemoveParticipant(identity samvaad.ParticipantIdentity, pID samvaad.ParticipantID, reason ParticipantCloseReason)
	UpdateSubscriptions(
		participant LocalParticipant,
		trackIDs []samvaad.TrackID,
		participantTracks []*samvaad.ParticipantTracks,
		subscribe bool,
	)
	ResolveMediaTrackForSubscriber(sub LocalParticipant, trackID samvaad.TrackID) MediaResolverResult
	ResolveDataTrackForSubscriber(sub LocalParticipant, trackID samvaad.TrackID) DataResolverResult
	GetLocalParticipants() []LocalParticipant
	IsDataMessageUserPacketDuplicate(ip *samvaad.UserPacket) bool
}

// MediaTrack represents a media track
//
//counterfeiter:generate . MediaTrack
type MediaTrack interface {
	ID() samvaad.TrackID
	Kind() samvaad.TrackType
	Name() string
	Source() samvaad.TrackSource
	Stream() string

	UpdateTrackInfo(ti *samvaad.TrackInfo)
	UpdateAudioTrack(update *samvaad.UpdateLocalAudioTrack)
	UpdateVideoTrack(update *samvaad.UpdateLocalVideoTrack)
	ToProto() *samvaad.TrackInfo

	PublisherID() samvaad.ParticipantID
	PublisherIdentity() samvaad.ParticipantIdentity
	PublisherVersion() uint32
	Logger() logger.Logger

	IsMuted() bool
	SetMuted(muted bool)

	GetAudioLevel() (level float64, active bool)

	Close(isExpectedToResume bool)
	IsOpen() bool

	// callbacks
	AddOnClose(func(isExpectedToResume bool))

	// subscribers
	AddSubscriber(participant LocalParticipant) (SubscribedTrack, error)
	RemoveSubscriber(participantID samvaad.ParticipantID, isExpectedToResume bool)
	IsSubscriber(subID samvaad.ParticipantID) bool
	RevokeDisallowedSubscribers(allowedSubscriberIdentities []samvaad.ParticipantIdentity) []samvaad.ParticipantIdentity
	GetAllSubscribers() []samvaad.ParticipantID
	GetNumSubscribers() int
	OnTrackSubscribed()

	// returns quality information that's appropriate for width & height
	GetQualityForDimension(mimeType mime.MimeType, width, height uint32) samvaad.VideoQuality

	// returns temporal layer that's appropriate for fps
	GetTemporalLayerForSpatialFps(mimeType mime.MimeType, spatial int32, fps uint32) int32

	Receivers() []sfu.TrackReceiver
	ClearAllReceivers(isExpectedToResume bool)

	IsEncrypted() bool
	HasPacketTrailer() bool
}

//counterfeiter:generate . LocalMediaTrack
type LocalMediaTrack interface {
	MediaTrack

	Restart()

	HasSignalCid(cid string) bool
	HasSdpCid(cid string) bool

	GetConnectionScoreAndQuality() (float32, samvaad.ConnectionQuality)
	GetTrackStats() *samvaad.RTPStats

	SetRTT(rtt uint32)

	NotifySubscriberNodeMaxQuality(nodeID samvaad.NodeID, qualities []SubscribedCodecQuality)
	NotifySubscriptionNode(nodeID samvaad.NodeID, codecs []*samvaad.SubscribedAudioCodec)
	ClearSubscriberNodes()
	NotifySubscriberNodeMediaLoss(nodeID samvaad.NodeID, fractionalLoss uint8)
}

// DataTrack represents a data track
//
//counterfeiter:generate . DataTrack
type DataTrack interface {
	ID() samvaad.TrackID
	PubHandle() uint16
	Name() string
	ToProto() *samvaad.DataTrackInfo

	PublisherID() samvaad.ParticipantID
	PublisherIdentity() samvaad.ParticipantIdentity

	AddSubscriber(sub LocalParticipant) (DataDownTrack, error)
	RemoveSubscriber(participantID samvaad.ParticipantID)
	IsSubscriber(subID samvaad.ParticipantID) bool

	AddDataDownTrack(sender DataTrackSender) error
	DeleteDataDownTrack(subscriberID samvaad.ParticipantID)

	HandlePacket(data []byte, packet *datatrack.Packet, arrivalTime int64)

	Close()
}

//counterfeiter:generate . DataDownTrack
type DataDownTrack interface {
	Close()

	Handle() uint16
	PublishDataTrack() DataTrack

	UpdateSubscriptionOptions(subscriptionOptions *samvaad.DataTrackSubscriptionOptions)
}

//counterfeiter:generate . DataTrackSender
type DataTrackSender interface {
	SubscriberID() samvaad.ParticipantID

	WritePacket(data []byte, packet *datatrack.Packet, arrivalTime int64)
}

//counterfeiter:generate . DataTrackTransport
type DataTrackTransport interface {
	SendDataTrackMessage(data []byte) error
}

//counterfeiter:generate . SubscribedTrack
type SubscribedTrack interface {
	AddOnBind(f func(error))
	IsBound() bool
	Close(isExpectedToResume bool)
	OnClose(f func(isExpectedToResume bool))
	ID() samvaad.TrackID
	PublisherID() samvaad.ParticipantID
	PublisherIdentity() samvaad.ParticipantIdentity
	PublisherVersion() uint32
	SubscriberID() samvaad.ParticipantID
	SubscriberIdentity() samvaad.ParticipantIdentity
	Subscriber() LocalParticipant
	DownTrack() *sfu.DownTrack
	MediaTrack() MediaTrack
	RTPSender() *webrtc.RTPSender
	IsMuted() bool
	SetPublisherMuted(muted bool)
	UpdateSubscriberSettings(settings *samvaad.UpdateTrackSettings, isImmediate bool)
	// selects appropriate video layer according to subscriber preferences
	UpdateVideoLayer()
	NeedsNegotiation() bool
}

type ChangeNotifier interface {
	AddObserver(key string, onChanged func())
	RemoveObserver(key string)
	HasObservers() bool
	NotifyChanged()
}

type MediaResolverResult struct {
	TrackChangedNotifier ChangeNotifier
	TrackRemovedNotifier ChangeNotifier
	Track                MediaTrack
	// is permission given to the requesting participant
	HasPermission     bool
	PublisherID       samvaad.ParticipantID
	PublisherIdentity samvaad.ParticipantIdentity
}

type DataResolverResult struct {
	TrackChangedNotifier ChangeNotifier
	TrackRemovedNotifier ChangeNotifier
	DataTrack            DataTrack
	PublisherID          samvaad.ParticipantID
	PublisherIdentity    samvaad.ParticipantIdentity
}

// MediaTrackResolver locates a specific media track for a subscriber
type MediaTrackResolver func(LocalParticipant, samvaad.TrackID) MediaResolverResult

// DataTrackResolver locates a specific data track for a subscriber
type DataTrackResolver func(LocalParticipant, samvaad.TrackID) DataResolverResult

// Supervisor/operation monitor related definitions
type OperationMonitorEvent int

const (
	OperationMonitorEventPublisherPeerConnectionConnected OperationMonitorEvent = iota
	OperationMonitorEventAddPendingPublication
	OperationMonitorEventSetPublicationMute
	OperationMonitorEventSetPublishedTrack
	OperationMonitorEventClearPublishedTrack
)

func (o OperationMonitorEvent) String() string {
	switch o {
	case OperationMonitorEventPublisherPeerConnectionConnected:
		return "PUBLISHER_PEER_CONNECTION_CONNECTED"
	case OperationMonitorEventAddPendingPublication:
		return "ADD_PENDING_PUBLICATION"
	case OperationMonitorEventSetPublicationMute:
		return "SET_PUBLICATION_MUTE"
	case OperationMonitorEventSetPublishedTrack:
		return "SET_PUBLISHED_TRACK"
	case OperationMonitorEventClearPublishedTrack:
		return "CLEAR_PUBLISHED_TRACK"
	default:
		return fmt.Sprintf("%d", int(o))
	}
}

type OperationMonitorData any

type OperationMonitor interface {
	PostEvent(ome OperationMonitorEvent, omd OperationMonitorData)
	Check() error
	IsIdle() bool
}


