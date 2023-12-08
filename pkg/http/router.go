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

package http

import (
	"context"
	"fmt"
	"github.com/acquirecloud/golibs/logging"
	"github.com/gin-gonic/gin"
	"github.com/logrange/linker"
	"net/http"
	"time"
)

// Config provides the Http router configuration
type Config struct {
	// HttpAddr specifies the address to listen incoming HTTP connections
	HttpAddr string // could be empty
	// HttpPort specifies the listening port for the incoming HTTP connections
	HttpPort int
	// RestRegistrar is the endpoints registrar
	RestRegistrar EndpointsRegistrar
}

// EndpointsRegistrar is a component which provides a callback for registering REST endpoints in the Router server
type EndpointsRegistrar func(*gin.Engine) error

// Router service manages http endpoints
type Router struct {
	linker.PostConstructor
	linker.Initializer
	linker.Shutdowner

	config Config

	r      *gin.Engine
	srv    *http.Server
	logger logging.Logger
}

// NewRouter creates new Router instance
func NewRouter(cfg Config) *Router {
	return &Router{config: cfg}
}

// PostConstruct implements linker.PostConstructor
func (r *Router) PostConstruct() {
	r.logger = logging.NewLogger("rest.Router")
}

// Init implements linker.Initializer
func (r *Router) Init(_ context.Context) error {
	r.logger.Infof("Initializing")
	r.r = gin.Default()
	r.r.UseRawPath = true
	r.r.UnescapePathValues = false

	if r.config.RestRegistrar == nil {
		r.logger.Warnf("RestRegistrar is not provided, will register /ping only...")
		r.registerPingOnly(r.r)
	} else {
		err := r.config.RestRegistrar(r.r)
		if err != nil {
			return fmt.Errorf("could not add endpoints into HTTP Router %w", err)
		}
	}

	addr := fmt.Sprintf("%s:%d", r.config.HttpAddr, r.config.HttpPort)
	r.srv = &http.Server{
		Addr:    addr,
		Handler: r.r,
	}

	go func() {
		// service connections
		r.logger.Infof("Starting serving connections")
		_ = r.srv.ListenAndServe()
	}()

	r.logger.Infof("Initialized to serve HTTP requests on %s", addr)
	return nil
}

// Shutdown implements linker.Shutdowner
func (r *Router) Shutdown() {
	r.logger.Infof("Shutdown...")
	if r.srv == nil {
		r.logger.Infof("Router seems to be not initialized. Skipping shutdown action.")
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	_ = r.srv.Shutdown(ctx)
}

func (r *Router) registerPingOnly(g *gin.Engine) error {
	g.GET("/ping", r.hGetPing)
	return nil
}

func (r *Router) hGetPing(c *gin.Context) {
	r.logger.Debugf("ping")
	c.String(http.StatusOK, "pong")
}
