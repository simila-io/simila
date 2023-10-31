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
	"github.com/acquirecloud/golibs/logging"
	"github.com/gin-gonic/gin"
	"net/http"
)

// HttpEP provides the api endpoints for the HTTP interface
type HttpEP struct {
	logger logging.Logger
}

func NewHttpEP() *HttpEP {
	return &HttpEP{logger: logging.NewLogger("api.rest")}
}

func (hep *HttpEP) RegisterEPs(g *gin.Engine) error {
	g.GET("/v1/ping", hep.hGetPing)
	return nil
}

func (hep *HttpEP) hGetPing(c *gin.Context) {
	hep.logger.Debugf("ping")
	c.String(http.StatusOK, "pong")
}
