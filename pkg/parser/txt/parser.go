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

func (tp *txtParser) ScanRecords(ctx context.Context, mtx persistence.ModelTx, idxId string, body io.Reader) (int64, error) {
	tp.logger.Infof("scanning for id=%s", idxId)

	scanner := bufio.NewScanner(body)
	var recs []persistence.IndexRecord

	start := ""
	paragraph := 0
	for {
		var sb strings.Builder
		sb.WriteString(start)
		start = ""
		for scanner.Scan() {
			ln := scanner.Text()
			trimmed := strings.Trim(ln, " \t\n\v\f\r\x85\xA0")
			if len(trimmed) == 0 {
				// end of paragraph?
				if sb.Len() > 0 {
					break
				}
				continue
			}
			sb.WriteString(" ")
			sb.WriteString(trimmed)
			if sb.Len() > 2048 {
				s := sb.String()
				sb.Reset()
				idx := strings.LastIndex(s, ".")
				if idx > -1 {
					start = s[idx+1:]
					s = s[:idx+1]
				}
				sb.WriteString(s)
			}
		}
		if sb.Len() == 0 {
			break
		}
		paragraph++
		recs = append(recs, persistence.IndexRecord{IndexID: idxId, ID: fmt.Sprintf("%08x", paragraph), Segment: sb.String()})
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

	return int64(paragraph), nil
}

func (tp *txtParser) String() string {
	return "txtParser"
}
