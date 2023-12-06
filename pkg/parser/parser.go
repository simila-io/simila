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

package parser

import (
	"context"
	"github.com/acquirecloud/golibs/container"
	"github.com/acquirecloud/golibs/logging"
	"github.com/simila-io/simila/pkg/indexer/persistence"
	"io"
	"sync"
	"sync/atomic"
)

type (
	// Parser allows to scan data in a specific format and update the index records
	Parser interface {
		// ScanRecords walks through the body, extracts the document records, and writes them via mtx associating with the nodeId
		ScanRecords(ctx context.Context, mtx persistence.ModelTx, nodeId int64, body io.Reader) (int64, error)
	}

	// Provider is a parsers holder.
	Provider interface {
		RegisterParser(format string, p Parser)
		Parser(format string) Parser
	}

	psMap map[string]Parser
	pp    struct {
		lock   sync.Mutex
		ps     atomic.Value
		logger logging.Logger
	}
)

var _ Provider = (*pp)(nil)

func NewParserProvider() *pp {
	ppr := &pp{}
	ppr.ps.Store(psMap{})
	ppr.logger = logging.NewLogger("parser.Provider")
	return ppr
}

func (ppr *pp) RegisterParser(format string, p Parser) {
	ppr.lock.Lock()
	defer ppr.lock.Unlock()
	ppr.logger.Infof("registering parser %v for %s", p, format)
	pm := ppr.ps.Load().(psMap)
	pm = container.CopyMap(pm)
	pm[format] = p
	ppr.ps.Store(pm)
}

func (ppr *pp) Parser(format string) Parser {
	pm := ppr.ps.Load().(psMap)
	return pm[format]
}
