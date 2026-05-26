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
	"google.golang.org/protobuf/proto"

	"github.com/msmclass/samvaad/pkg/utils"
	"github.com/msmclass/samvaad/pkg/proto/samvaad"
	"github.com/msmclass/samvaad/pkg/samvaad/logger"
	protoutils "github.com/msmclass/samvaad/pkg/samvaad/utils"
)

type ConfigurationItem struct {
	Match
	Configuration *samvaad.ClientConfiguration
	Merge         bool
}

type StaticClientConfigurationManager struct {
	confs []ConfigurationItem
}

func NewStaticClientConfigurationManager(confs []ConfigurationItem) *StaticClientConfigurationManager {
	return &StaticClientConfigurationManager{confs: confs}
}

func (s *StaticClientConfigurationManager) GetConfiguration(clientInfo *samvaad.ClientInfo) *samvaad.ClientConfiguration {
	var matchedConf []*samvaad.ClientConfiguration
	for _, c := range s.confs {
		matched, err := c.Match.Match(clientInfo)
		if err != nil {
			logger.Errorw("matchrule failed", err,
				"clientInfo", logger.Proto(utils.ClientInfoWithoutAddress(clientInfo)),
			)
			continue
		}
		if !matched {
			continue
		}
		if !c.Merge {
			return c.Configuration
		}
		matchedConf = append(matchedConf, c.Configuration)
	}

	var conf *samvaad.ClientConfiguration
	for k, v := range matchedConf {
		if k == 0 {
			conf = protoutils.CloneProto(matchedConf[0])
		} else {
			// TODO : there is a problem use protobuf merge, we don't have flag to indicate 'no value',
			// don't override default behavior or other configuration's field. So a bool value = false or
			// a int value = 0 will override same field in other configuration
			proto.Merge(conf, v)
		}
	}
	return conf
}


