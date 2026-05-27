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

//go:build !darwin || cgo

package streamallocator

import (
	osstat "github.com/mackerelio/go-osstat/cpu"
)

type platformCPUSampler struct {
	prev *osstat.Stats
}

func newPlatformCPUSampler() (*platformCPUSampler, error) {
	prev, err := osstat.Get()
	if err != nil {
		return nil, err
	}
	return &platformCPUSampler{prev: prev}, nil
}

func (s *platformCPUSampler) sample() (float64, error) {
	curr, err := osstat.Get()
	if err != nil {
		return 0, err
	}

	pct := cpuUsagePct(s.prev, curr)
	s.prev = curr
	return pct, nil
}

func cpuUsagePct(prev, curr *osstat.Stats) float64 {
	totalDelta := curr.Total - prev.Total
	if totalDelta == 0 {
		return 0
	}

	idleDelta := curr.Idle - prev.Idle
	busyDelta := float64(totalDelta) - float64(idleDelta)
	return (busyDelta / float64(totalDelta)) * 100.0
}
