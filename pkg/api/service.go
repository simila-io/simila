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
	"bytes"
	"context"
	"fmt"
	"github.com/acquirecloud/golibs/cast"
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

// createIndex allows to create a new index. The body represents a file stream,
// if presents. body may be nil, then the body may be taken from the request.
func (s *Service) createIndex(ctx context.Context, request *index.CreateIndexRequest, body io.Reader) (*index.Index, error) {
	s.logger.Infof("createIndex(): id=%s, format=%s, tags=%v records=%d, documentSize=%d, body?=%t", request.Id,
		request.Format, request.Tags, len(request.Records), len(request.Document), body != nil)
	if request == nil {
		return &index.Index{}, errors.GRPCWrap(errors.ErrInvalid)
	}

	var p parser.Parser
	if body == nil && len(request.Document) > 0 {
		body = bytes.NewReader(request.Document)
		s.logger.Infof("createIndex(): will use Document for reading records")
	}
	if body != nil {
		p = s.PProvider.Parser(request.Format)
		if p == nil {
			return &index.Index{}, errors.GRPCWrap(fmt.Errorf("the format %s is not supported: %w", request.Format, errors.ErrInvalid))
		}
	}

	mtx := s.Db.NewModelTx(ctx)
	defer func() {
		_ = mtx.Rollback()
	}()
	idx, err := mtx.CreateIndex(toModelIndexFromApiCreateIdxReq(request))
	if err != nil {
		return nil, errors.GRPCWrap(fmt.Errorf("could not create new index with ID=%s: %w", request.Id, err))
	}
	if p != nil {
		count, err := p.ScanRecords(ctx, mtx, request.Id, body)
		if err != nil {
			return nil, errors.GRPCWrap(fmt.Errorf("could not read records for %s format: %w", request.Format, err))
		}
		s.logger.Infof("createIndex(): read %d records by parser %s for the new index %s", count, p, request.Id)
	} else {
		if err = mtx.UpsertIndexRecords(toModelIndexRecordsFromApiRecords(idx.ID, request.Records)...); err != nil {
			return nil, errors.GRPCWrap(fmt.Errorf("could not create the index records: %w", err))
		}
	}
	_ = mtx.Commit()
	return toApiIndex(idx), nil
}

func (s *Service) deleteIndex(ctx context.Context, id *index.Id) (*emptypb.Empty, error) {
	s.logger.Infof("deleteIndex(): id=%s", id.GetId())
	if id == nil {
		return &emptypb.Empty{}, errors.GRPCWrap(errors.ErrInvalid)
	}
	mtx := s.Db.NewModelTx(ctx)
	if err := mtx.DeleteIndex((*id).Id); err != nil {
		return &emptypb.Empty{}, errors.GRPCWrap(err)
	}
	return &emptypb.Empty{}, nil
}

func (s *Service) getIndex(ctx context.Context, id *index.Id) (*index.Index, error) {
	s.logger.Debugf("getIndex(): id=%s", id.GetId())
	if id == nil {
		return &index.Index{}, errors.GRPCWrap(errors.ErrInvalid)
	}
	mtx := s.Db.NewModelTx(ctx)
	idx, err := mtx.GetIndex((*id).Id)
	if err != nil {
		return &index.Index{}, errors.GRPCWrap(err)
	}
	return toApiIndex(idx), nil
}

func (s *Service) putIndex(ctx context.Context, idx *index.Index) (*index.Index, error) {
	s.logger.Infof("putIndex(): id=%s", idx.Id)
	if idx == nil {
		return &index.Index{}, errors.GRPCWrap(errors.ErrInvalid)
	}
	mtx := s.Db.NewModelTx(ctx)
	mtx.MustBegin()
	defer func() {
		_ = mtx.Rollback()
	}()
	if err := mtx.UpdateIndex(toModelIndex(idx)); err != nil {
		return &index.Index{}, errors.GRPCWrap(err)
	}
	mIdx, err := mtx.GetIndex(idx.Id)
	if err != nil {
		return &index.Index{}, errors.GRPCWrap(err)
	}
	_ = mtx.Commit()
	return toApiIndex(mIdx), nil
}

func (s *Service) listIndexes(ctx context.Context, request *index.ListRequest) (*index.Indexes, error) {
	s.logger.Debugf("listIndexes(): %s", request)
	if request == nil {
		return &index.Indexes{}, errors.GRPCWrap(errors.ErrInvalid)
	}
	qry := persistence.IndexQuery{
		Format: cast.String(request.Format, ""),
		Tags:   request.Tags,
		FromID: request.StartIndexId,
		Limit:  int(cast.Int64(request.Limit, 0)),
	}
	if request.CreatedAfter != nil {
		qry.CreatedAfter = request.CreatedAfter.AsTime()
	}
	if request.CreatedBefore != nil {
		qry.CreatedBefore = request.CreatedBefore.AsTime()
	}
	mtx := s.Db.NewModelTx(ctx)
	mIdxs, err := mtx.QueryIndexes(qry)
	if err != nil {
		return &index.Indexes{}, errors.GRPCWrap(err)
	}
	aIdxs := make([]*index.Index, len(mIdxs.Items))
	for i := 0; i < len(mIdxs.Items); i++ {
		aIdxs[i] = toApiIndex(mIdxs.Items[i])
	}
	res := &index.Indexes{Total: mIdxs.Total}
	if len(aIdxs) > 0 {
		res.Indexes = aIdxs
	}
	if len(mIdxs.NextID) > 0 {
		res.NextIndexId = &mIdxs.NextID
	}
	return res, nil
}

