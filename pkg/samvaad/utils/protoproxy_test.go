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
	"runtime"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/msmclass/samvaad/pkg/proto/samvaad"
)

func TestProtoProxy(t *testing.T) {
	t.Run("basics", func(t *testing.T) {
		numGoRoutines := runtime.NumGoroutine()
		proxy, numParticipants, freeze := createTestProxy()

		select {
		case <-proxy.Updated():
			t.Fatal("should not have received an update")
		default:
		}

		// should not have changed, initial value should persist
		require.EqualValues(t, 0, proxy.Get().NumParticipants)

		// immediate change
		<-proxy.MarkDirty(true)
		awaitProxyUpdate(t, proxy)

		require.EqualValues(t, 2, numParticipants.Load())
		require.EqualValues(t, 1, proxy.Get().NumParticipants)

		// queue an update while the cached value is still fresh so the worker
		// picks it up on the next tick instead of triggering an immediate refresh.
		proxy.MarkDirty(false)
		assertNoProxyUpdate(t, proxy, 50*time.Millisecond)
		require.EqualValues(t, 1, proxy.Get().NumParticipants)

		// freeze and ensure that updates are not triggered
		freeze.Store(true)
		awaitProxyUpdate(t, proxy)
		require.EqualValues(t, 2, proxy.Get().NumParticipants)

		// trigger another update, but should not get notification as freeze is in place and the model should not have changed
		proxy.MarkDirty(false)
		assertNoProxyUpdate(t, proxy, 100*time.Millisecond)
		require.EqualValues(t, 2, proxy.Get().NumParticipants)

		// ensure we didn't leak
		proxy.Stop()

		for range 10 {
			if runtime.NumGoroutine() <= numGoRoutines {
				break
			}
			time.Sleep(100 * time.Millisecond)
		}
		require.LessOrEqual(t, runtime.NumGoroutine(), numGoRoutines)
	})

	t.Run("await next update after marking dirty", func(t *testing.T) {
		proxy, _, _ := createTestProxy()
		require.EqualValues(t, 0, proxy.Get().NumParticipants)
		<-proxy.MarkDirty(true)
		require.EqualValues(t, 1, proxy.Get().NumParticipants)
	})

	t.Run("await resolves when proxy is stopped", func(t *testing.T) {
		proxy, _, _ := createTestProxy()
		done := proxy.MarkDirty(true)
		proxy.Stop()
		<-done
	})

	t.Run("multiple awaits resolve for one update", func(t *testing.T) {
		proxy, _, _ := createTestProxy()
		done0 := proxy.MarkDirty(false)
		done1 := proxy.MarkDirty(true)
		<-done0
		<-done1
		require.EqualValues(t, 1, proxy.Get().NumParticipants)
	})

	t.Run("await resolve when there is no change", func(t *testing.T) {
		proxy := NewProtoProxy(10*time.Millisecond, func() *samvaad.Room { return nil })
		done := proxy.MarkDirty(true)
		time.Sleep(100 * time.Millisecond)
		select {
		case <-done:
		default:
			t.FailNow()
		}
	})
}

func awaitProxyUpdate(t *testing.T, proxy *ProtoProxy[*samvaad.Room]) {
	t.Helper()

	select {
	case <-proxy.Updated():
	case <-time.After(1 * time.Second):
		require.FailNow(t, "timed out waiting for proxy update")
	}
}

func assertNoProxyUpdate(t *testing.T, proxy *ProtoProxy[*samvaad.Room], d time.Duration) {
	t.Helper()

	select {
	case <-proxy.Updated():
		require.FailNow(t, "should not have received an update")
	case <-time.After(d):
	}
}

func createTestProxy() (*ProtoProxy[*samvaad.Room], *atomic.Uint32, *atomic.Bool) {
	// uses an update func that increments numParticipants each time
	var numParticipants atomic.Uint32
	var freeze atomic.Bool
	return NewProtoProxy(
		200*time.Millisecond,
		func() *samvaad.Room {
			if !freeze.Load() {
				defer numParticipants.Add(1)
			}
			return &samvaad.Room{
				NumParticipants: numParticipants.Load(),
			}
		},
	), &numParticipants, &freeze
}
