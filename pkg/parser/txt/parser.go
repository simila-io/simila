package txt

import (
	"context"
	"github.com/acquirecloud/golibs/logging"
	"github.com/logrange/linker"
	"github.com/simila-io/simila/pkg/indexer/persistence"
	"github.com/simila-io/simila/pkg/parser"
	"io"
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

	//scanner := bufio.NewScanner(body)
	//line := 1
	//for scanner.Scan() {
	//	segment := scanner.Text()
	//}
	return 0, nil
}

func (tp *txtParser) String() string {
	return "txtParser"
}
