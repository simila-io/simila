package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/acquirecloud/golibs/cast"
	"github.com/acquirecloud/golibs/logging"
	"github.com/simila-io/simila/api/gen/index/v1"
	"github.com/simila-io/simila/cmd/scli/commands"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"os"
	"path/filepath"
	"strings"

	"github.com/peterh/liner"
)

var (
	logger          = logging.NewLogger("scli")
	historyFileName = filepath.Join(os.TempDir(), ".scli_history")

	addr = flag.String("addr", "localhost:50051", "the address to connect to")
)

func main() {
	flag.Parse()
	host := cast.String(addr, "")
	conn, err := grpc.Dial(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Errorf("could not connect to %s: %s", host, err.Error())
		return
	}
	defer conn.Close()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	sc := index.NewServiceClient(conn)

	line := liner.NewLiner()
	defer line.Close()
	cs := commands.New(ctx, sc)
	line.SetCtrlCAborts(true)
	line.SetCompleter(func(ln string) (c []string) {
		for _, n := range cs.ListCommandsNames() {
			if strings.HasPrefix(n, strings.ToLower(ln)) {
				c = append(c, n)
			}
		}
		return
	})

	if f, err := os.Open(historyFileName); err == nil {
		line.ReadHistory(f)
		f.Close()
	}

	for {
		if p, err := line.Prompt(fmt.Sprintf("%s > ", host)); err == nil {
			line.AppendHistory(p)
			p = strings.Trim(p, commands.Spaces)
			if p == "" {
				continue
			}
			cmd, err := cs.GetCommand(p)
			if err != nil {
				fmt.Println(err.Error())
				continue
			}
			p = strings.TrimLeft(p[len(cmd.Prefix()):], commands.Spaces)
			if err := cmd.Run(p); err != nil {
				fmt.Println(err.Error())
			}
		} else if err == liner.ErrPromptAborted {
			// Aborted
			break
		} else {
			logger.Errorf("Error reading line: %s", err.Error())
		}
	}

	if f, err := os.Create(historyFileName); err != nil {
		logger.Errorf("Error writing history file %s: %s", historyFileName, err.Error())
	} else {
		line.WriteHistory(f)
		f.Close()
	}
}
