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

// Package streamallocator provides bandwidth-aware track allocation for the
// Samvaad SFU. This file adds a CPU-aware layer governor that enforces
// simulcast layer caps when the node is under CPU pressure.
//
// Governor tiers (sampled every 2 s via go-osstat):
//
//	CPUTierHigh      — CPU < 60%   : all layers permitted
//	CPUTierMid       — CPU 60–80%  : cap at 720p (spatial layer 1)
//	CPUTierLow       — CPU 80–95%  : cap at 360p (spatial layer 0)
//	CPUTierAudioOnly — CPU ≥ 95%   : video forwarding paused; audio preserved

package streamallocator

import (
	"context"
	"runtime"
	"sync/atomic"
	"time"

	"github.com/msmclass/samvaad/pkg/samvaad/logger"
	osstat "github.com/mackerelio/go-osstat/cpu"
)

// CPUTier represents the simulcast quality tier enforced by the CPU governor.
type CPUTier int32

const (
	// CPUTierHigh — CPU load below 60%: all simulcast layers are available.
	CPUTierHigh CPUTier = iota
	// CPUTierMid — CPU load 60–80%: capped at the medium spatial layer (720p).
	CPUTierMid
	// CPUTierLow — CPU load 80–95%: capped at the low spatial layer (360p).
	CPUTierLow
	// CPUTierAudioOnly — CPU load ≥ 95%: all video forwarding is paused.
	// Audio tracks are never affected by the governor.
	CPUTierAudioOnly
)

func (t CPUTier) String() string {
	switch t {
	case CPUTierHigh:
		return "high"
	case CPUTierMid:
		return "mid"
	case CPUTierLow:
		return "low"
	case CPUTierAudioOnly:
		return "audio_only"
	default:
		return "unknown"
	}
}

// MaxSpatialLayer returns the maximum spatial layer index permitted under this
// CPU tier. Returns -1 when video should be completely paused (audio-only).
func (t CPUTier) MaxSpatialLayer() int32 {
	switch t {
	case CPUTierHigh:
		return 2 // up to 1080p
	case CPUTierMid:
		return 1 // up to 720p
	case CPUTierLow:
		return 0 // up to 360p
	default: // CPUTierAudioOnly
		return -1
	}
}

// cpuGovernorThresholds holds the CPU-usage percentage boundaries for tier
// transitions. They are package-level vars so tests can override them.
var (
	cpuThresholdMid       float64 = 60.0
	cpuThresholdLow       float64 = 80.0
	cpuThresholdAudioOnly float64 = 95.0
)

// CPUGovernorParams holds dependencies for the CPU governor goroutine.
type CPUGovernorParams struct {
	// SampleInterval is how often CPU usage is measured. Default: 2 s.
	SampleInterval time.Duration
	Logger         logger.Logger
}

// CPUGovernor samples CPU usage at a fixed interval and maintains the current
// CPUTier in an atomic variable for lock-free reads by the allocator hot-path.
type CPUGovernor struct {
	params      CPUGovernorParams
	currentTier atomic.Int32 // stores CPUTier values
}

// NewCPUGovernor creates a CPUGovernor. Call Start() to begin sampling.
func NewCPUGovernor(params CPUGovernorParams) *CPUGovernor {
	if params.SampleInterval <= 0 {
		params.SampleInterval = 2 * time.Second
	}
	g := &CPUGovernor{params: params}
	// Default to high tier until the first sample completes.
	g.currentTier.Store(int32(CPUTierHigh))
	return g
}

// CurrentTier returns the most recently measured CPU tier. This method is safe
// to call from multiple goroutines without locking.
func (g *CPUGovernor) CurrentTier() CPUTier {
	return CPUTier(g.currentTier.Load())
}

// Start launches the CPU sampling loop. It runs until ctx is cancelled.
func (g *CPUGovernor) Start(ctx context.Context) {
	go g.loop(ctx)
}

func (g *CPUGovernor) loop(ctx context.Context) {
	ticker := time.NewTicker(g.params.SampleInterval)
	defer ticker.Stop()

	// Capture the initial CPU counters so the first delta is meaningful.
	prev, err := osstat.Get()
	if err != nil {
		g.params.Logger.Warnw("[cpu-governor] cannot read initial CPU stats, defaulting to high tier", err)
		return
	}

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			curr, err := osstat.Get()
			if err != nil {
				g.params.Logger.Warnw("[cpu-governor] CPU stat read error", err)
				// Keep the previous snapshot so next iteration can produce a delta.
				continue
			}

			pct := cpuUsagePct(prev, curr)
			tier := tierFromPct(pct)
			prev = curr

			old := CPUTier(g.currentTier.Swap(int32(tier)))
			if old != tier {
				g.params.Logger.Infow("[cpu-governor] tier changed",
					"cpu_pct", pct,
					"old_tier", old.String(),
					"new_tier", tier.String(),
					"max_spatial_layer", tier.MaxSpatialLayer(),
					"gomaxprocs", runtime.GOMAXPROCS(0),
				)
			}
		}
	}
}

// cpuUsagePct computes the CPU busy-percentage between two stat snapshots.
// The go-osstat Stats struct provides: User, System, Idle, Nice, Total.
// We compute: busy / total * 100 where busy = Total - Idle.
//
// Both snapshots are pointers as returned by osstat.Get().
func cpuUsagePct(prev, curr *osstat.Stats) float64 {
	totalDelta := curr.Total - prev.Total
	if totalDelta == 0 {
		return 0
	}

	idleDelta := curr.Idle - prev.Idle
	busyDelta := float64(totalDelta) - float64(idleDelta)
	return (busyDelta / float64(totalDelta)) * 100.0
}

// tierFromPct maps a raw CPU percentage to the appropriate CPUTier.
func tierFromPct(pct float64) CPUTier {
	switch {
	case pct >= cpuThresholdAudioOnly:
		return CPUTierAudioOnly
	case pct >= cpuThresholdLow:
		return CPUTierLow
	case pct >= cpuThresholdMid:
		return CPUTierMid
	default:
		return CPUTierHigh
	}
}
