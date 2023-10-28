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
	"github.com/simila-io/simila/pkg/indexer/persistence"
	"github.com/simila-io/simila/pkg/version"

	"github.com/davecgh/go-spew/spew"
	"github.com/logrange/linker"
)

// Run is an entry point of the Simila server
func Run(ctx context.Context, cfg *Config) error {
	log := logging.NewLogger("server")
	log.Infof("starting server: %s", version.BuildVersionString())

	log.Infof(spew.Sprint(cfg))
	defer log.Infof("server is stopped")

	// DB
	db := persistence.NewDb(cfg.DB.Driver, cfg.DB.SourceName())
	migrations := persistence.NewMigration(cfg.DB.Driver, cfg.DB.SourceName())

	inj := linker.New()
	inj.Register(linker.Component{Name: "", Value: db})
	inj.Register(linker.Component{Name: "", Value: migrations})

	inj.Init(ctx)
	<-ctx.Done()
	inj.Shutdown()
	return nil
}
