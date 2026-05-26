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
	"github.com/msmclass/samvaad/pkg/samvaad/utils/guid"
)

const (
	GuidSize = guid.Size
)

const (
	RoomPrefix                         = guid.RoomPrefix
	NodePrefix                         = guid.NodePrefix
	ParticipantPrefix                  = guid.ParticipantPrefix
	TrackPrefix                        = guid.TrackPrefix
	DataTrackPrefix                    = guid.DataTrackPrefix
	APIKeyPrefix                       = guid.APIKeyPrefix
	EgressPrefix                       = guid.EgressPrefix
	IngressPrefix                      = guid.IngressPrefix
	SIPTrunkPrefix                     = guid.SIPTrunkPrefix
	SIPDispatchRulePrefix              = guid.SIPDispatchRulePrefix
	SIPCallPrefix                      = guid.SIPCallPrefix
	SIPTransferPrefix                  = guid.SIPTransferPrefix
	RPCPrefix                          = guid.RPCPrefix
	WHIPResourcePrefix                 = guid.WHIPResourcePrefix
	RTMPResourcePrefix                 = guid.RTMPResourcePrefix
	URLResourcePrefix                  = guid.URLResourcePrefix
	SIPHostnamePrefix                  = guid.SIPHostnamePrefix
	AgentWorkerPrefix                  = guid.AgentWorkerPrefix
	AgentJobPrefix                     = guid.AgentJobPrefix
	AgentDispatchPrefix                = guid.AgentDispatchPrefix
	CloudAgentPrefix                   = guid.CloudAgentPrefix
	CloudAgentRegionPrefix             = guid.CloudAgentRegionPrefix
	CloudAgentVersionPrefix            = guid.CloudAgentVersionPrefix
	CloudAgentSecretPrefix             = guid.CloudAgentSecretPrefix
	CloudAgentWorkerPrefix             = guid.CloudAgentWorkerPrefix
	CloudAgentPrivateLinkPrefix        = guid.CloudAgentPrivateLinkPrefix
	CloudAgentPrivateLinkClusterPrefix = guid.CloudAgentPrivateLinkClusterPrefix
	CloudAgentPrivateLinkGatewayPrefix = guid.CloudAgentPrivateLinkGatewayPrefix
	CloudAgentPrivateLinkSecretPrefix  = guid.CloudAgentPrivateLinkSecretPrefix
)

func NewGuid(prefix string) string {
	return guid.New(prefix)
}

// HashedID creates a hashed ID from a unique string
func HashedID(id string) string {
	return guid.HashedID(id)
}

func LocalNodeID() (string, error) {
	return guid.LocalNodeID()
}

func MarshalGuid[T samvaad.Guid](id T) samvaad.GuidBlock {
	return guid.Marshal(id)
}

func UnmarshalGuid[T samvaad.Guid](b samvaad.GuidBlock) T {
	return guid.Unmarshal[T](b)
}
