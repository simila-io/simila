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
package gateway

import (
	"context"
	"fmt"
	"github.com/acquirecloud/golibs/logging"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/logrange/linker"
	"net/http"
	"path"
	"strings"
	"time"
)

// Config provides the Http router configuration
type Config struct {
	// HttpAddr specifies the address to listen incoming HTTP connections
	HttpAddr string // could be empty
	// HttpPort specifies the listening port for the incoming HTTP connections
	HttpPort int
	// Mux is the reverse proxy to gRPC connection
	Mux *runtime.ServeMux
}

// Router service manages http endpoints
type Router struct {
	linker.PostConstructor
	linker.Initializer
	linker.Shutdowner

	srv    *http.Server
	config Config
	logger logging.Logger
}

// NewRouter creates new Router instance
func NewRouter(cfg Config) *Router {
	return &Router{config: cfg}
}

// PostConstruct implements linker.PostConstructor
func (r *Router) PostConstruct() {
	r.logger = logging.NewLogger("gateway.Router")
}

// Init implements linker.Initializer
func (r *Router) Init(_ context.Context) error {
	r.logger.Infof("Initializing")

	addr := fmt.Sprintf("%s:%d", r.config.HttpAddr, r.config.HttpPort)
	r.srv = &http.Server{
		Addr:    addr,
		Handler: r.config.Mux,
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

// ComposeURI helper function which composes URI, adding ID to the request path
func ComposeURI(r *http.Request, id string) string {
	var resURL string
	if r.URL.IsAbs() {
		resURL = path.Join(r.URL.String(), id)
	} else {
		resURL = ResolveScheme(r) + "://" + path.Join(ResolveHost(r), r.URL.String(), id)
	}
	return resURL
}

// ResolveScheme resolves initial request type by r
func ResolveScheme(r *http.Request) string {
	switch {
	case r.Header.Get("X-Forwarded-Proto") == "https":
		return "https"
	case r.URL.Scheme == "https":
		return "https"
	case r.TLS != nil:
		return "https"
	case strings.HasPrefix(r.Proto, "HTTPS"):
		return "https"
	default:
		return "http"
	}
}

// ResolveHost returns host part of r
func ResolveHost(r *http.Request) (host string) {
	switch {
	case r.Header.Get("X-Forwarded-For") != "":
		return r.Header.Get("X-Forwarded-For")
	case r.Header.Get("X-Host") != "":
		return r.Header.Get("X-Host")
	case r.Host != "":
		return r.Host
	case r.URL.Host != "":
		return r.URL.Host
	default:
		return ""
	}
}
