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

package txt

import (
	"bufio"
	"context"
	"fmt"
	"github.com/acquirecloud/golibs/logging"
	"github.com/logrange/linker"
	"github.com/simila-io/simila/pkg/indexer/persistence"
	"github.com/simila-io/simila/pkg/parser"
	"io"
	"strings"
)

type (
	txtParser struct {
		PProvider parser.Provider `inject:""`

		logger logging.Logger
	}
)

var _ parser.Parser = (*txtParser)(nil)

func New() *txtParser {
	return new(txtParser)
}

var _ linker.Initializer = (*txtParser)(nil)

func (tp *txtParser) Init(ctx context.Context) error {
	tp.logger = logging.NewLogger("parser.txt")
	tp.logger.Infof("Initializing")
	tp.PProvider.RegisterParser("txt", tp)
	return nil
}

func (tp *txtParser) ScanRecords(ctx context.Context, mtx persistence.ModelTx, nodeID int64, body io.Reader) (int64, error) {
	tp.logger.Infof("scanning for node id=%d", nodeID)

	scanner := bufio.NewScanner(body)
	var recs []persistence.IndexRecord

	line := 0
	for scanner.Scan() {
		line++
		sgmnt := scanner.Text()
		trimmed := strings.Trim(sgmnt, " \t\n\v\f\r\x85\xA0")
		if len(trimmed) == 0 {
			continue
		}
		recs = append(recs, persistence.IndexRecord{NodeID: nodeID, ID: fmt.Sprintf("%08x", line), Segment: sgmnt, RankMult: 1.0})
		if len(recs) >= 100 {
			if err := mtx.UpsertIndexRecords(recs...); err != nil {
				return 0, err
			}
			recs = recs[:0]
		}
	}

	if len(recs) > 0 {
		if err := mtx.UpsertIndexRecords(recs...); err != nil {
			return 0, err
		}
	}

	return int64(line), nil
}

func (tp *txtParser) String() string {
	return "txtParser"
}
