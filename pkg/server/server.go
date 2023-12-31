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

package server

import (
	"context"
	"github.com/acquirecloud/golibs/logging"
	"github.com/simila-io/simila/api/gen/format/v1"
	"github.com/simila-io/simila/api/gen/index/v1"
	"github.com/simila-io/simila/pkg/api"
	"github.com/simila-io/simila/pkg/grpc"
	"github.com/simila-io/simila/pkg/http"
	"github.com/simila-io/simila/pkg/indexer/persistence/postgres"
	"github.com/simila-io/simila/pkg/parser"
	"github.com/simila-io/simila/pkg/parser/txt"
	"github.com/simila-io/simila/pkg/version"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"

	"github.com/davecgh/go-spew/spew"
	"github.com/logrange/linker"
	ggrpc "google.golang.org/grpc"
)

// Run is an entry point of the Simila server
func Run(ctx context.Context, cfg *Config) error {
	log := logging.NewLogger("server")
	log.Infof("starting server: %s", version.BuildVersionString())

	log.Infof(spew.Sprint(cfg))
	defer log.Infof("server is stopped")

	// gRPC server
	gsvc := api.NewService()
	var grpcRegF grpc.RegisterF = func(gs *ggrpc.Server) error {
		grpc_health_v1.RegisterHealthServer(gs, health.NewServer())
		index.RegisterServiceServer(gs, gsvc.IndexServiceServer())
		format.RegisterServiceServer(gs, gsvc.FormatServiceServer())
		return nil
	}

	// Http API (endpoints)
	rst := api.NewRest(gsvc)

	// DB
	db := postgres.MustGetDb(ctx, cfg.DB.SourceName(), postgres.SearchModuleName(cfg.SearchEngine))

	inj := linker.New()
	inj.Register(linker.Component{Name: "", Value: db})
	inj.Register(linker.Component{Name: "", Value: gsvc})
	inj.Register(linker.Component{Name: "", Value: grpc.NewServer(grpc.Config{Transport: *cfg.GrpcTransport, RegisterEndpoints: grpcRegF})})
	inj.Register(linker.Component{Name: "", Value: http.NewRouter(http.Config{HttpPort: cfg.HttpPort, RestRegistrar: rst.RegisterEPs})})
	inj.Register(linker.Component{Name: "", Value: parser.NewParserProvider()})
	inj.Register(linker.Component{Name: "", Value: txt.New()})

	inj.Init(ctx)
	<-ctx.Done()
	inj.Shutdown()
	return nil
}
