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
)

func IsConnectionQualityLower(prev samvaad.ConnectionQuality, curr samvaad.ConnectionQuality) bool {
	if prev != samvaad.ConnectionQuality_LOST && curr == samvaad.ConnectionQuality_LOST {
		return true
	}

	if prev == samvaad.ConnectionQuality_LOST && curr != samvaad.ConnectionQuality_LOST {
		return false
	}

	// WARNING NOTE: comparing protobuf enums directly
	return curr < prev
}

func IsConnectionQualityHigher(prev samvaad.ConnectionQuality, curr samvaad.ConnectionQuality) bool {
	if IsConnectionQualityLower(prev, curr) {
		return false
	}

	return curr != prev
}
