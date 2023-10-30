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

package grpc

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/acquirecloud/golibs/logging"
	"github.com/acquirecloud/golibs/transport"
	"net"
	"sync/atomic"

	"github.com/logrange/linker"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// Config is used for providing settings to the server
type Config struct {
	// A network transport configuration
	Transport transport.Config
	// RegisterEndpoints allows to add gRPC endpoints into the server
	RegisterEndpoints RegisterF
}

// RegisterF is a function which allows to add endpoints into the server. It is called in Init
type RegisterF func(*grpc.Server) error

// Server provides gRPC server interface. it can be used by one or many versions. The Server is about
// transport layer, it doesn't provide any specific details like concrete APIs, which should be
// registered in RegisterF function callback, provided in the Server config.
type Server struct {
	// Config contains the server connection interface
	cfg Config

	listnr net.Listener
	closed int32
	logger logging.Logger
}

// NewServer creates a new instance of the Server
func NewServer(cfg Config) *Server {
	return &Server{cfg: cfg}
}

var _ linker.Initializer = (*Server)(nil)
var _ linker.Shutdowner = (*Server)(nil)

// Init is part of linker.Initializer. it is called by the dependency injection mechanism,
// so it must be called once and it is not thread-safe
func (s *Server) Init(ctx context.Context) error {
	s.logger = logging.NewLogger("grpc.Server")
	s.logger.Infof("Initializing with %s %s", s.cfg.Transport.Network, s.cfg.Transport.Addr())

	lis, err := transport.NewServerListener(s.cfg.Transport)
	if err != nil {
		return fmt.Errorf("could not start listener for %v, err=%w", s.cfg.Transport, err)
	}

	s.listnr = lis
	gs := grpc.NewServer()
	err = s.cfg.RegisterEndpoints(gs)
	if err != nil {
		return fmt.Errorf("could not register endpoints: %w", err)
	}

	// Register reflection service on gRPC server.
	reflection.Register(gs)
	go func() {
		s.logger.Infof("Starting go routine by listening gRPC simila connections")
		if err := gs.Serve(lis); err != nil && atomic.LoadInt32(&s.closed) == 0 {
			s.logger.Errorf("failed to serve: %v", err)
		}
		s.logger.Infof("Finishing go routine by listening gRPC Server connections")
	}()

	return nil
}

// Shutdown is an implementation of linker.Shutdowner. It must be called once, not thread-safe.
func (s *Server) Shutdown() {
	s.logger.Infof("Shutting down...")
	if s.listnr != nil {
		atomic.StoreInt32(&s.closed, 1)
		s.listnr.Close()
	}
}

// String implements fmt.Stringify
func (c *Config) String() string {
	b, _ := json.MarshalIndent(*c, "", "  ")
	return string(b)
}
