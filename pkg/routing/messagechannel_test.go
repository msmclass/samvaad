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

package routing_test

import (
	"sync"
	"testing"

	"github.com/msmclass/samvaad/pkg/proto/samvaad"

	"github.com/msmclass/samvaad/pkg/routing"
)

func TestMessageChannel_WriteMessageClosed(t *testing.T) {
	// ensure it doesn't panic when written to after closing
	m := routing.NewMessageChannel(samvaad.ConnectionID("test"), routing.DefaultMessageChannelSize)
	go func() {
		for msg := range m.ReadChan() {
			if msg == nil {
				return
			}
		}
	}()

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for range 100 {
			_ = m.WriteMessage(&samvaad.SignalRequest{})
		}
	}()
	_ = m.WriteMessage(&samvaad.SignalRequest{})
	m.Close()
	_ = m.WriteMessage(&samvaad.SignalRequest{})

	wg.Wait()
}


