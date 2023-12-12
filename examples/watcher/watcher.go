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
	curFiles := map[string]int64{}
	for ctx.Err() == nil {
		seen := map[string]bool{}
		fis := scanTxtFiles(path)
		for fn, sz := range fis {
			seen[fn] = true
			curSz, ok := curFiles[fn]
			if ok && curSz == sz {
				continue
			}
			if err := uploadfile(ctx, path, fn, sc); err != nil {
				logger.Errorf("could not upload file %s to the server: %s", fn, err)
				continue
			}
			curFiles[fn] = sz
			if !ok {
				logger.Infof("found the new file %s and uploaded successfully", fn)
			} else {
				logger.Infof("the file %s was updated successfully", fn)
			}
		}
		for fn := range curFiles {
			if _, ok := seen[fn]; !ok {
				logger.Infof("the file %s seems to be removed from the folder ", fn)
				if _, err := sc.DeleteNode(ctx, &index.Path{Path: fn}); err != nil && !errors.Is(err, errors.ErrNotExist) {
					logger.Errorf("coud not delete the index %s: %s", fn, err)
					continue
				}
				delete(curFiles, fn)
			}
		}
		context2.Sleep(ctx, time.Second)
	}
}

func scanTxtFiles(path string) map[string]int64 {
	res := make(map[string]int64)
	fis := files.ListDir(path)
	for _, fi := range fis {
		if fi.IsDir() {
			sf := scanTxtFiles(filepath.Join(path, fi.Name()))
			for n, s := range sf {
				res[filepath.Join(fi.Name(), n)] = s
			}
			continue
		}
		if !strings.HasSuffix(fi.Name(), ".txt") {
			continue
		}
		res[fi.Name()] = fi.Size()
	}
	return res
}

func uploadfile(ctx context.Context, dir string, fn string, sc index.ServiceClient) error {
	for {
		err := uploadfile2(ctx, dir, fn, sc)
		if errors.Is(err, errors.ErrExist) {
			logger.Infof("the index with id=%s already exists, let's delete it and rescan ", fn)
			if _, err = sc.DeleteNode(ctx, &index.Path{Path: fn}); err != nil {
				return err
			}
			continue
		}
		return err
	}
}

func uploadfile2(ctx context.Context, dir string, fn string, sc index.ServiceClient) error {
	ext := fn[len(fn)-3:]
	f, err := os.Open(filepath.Join(dir, fn))
	if err != nil {
		logger.Errorf("coud not open file for read %s: %s", fn, err.Error())
		return err
	}
	defer f.Close()

	crr := &index.CreateRecordsRequest{Path: fn, Parser: cast.Ptr(ext)}
	buf := make([]byte, 4096)
	stream, err := sc.CreateWithStreamData(ctx)
	if err != nil {
		logger.Errorf("coud not make call to the server: %s", err.Error())
		return err
	}
	for {
		n, err := f.Read(buf)
		if err != nil {
			break
		}
		err = stream.Send(&index.CreateIndexStreamRequest{Meta: crr, Data: buf[:n]})
		if err != nil {
			break
		}
		crr = nil
	}
	_, err = stream.CloseAndRecv()
	return err
}
