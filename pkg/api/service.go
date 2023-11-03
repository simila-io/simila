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
	"context"
	"fmt"
	"github.com/acquirecloud/golibs/errors"
	"github.com/acquirecloud/golibs/logging"
	"github.com/simila-io/simila/api/gen/format/v1"
	"github.com/simila-io/simila/api/gen/index/v1"
	"github.com/simila-io/simila/pkg/indexer/persistence"
	"github.com/simila-io/simila/pkg/parser"
	"google.golang.org/protobuf/types/known/emptypb"
	"io"
)

// Service implements the gRPC API endpoints
type (
	Service struct {
		PProvider parser.Provider `inject:""`
		Db        persistence.Db  `inject:""`

		idxService idxService
		fmtService fmtService
		logger     logging.Logger
	}

	idxService struct {
		index.UnsafeServiceServer
		s *Service
	}

	fmtService struct {
		format.UnimplementedServiceServer
		s *Service
	}
)

var _ index.ServiceServer = idxService{}
var _ format.ServiceServer = fmtService{}

func NewService() *Service {
	s := &Service{logger: logging.NewLogger("api.Service")}
	s.idxService = idxService{s: s}
	s.fmtService = fmtService{s: s}
	return s
}

// IndexServiceServer returns index.ServiceServer
func (s *Service) IndexServiceServer() index.ServiceServer {
	return s.idxService
}

func (s *Service) FormatServiceServer() format.ServiceServer {
	return s.fmtService
}

// createIndex allows to create a new index. The body represents a file stream, if presents. body may be nil, then
// the body may be taken from the request.
func (s *Service) createIndex(ctx context.Context, request *index.CreateIndexRequest, body io.Reader) (*index.Index, error) {
	s.logger.Infof("createIndex(): request=%s, body?=%t", request, body != nil)
	var p parser.Parser
	if body != nil {
		p = s.PProvider.Parser(request.Format)
		if p == nil {
			return &index.Index{}, fmt.Errorf("the format %s is not supported: %w", request.Format, errors.ErrInvalid)
		}
	}

	mtx := s.Db.NewModelTx(ctx)
	defer mtx.Commit()

	//TODO: no tags yet
	_, err := mtx.CreateIndex(persistence.Index{ID: request.Id, Format: request.Format})
	if err != nil {
		mtx.Rollback()
		return nil, fmt.Errorf("could not create new index with ID=%s: %w", request.Id, err)
	}

	return &index.Index{}, nil
}

func (s *Service) deleteIndex(ctx context.Context, id *index.Id) (*emptypb.Empty, error) {
	s.logger.Infof("deleteIndex(): id=%s", id.GetId())
	return &emptypb.Empty{}, errors.ErrNotExist
}

func (s *Service) getIndex(ctx context.Context, id *index.Id) (*index.Index, error) {
	s.logger.Debugf("getIndex(): id=%s", id.GetId())
	return &index.Index{}, errors.ErrNotExist
}

func (s *Service) putIndex(ctx context.Context, idx *index.Index) (*index.Index, error) {
	s.logger.Infof("putIndex(): id=%s", idx.Id)
	return &index.Index{}, errors.ErrNotExist
}

func (s *Service) listIndexes(ctx context.Context, request *index.ListRequest) (*index.Indexes, error) {
	s.logger.Debugf("listIndexes(): %s", request)
	return &index.Indexes{}, errors.ErrNotExist
}

func (s *Service) patchIndexRecords(ctx context.Context, request *index.PatchRecordsRequest) (*index.PatchRecordsResult, error) {
	s.logger.Debugf("listIndexes(): %s", request)
	return &index.PatchRecordsResult{}, errors.ErrNotExist
}

func (s *Service) listIndexRecords(ctx context.Context, request *index.ListRecordsRequest) (*index.ListRecordsResult, error) {
	s.logger.Debugf("listIndexRecords(): %s", request)
	return &index.ListRecordsResult{}, errors.ErrNotExist
}

func (s *Service) searchRecords(ctx context.Context, request *index.SearchRecordsRequest) (*index.SearchRecordsResult, error) {
	s.logger.Debugf("searchRecords(): %s", request)
	return &index.SearchRecordsResult{}, errors.ErrNotExist
}

