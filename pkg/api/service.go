// Copyright 2023 The Simila Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package api

import (
	"context"
	"github.com/simila-io/simila/api/gen/index/v1"
	"google.golang.org/protobuf/types/known/emptypb"
)

// Service implements the gRPC API endpoints
type Service struct {
	index.UnimplementedServiceServer
}

var _ index.ServiceServer = (*Service)(nil)

func NewService() *Service {
	return &Service{}
}

// Create is a part of index.ServiceServer gRPC interface
func (s Service) Create(ctx context.Context, request *index.CreateIndexRequest) (*emptypb.Empty, error) {
	return nil, nil
}
