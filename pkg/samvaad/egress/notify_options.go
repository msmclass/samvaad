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
	"github.com/msmclass/samvaad/pkg/proto/samvaad"
	"github.com/msmclass/samvaad/pkg/samvaad/webhook"
)

func GetEgressNotifyOptions(egressInfo *samvaad.EgressInfo) []webhook.NotifyOption {
	if egressInfo == nil {
		return nil
	}

	if egressInfo.Request == nil {
		return nil
	}

	var whs []*samvaad.WebhookConfig

	switch req := egressInfo.Request.(type) {
	// case *samvaad.EgressInfo_Egress:
	// 	if req.Egress != nil {
	// 		whs = req.Egress.Webhooks
	// 	}
	case *samvaad.EgressInfo_Replay:
		if req.Replay != nil {
			whs = req.Replay.Webhooks
		}
	case *samvaad.EgressInfo_RoomComposite:
		if req.RoomComposite != nil {
			whs = req.RoomComposite.Webhooks
		}
	case *samvaad.EgressInfo_Web:
		if req.Web != nil {
			whs = req.Web.Webhooks
		}
	case *samvaad.EgressInfo_Participant:
		if req.Participant != nil {
			whs = req.Participant.Webhooks
		}
	case *samvaad.EgressInfo_TrackComposite:
		if req.TrackComposite != nil {
			whs = req.TrackComposite.Webhooks
		}
	case *samvaad.EgressInfo_Track:
		if req.Track != nil {
			whs = req.Track.Webhooks
		}
	}

	if len(whs) > 0 {
		return []webhook.NotifyOption{webhook.WithExtraWebhooks(whs)}
	}

	return nil
}