func (s *Service) createFormat(ctx context.Context, req *format.Format) (*format.Format, error) {
	s.logger.Infof("createFormat(): request=%s", req)
	if req == nil {
		return &format.Format{}, errors.ErrInvalid
	}
	mtx := s.Db.NewModelTx(ctx)
	if _, err := mtx.CreateFormat(persistence.Format{Name: req.Name}); err != nil {
		return &format.Format{}, err
	}
	return &format.Format{Name: req.Name}, nil
}

func (s *Service) getFormat(ctx context.Context, id *format.Id) (*format.Format, error) {
	s.logger.Debugf("getFormat(): id=%s", id)
	if id == nil {
		return &format.Format{}, errors.ErrInvalid
	}
	mtx := s.Db.NewModelTx(ctx)
	frmt, err := mtx.GetFormat((*id).Id)
	if err != nil {
		return &format.Format{}, err
	}
	return &format.Format{Name: frmt.Name}, nil
}

func (s *Service) deleteFormat(ctx context.Context, id *format.Id) (*emptypb.Empty, error) {
	s.logger.Infof("deleteFormat(): id=%s", id)
	if id == nil {
		return &emptypb.Empty{}, errors.ErrInvalid
	}
	mtx := s.Db.NewModelTx(ctx)
	if err := mtx.DeleteFormat((*id).Id); err != nil {
		return &emptypb.Empty{}, err
	}
	return &emptypb.Empty{}, nil
}

func (s *Service) listFormat(ctx context.Context, empty *emptypb.Empty) (*format.Formats, error) {
	s.logger.Debugf("listFormat()")
	mtx := s.Db.NewModelTx(ctx)
	mFrmts, err := mtx.ListFormats()
	if err != nil {
		return &format.Formats{}, err
	}
	aFrmts := make([]*format.Format, len(mFrmts))
	for i := 0; i < len(mFrmts); i++ {
		aFrmts[i] = &format.Format{Name: mFrmts[i].Name}
	}
	return &format.Formats{Formats: aFrmts}, nil
}

// -------------------------- index.Service ---------------------------
func (ids idxService) Create(ctx context.Context, request *index.CreateIndexRequest) (*index.Index, error) {
	return ids.s.createIndex(ctx, request, nil)
}

func (ids idxService) Delete(ctx context.Context, id *index.Id) (*emptypb.Empty, error) {
	return ids.s.deleteIndex(ctx, id)
}

func (ids idxService) Get(ctx context.Context, id *index.Id) (*index.Index, error) {
	return ids.s.getIndex(ctx, id)
}

func (ids idxService) Put(ctx context.Context, idx *index.Index) (*index.Index, error) {
	return ids.s.putIndex(ctx, idx)
}

func (ids idxService) List(ctx context.Context, request *index.ListRequest) (*index.Indexes, error) {
	return ids.s.listIndexes(ctx, request)
}

func (ids idxService) PatchRecords(ctx context.Context, request *index.PatchRecordsRequest) (*index.PatchRecordsResult, error) {
	return ids.s.patchIndexRecords(ctx, request)
}

func (ids idxService) ListRecords(ctx context.Context, request *index.ListRecordsRequest) (*index.ListRecordsResult, error) {
	return ids.s.listIndexRecords(ctx, request)
}

func (ids idxService) SearchRecords(ctx context.Context, request *index.SearchRecordsRequest) (*index.SearchRecordsResult, error) {
	return ids.s.searchRecords(ctx, request)
}

// ----------------------------- format.Service ---------------------------------
func (fs fmtService) Create(ctx context.Context, f *format.Format) (*format.Format, error) {
	return fs.s.createFormat(ctx, f)
}

func (fs fmtService) Get(ctx context.Context, id *format.Id) (*format.Format, error) {
	return fs.s.getFormat(ctx, id)
}

func (fs fmtService) Delete(ctx context.Context, id *format.Id) (*emptypb.Empty, error) {
	return fs.s.deleteFormat(ctx, id)
}

func (fs fmtService) List(ctx context.Context, empty *emptypb.Empty) (*format.Formats, error) {
	return fs.s.listFormat(ctx, empty)
}
