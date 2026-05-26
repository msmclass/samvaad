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

package clientconfiguration

import (
	"github.com/msmclass/samvaad/pkg/samvaad/codecs/mime"
	"github.com/msmclass/samvaad/pkg/proto/samvaad"
	"github.com/msmclass/samvaad/pkg/samvaad/utils/must"
)

// StaticConfigurations list specific device-side limitations that should be disabled at a global level
var StaticConfigurations = []ConfigurationItem{
	// {
	// 	Match:         must.Get(NewScriptMatch(`c.protocol <= 5 || c.browser == "firefox"`)),
	// 	Configuration: &samvaad.ClientConfiguration{ResumeConnection: samvaad.ClientConfigSetting_DISABLED},
	// 	Merge:         false,
	// },
	{
		Match: must.Get(NewScriptMatch(`c.browser == "safari"`)),
		Configuration: &samvaad.ClientConfiguration{
			DisabledCodecs: &samvaad.DisabledCodecs{
				Codecs: []*samvaad.Codec{
					{Mime: mime.MimeTypeAV1.String()},
				},
			},
		},
		Merge: true,
	},
	{
		Match: must.Get(NewScriptMatch(`c.browser == "safari" && c.browser_version > "18.3"`)),
		Configuration: &samvaad.ClientConfiguration{
			DisabledCodecs: &samvaad.DisabledCodecs{
				Publish: []*samvaad.Codec{
					{Mime: mime.MimeTypeVP9.String()},
				},
			},
		},
		Merge: true,
	},
	{
		Match: must.Get(NewScriptMatch(`(c.device_model == "xiaomi 2201117ti" && c.os == "android") ||
		  ((c.browser == "firefox" || c.browser == "firefox mobile") && (c.os == "linux" || c.os == "android"))`)),
		Configuration: &samvaad.ClientConfiguration{
			DisabledCodecs: &samvaad.DisabledCodecs{
				Publish: []*samvaad.Codec{
					{Mime: mime.MimeTypeH264.String()},
				},
			},
		},
		Merge: false,
	},
}


