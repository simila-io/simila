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

	line := 0
	for scanner.Scan() {
		line++
		sgmnt := scanner.Text()
		sgmnt = strings.Trim(sgmnt, "\t\n\v\f\r\x85\xA0")
		if len(sgmnt) == 0 {
			continue
		}
		recs = append(recs, persistence.IndexRecord{IndexID: idxId, ID: fmt.Sprintf("%08x", line), Segment: scanner.Text()})
		if len(recs) >= 100 {
			if err := mtx.CreateIndexRecords(recs...); err != nil {
				return 0, err
			}
			recs = recs[:0]
		}
	}

	if len(recs) > 0 {
		if err := mtx.CreateIndexRecords(recs...); err != nil {
			return 0, err
		}
	}

	return int64(line), nil
}

func (tp *txtParser) String() string {
	return "txtParser"
}
