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
	"encoding/json"
	"fmt"
	"github.com/acquirecloud/golibs/config"
	"github.com/acquirecloud/golibs/logging"
	"github.com/acquirecloud/golibs/transport"
	"github.com/simila-io/simila/pkg/indexer/persistence/postgres"
)

type (
	// Config defines the scaffolding-golang server configuration
	Config struct {
		// GrpcTransport specifies grpc transport configuration
		GrpcTransport *transport.Config
		// HttpPort defines the port for listening incoming HTTP connections
		HttpPort int
		// SearchEngine specifies which engine is used for search
		SearchEngine string
		// DB specifies settings for DB used as a full text search engine (e.g. postgres)
		DB *DB
	}

	DB struct {
		Driver   string
		Host     string
		Port     string
		Username string
		Password string
		DBName   string
		SSLMode  string
	}
)

func (d *DB) SourceName() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		d.Host, d.Port, d.Username, d.Password, d.DBName, d.SSLMode)
}

func (d *DB) URL() string {
	return fmt.Sprintf("%s://%s:%s@%s:%s/%s?sslmode=%s",
		d.Driver, d.Username, d.Password, d.Host, d.Port, d.DBName, d.SSLMode)
}

// getDefaultConfig returns the default server config
func getDefaultConfig() *Config {
	return &Config{
		GrpcTransport: transport.GetDefaultGRPCConfig(),
		HttpPort:      8080,
		SearchEngine:  postgres.SearchModuleTrigram,
		DB: &DB{
			Driver:   "postgres",
			Host:     "localhost",
			Port:     "5432",
			Username: "postgres",
			Password: "postgres",
			DBName:   "simila",
			SSLMode:  "disable",
		},
	}
}

func BuildConfig(cfgFile string) (*Config, error) {
	log := logging.NewLogger("simila.ConfigBuilder")
	log.Infof("trying to build config. cfgFile=%s", cfgFile)
	e := config.NewEnricher(*getDefaultConfig())
	fe := config.NewEnricher(Config{})
	err := fe.LoadFromFile(cfgFile)
	if err != nil {
		return nil, fmt.Errorf("could not read data from the file %s: %w", cfgFile, err)
	}
	// overwrite default
	_ = e.ApplyOther(fe)
	_ = e.ApplyEnvVariables("SIMILA", "_")
	cfg := e.Value()
	return &cfg, nil
}

// String implements fmt.Stringify interface in a pretty console form
func (c *Config) String() string {
	b, _ := json.MarshalIndent(*c, "", "  ")
	return string(b)
}