func (s *Service) patchIndexRecords(ctx context.Context, request *index.PatchRecordsRequest) (*index.PatchRecordsResult, error) {
	s.logger.Debugf("patchIndexRecords(): %s", request)
	if request == nil {
		return &index.PatchRecordsResult{}, errors.GRPCWrap(errors.ErrInvalid)
	}
	mtx := s.Db.NewModelTx(ctx)
	mtx.MustBegin()
	defer func() {
		_ = mtx.Rollback()
	}()

	addRecs := toModelIndexRecordsFromApiRecords(request.Id, request.UpsertRecords)
	delRecs := toModelIndexRecordsFromApiRecords(request.Id, request.DeleteRecords)

	err := mtx.UpsertIndexRecords(addRecs...)
	if err != nil {
		return &index.PatchRecordsResult{}, errors.GRPCWrap(err)
	}
	nDel, err := mtx.DeleteIndexRecords(delRecs...)
	if err != nil {
		return &index.PatchRecordsResult{}, errors.GRPCWrap(err)
	}
	_ = mtx.Commit()
	return &index.PatchRecordsResult{Upserted: int64(len(addRecs)), Deleted: int64(nDel)}, nil
}

func (s *Service) listIndexRecords(ctx context.Context, request *index.ListRecordsRequest) (*index.ListRecordsResult, error) {
	s.logger.Debugf("listIndexRecords(): %s", request)
	if request == nil {
		return &index.ListRecordsResult{}, errors.GRPCWrap(errors.ErrInvalid)
	}
	mtx := s.Db.NewModelTx(ctx)
	qry := persistence.IndexRecordQuery{
		FromID: cast.String(request.StartRecordId, ""),
		Limit:  int(cast.Int64(request.Limit, 100)),
	}
	if len(request.Id) > 0 {
		qry.IndexIDs = []string{request.Id}
	}
	mRecs, err := mtx.QueryIndexRecords(qry)
	if err != nil {
		return &index.ListRecordsResult{}, errors.GRPCWrap(err)
	}
	aRecs := make([]*index.Record, len(mRecs.Items))
	for i := 0; i < len(mRecs.Items); i++ {
		aRecs[i] = toApiRecord(mRecs.Items[i])
	}
	res := &index.ListRecordsResult{
		Total: mRecs.Total,
	}
	if len(aRecs) > 0 {
		res.Records = aRecs
	}
	if len(mRecs.NextID) > 0 {
		res.NextRecordId = &mRecs.NextID
	}
	return res, nil
}

func (s *Service) searchRecords(ctx context.Context, request *index.SearchRecordsRequest) (*index.SearchRecordsResult, error) {
	s.logger.Debugf("searchRecords(): %s", request)
	if request == nil {
		return &index.SearchRecordsResult{}, errors.GRPCWrap(errors.ErrInvalid)
	}
	mtx := s.Db.NewModelTx(ctx)
	qry := persistence.SearchQuery{
		IndexIDs: request.IndexIDs,
		Query:    request.Text,
		Tags:     request.Tags,
		Distinct: cast.Bool(request.Distinct, false),
		FromID:   cast.String(request.PageId, ""),
		Limit:    int(cast.Int64(request.Limit, 100)),
	}
	mRecs, err := mtx.Search(qry)
	if err != nil {
		return &index.SearchRecordsResult{}, errors.GRPCWrap(err)
	}
	aRecs := make([]*index.IndexRecord, len(mRecs.Items))
	for i := 0; i < len(mRecs.Items); i++ {
		aRecs[i] = toApiIndexRecord(mRecs.Items[i])
	}
	res := &index.SearchRecordsResult{
		Total: mRecs.Total,
	}
	if len(aRecs) > 0 {
		res.Records = aRecs
	}
	if len(mRecs.NextID) > 0 {
		res.NextPageId = &mRecs.NextID
	}
	return res, nil
}

func (s *Service) createFormat(ctx context.Context, req *format.Format) (*format.Format, error) {
	s.logger.Infof("createFormat(): request=%s", req)
	if req == nil {
		return &format.Format{}, errors.GRPCWrap(errors.ErrInvalid)
	}
	mtx := s.Db.NewModelTx(ctx)
	frmt, err := mtx.CreateFormat(toModelFormat(req))
	if err != nil {
		return &format.Format{}, errors.GRPCWrap(err)
	}
	return toApiFormat(frmt), nil
}

func (s *Service) getFormat(ctx context.Context, id *format.Id) (*format.Format, error) {
	s.logger.Debugf("getFormat(): id=%s", id)
	if id == nil {
		return &format.Format{}, errors.GRPCWrap(errors.ErrInvalid)
	}
	mtx := s.Db.NewModelTx(ctx)
	frmt, err := mtx.GetFormat((*id).Id)
	if err != nil {
		return &format.Format{}, errors.GRPCWrap(err)
	}
	return toApiFormat(frmt), nil
}

func (s *Service) deleteFormat(ctx context.Context, id *format.Id) (*emptypb.Empty, error) {
	s.logger.Infof("deleteFormat(): id=%s", id)
	if id == nil {
		return &emptypb.Empty{}, errors.GRPCWrap(errors.ErrInvalid)
	}
	mtx := s.Db.NewModelTx(ctx)
	if err := mtx.DeleteFormat((*id).Id); err != nil {
		return &emptypb.Empty{}, errors.GRPCWrap(err)
	}
	return &emptypb.Empty{}, nil
}

func (s *Service) listFormat(ctx context.Context, _ *emptypb.Empty) (*format.Formats, error) {
	s.logger.Debugf("listFormat()")
	mtx := s.Db.NewModelTx(ctx)
	mFrmts, err := mtx.ListFormats()
	if err != nil {
		return &format.Formats{}, errors.GRPCWrap(err)
	}
	aFrmts := make([]*format.Format, len(mFrmts))
	for i := 0; i < len(mFrmts); i++ {
		aFrmts[i] = toApiFormat(mFrmts[i])
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
