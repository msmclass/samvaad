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

package datatrack

import (
	"errors"

	"github.com/msmclass/samvaad/pkg/proto/samvaad"
)

type ExtensionParticipantSid struct {
	participantID samvaad.ParticipantID
}

func NewExtensionParticipantSid(participantID samvaad.ParticipantID) (*ExtensionParticipantSid, error) {
	if len(participantID) >= 256 {
		return nil, errors.New("participantID too long")
	}

	return &ExtensionParticipantSid{participantID}, nil
}

func (e *ExtensionParticipantSid) ParticipantID() samvaad.ParticipantID {
	return e.participantID
}

func (e *ExtensionParticipantSid) Marshal() (Extension, error) {
	data := make([]byte, len(e.participantID))
	copy(data, e.participantID)
	return Extension{
		id:   uint8(samvaad.DataTrackExtensionID_DTEI_PARTICIPANT_SID),
		data: data,
	}, nil
}

func (e *ExtensionParticipantSid) Unmarshal(ext Extension) error {
	if ext.id != uint8(samvaad.DataTrackExtensionID_DTEI_PARTICIPANT_SID) {
		return errors.New("invalid extension ID")
	}

	if len(ext.data) == 0 {
		return errors.New("empty extension data")
	}

	e.participantID = samvaad.ParticipantID(ext.data)
	return nil
}


