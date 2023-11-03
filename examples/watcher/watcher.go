package main

import (
	"context"
	"flag"
	"github.com/acquirecloud/golibs/cast"
	context2 "github.com/acquirecloud/golibs/context"
	"github.com/acquirecloud/golibs/errors"
	"github.com/acquirecloud/golibs/files"
	"github.com/acquirecloud/golibs/logging"
	"github.com/simila-io/simila/api/gen/index/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

var (
	addr = flag.String("addr", "localhost:50051", "the address to connect to")
	path = flag.String("path", "", "path to the directory to scan .TXT and .PDF files")

	logger logging.Logger
)

func main() {
	logger = logging.NewLogger("watcher")
	flag.Parse()
	watchPath := cast.String(path, "")
	if watchPath == "" {
		logger.Infof("path to the directory to watch is expected.")
		return
	}

	d, err := os.Open(watchPath)
	if err != nil {
		if os.IsNotExist(err) {
			logger.Errorf("the path=%s seems not exist", watchPath)
			return
		}
		logger.Errorf("could not access to the path=%s: %s", watchPath, err.Error())
		return
	}
	d.Close()

	if err := files.EnsureDirExists(watchPath); err != nil {
		logger.Errorf("the path=%s seems doesn't exist: %s", err.Error())
		return
	}

	conn, err := grpc.Dial(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Errorf("did not connect: %v", err)
		return
	}
	defer conn.Close()
	sc := index.NewServiceClient(conn)
	mainCtx := context2.NewSignalsContext(os.Interrupt, syscall.SIGTERM)
	go serve(mainCtx, watchPath, sc)
	<-mainCtx.Done()
}

func serve(ctx context.Context, path string, sc index.ServiceClient) {
	logger.Infof("start scanning %s", path)
	defer logger.Infof("end scanning %s", path)
	curFiles := map[string]os.FileInfo{}
	for ctx.Err() == nil {
		seen := map[string]bool{}
		fis := files.ListDir(path)
		for _, fi := range fis {
			if fi.IsDir() || !strings.HasSuffix(fi.Name(), ".txt") {
				continue
			}
			cf, ok := curFiles[fi.Name()]
			seen[fi.Name()] = true
			if ok && cf.Size() == fi.Size() {
				continue
			}
			if err := uploadfile(ctx, path, fi, sc); err != nil {
				continue
			}
			curFiles[fi.Name()] = fi
			if !ok {
				logger.Infof("found the new file %s and uploaded successfully", fi.Name())
			} else {
				logger.Infof("the file %s was updated successfully", fi.Name())
			}
		}
		for fn := range curFiles {
			if _, ok := seen[fn]; !ok {
				logger.Infof("the file %s seems to be removed from the folder ", fn)
				if _, err := sc.Delete(ctx, &index.Id{Id: fn}); err != nil && !errors.Is(err, errors.ErrNotExist) {
					logger.Errorf("coud not delete the index %s: %s", fn, err)
					continue
				}
				delete(curFiles, fn)
			}
		}
		context2.Sleep(ctx, time.Second)
	}
}

func uploadfile(ctx context.Context, dir string, fi os.FileInfo, sc index.ServiceClient) error {
	fn := fi.Name()
	ext := fn[len(fn)-3:]
	f, err := os.Open(filepath.Join(dir, fn))
	if err != nil {
		logger.Errorf("coud not open file for read %s: %s", fn, err.Error())
		return err
	}
	defer f.Close()

	buf := make([]byte, fi.Size())
	_, err = f.Read(buf)
	if err != nil {
		logger.Errorf("could not read file %s: %s", fn, err.Error())
		return err
	}
	for {
		_, err = sc.Create(ctx, &index.CreateIndexRequest{Id: fn, Format: ext, Document: buf})
		if err != nil {
			if errors.Is(err, errors.ErrExist) {
				logger.Infof("the index with id=%s already exists, let's delete it and rescan ", fn)
				_, err = sc.Delete(ctx, &index.Id{Id: fn})
				if err != nil {
					logger.Errorf("could not delete the index with id=%s: %s", fn, err)
				} else {
					continue
				}
			}
			logger.Errorf("could not create the new index for file=%s: %s", fn, err.Error())
		}
		break
	}
	return nil
}
