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

package service

import (
	"context"

	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/msmclass/samvaad/pkg/proto/samvaad"
	"github.com/msmclass/samvaad/pkg/samvaad/rpc"
)

//counterfeiter:generate . IOClient
type IOClient interface {
	CreateEgress(ctx context.Context, info *samvaad.EgressInfo) (*emptypb.Empty, error)
	GetEgress(ctx context.Context, req *rpc.GetEgressRequest) (*samvaad.EgressInfo, error)
	ListEgress(ctx context.Context, req *samvaad.ListEgressRequest) (*samvaad.ListEgressResponse, error)
	CreateIngress(ctx context.Context, req *samvaad.IngressInfo) (*emptypb.Empty, error)
	UpdateIngressState(ctx context.Context, req *rpc.UpdateIngressStateRequest) (*emptypb.Empty, error)
}


