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
package commands

import (
	"github.com/acquirecloud/golibs/context"
	"github.com/simila-io/simila/pkg/server"
	"github.com/spf13/cobra"
	"os"
	"syscall"
)

var startCmd = &cobra.Command{
	Use: "start",
	RunE: func(c *cobra.Command, args []string) error {
		configPath, _ := c.Flags().GetString("config")
		cfg, err := server.BuildConfig(configPath)
		if err != nil {
			return err
		}
		mainCtx := context.NewSignalsContext(os.Interrupt, syscall.SIGTERM)
		return server.Run(mainCtx, cfg)
	},
}
