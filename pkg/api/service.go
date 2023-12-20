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

// Service implements the gRPC API endpoints v1
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

// createRecords allows to create a new index. The body represents a file stream,
// if presents, body may be nil, then the body may be taken from the request.
func (s *Service) createRecords(ctx context.Context, request *index.CreateRecordsRequest, body io.Reader) (*index.CreateRecordsResult, error) {
	if request == nil {
		return &index.CreateRecordsResult{}, errors.GRPCWrap(fmt.Errorf("invalid nil request: %w", errors.ErrInvalid))
	}

	s.logger.Infof("createIndexRecords(): path=%s, nodeType=%s, tags=%v, parser=%s, rankMultiplier=%f, records=%d, document=%d, body=%t", request.Path,
		cast.Value(request.NodeType, index.NodeType(0)), request.Tags, cast.Value(request.Parser, "N/A"), request.RankMultiplier,
		len(request.Records), len(request.Document), body != nil)

	var p parser.Parser
	if body == nil && len(request.Document) > 0 {
		body = bytes.NewReader(request.Document)
		s.logger.Infof("createIndexRecords(): will use Document for reading records")
	}

	if body != nil {
		parser := cast.Value(request.Parser, "")
		p = s.PProvider.Parser(parser)
		if p == nil {
			return &index.CreateRecordsResult{}, errors.GRPCWrap(fmt.Errorf("the parser=%q is not supported: %w", parser, errors.ErrInvalid))
		}
	}

	mtx := s.Db.NewModelTx(ctx)
	mtx.MustBegin()
	defer func() {
		_ = mtx.Rollback()
	}()
	res := &index.CreateRecordsResult{}

	pths := persistence.SplitPath(request.Path)
	if len(pths) == 0 {
		return &index.CreateRecordsResult{}, errors.GRPCWrap(fmt.Errorf("the path=%q should not be empty: %w", request.Path, errors.ErrInvalid))
	}
	nodes, err := mtx.ListNodes(request.Path)
	if err != nil {
		return &index.CreateRecordsResult{}, errors.GRPCWrap(err)
	}
	n2c := nodes2Create(pths, nodes, request.Tags, cast.Value(request.NodeType, index.NodeType_FOLDER))
	if len(n2c) > 0 {
		nodes, err = mtx.CreateNodes(n2c...)
		if err != nil {
			return &index.CreateRecordsResult{}, errors.GRPCWrap(err)
		}
		res.NodesCreated = &index.Nodes{
			Nodes: toApiNodes(nodes),
		}
	}
	node := nodes[len(nodes)-1]
	if p != nil {
		count, err := p.ScanRecords(ctx, mtx, node.ID, body)
		if err != nil {
			return nil, errors.GRPCWrap(fmt.Errorf("could not read records for %q format: %w", cast.String(request.Parser, ""), err))
		}
		s.logger.Infof("createRecords(): read %d records by parser %s for the node %q(%d)", count, p, persistence.ConcatPath(node.Path, node.Name), node.ID)
	} else {
		rc, err := mtx.UpsertIndexRecords(toModelIndexRecordsFromApiRecords(node.ID, request.Records, 1.0)...)
		if err != nil {
			return &index.CreateRecordsResult{}, errors.GRPCWrap(err)
		}
		res.RecordsCreated = rc
	}
	_ = mtx.Commit()
	return res, nil
}

func nodes2Create(pths []string, nodes []persistence.Node, tags persistence.Tags, lastNodeType index.NodeType) []persistence.Node {
	nodes2Create := []persistence.Node{}
	if len(nodes) < len(pths) {
		p := "/"
		if len(nodes) > 0 {
			n := nodes[len(nodes)-1]
			p = persistence.ConcatPath(n.Path, n.Name)
		}
		for i := len(nodes); i < len(pths); i++ {
			nodes2Create = append(nodes2Create, persistence.Node{Path: p, Name: pths[i]})
			p = persistence.ConcatPath(p, pths[i])
		}
		nodes2Create[len(nodes2Create)-1].Tags = tags
		if lastNodeType == index.NodeType_DOCUMENT {
			nodes2Create[len(nodes2Create)-1].Flags = persistence.NodeFlagDocument
		}
	}
	return nodes2Create
}

func (s *Service) updateNode(ctx context.Context, request *index.UpdateNodeRequest) (*emptypb.Empty, error) {
	res := &emptypb.Empty{}
	if request == nil {
		return res, errors.GRPCWrap(fmt.Errorf("invalid nil request: %w", errors.ErrInvalid))
	}
	tags := cast.Value(request.Node, index.Node{}).Tags
	s.logger.Infof("updateNode(): path=%q, tags=%v", request.Path, tags)

	mtx := s.Db.NewModelTx(ctx)
	defer func() {
		_ = mtx.Rollback()
	}()
	n, err := mtx.GetNode(request.Path)
	if err != nil {
		return res, errors.GRPCWrap(err)
	}
	if err = mtx.UpdateNode(persistence.Node{ID: n.ID, Tags: tags}); err != nil {
		return res, errors.GRPCWrap(err)
	}
	mtx.Commit()
	return res, nil
}

func (s *Service) deleteNodes(ctx context.Context, dnr *index.DeleteNodesRequest) (*emptypb.Empty, error) {
	force := cast.Value(dnr.Force, false)
	s.logger.Infof("deleteNodes(): filter=%q, force=%t", dnr.FilterConditions, force)
	mtx := s.Db.NewModelTx(ctx)
	mtx.MustBegin()
	defer func() {
		_ = mtx.Rollback()
	}()
	res := &emptypb.Empty{}
	err := mtx.DeleteNodes(persistence.DeleteNodesQuery{FilterConditions: dnr.FilterConditions, Force: force})
	if err != nil {
		return res, errors.GRPCWrap(err)
	}
	mtx.Commit()
	return res, nil
}

func (s *Service) listNodes(ctx context.Context, path *index.Path) (*index.Nodes, error) {
	mtx := s.Db.NewModelTx(ctx)
	res := &index.Nodes{}
	nodes, err := mtx.ListChildren(path.Path)
	if err != nil {
		return res, errors.GRPCWrap(err)
	}
	res.Nodes = toApiNodes(nodes)
	return res, nil
}

func (s *Service) listRecords(ctx context.Context, request *index.ListRequest) (*index.ListRecordsResult, error) {
	mtx := s.Db.NewModelTx(ctx)
	res := &index.ListRecordsResult{}
	q := persistence.IndexRecordQuery{
		Format:        cast.Value(request.Format, ""),
		CreatedAfter:  protoTime2Time(request.CreatedAfter),
		CreatedBefore: protoTime2Time(request.CreatedBefore),
		FromID:        cast.Value(request.PageId, ""),
	}
	q.Limit = int(cast.Value(request.Limit, 100))
	if q.Limit < 1 || q.Limit > 1000 {
		q.Limit = 1000
	}
	node, err := mtx.GetNode(request.Path)
	if err != nil {
		return res, errors.GRPCWrap(err)
	}
	q.NodeID = node.ID
	qr, err := mtx.QueryIndexRecords(q)
	if err != nil {
		return res, errors.GRPCWrap(err)
	}
	res.Total = qr.Total
	if qr.NextID != "" {
		res.NextPageId = &qr.NextID
	}
	res.Records = toApiRecords(qr.Items)
	return res, nil
}

func (s *Service) search(ctx context.Context, request *index.SearchRecordsRequest) (*index.SearchRecordsResult, error) {
	mtx := s.Db.NewModelTx(ctx)
	mtx.MustBegin()
	defer func() {
		_ = mtx.Rollback()
	}()
	res := &index.SearchRecordsResult{}
	q := persistence.SearchQuery{
		TextQuery:        request.TextQuery,
		FilterConditions: request.FilterConditions,
		GroupByPathOff:   cast.Value(request.GroupByPathOff, false),
		Offset:           int(cast.Value(request.Offset, 0)),
		Limit:            int(cast.Value(request.Limit, 0)),
	}
	if q.Limit < 1 || q.Limit > 1000 {
		q.Limit = 1000
	}
	qr, err := mtx.Search(q)
	if err != nil {
		return res, errors.GRPCWrap(err)
	}
	res.Total = qr.Total
	res.Items = toApiSearchRecords(qr.Items)
	_ = mtx.Commit()
	return res, nil
}

func (s *Service) patchIndexRecords(ctx context.Context, request *index.PatchRecordsRequest) (*index.PatchRecordsResult, error) {
	s.logger.Debugf("patchIndexRecords(): %s", request)
	if request == nil {
		return &index.PatchRecordsResult{}, errors.GRPCWrap(fmt.Errorf("invalid nil request: %w", errors.ErrInvalid))
	}
	mtx := s.Db.NewModelTx(ctx)
	mtx.MustBegin()
	defer func() {
		_ = mtx.Rollback()
	}()
	res := &index.PatchRecordsResult{}
	node, err := mtx.GetNode(request.Path)
	if err != nil {
		return res, errors.GRPCWrap(err)
	}

	addRecs := toModelIndexRecordsFromApiRecords(node.ID, request.UpsertRecords, 1.0)
	delRecs := toModelIndexRecordsFromApiRecords(node.ID, request.DeleteRecords, 1.0)

	n, err := mtx.UpsertIndexRecords(addRecs...)
	if err != nil {
		return res, errors.GRPCWrap(fmt.Errorf("index records patch(upsert) failed: %w", err))
	}
	res.Upserted = n
	n, err = mtx.DeleteIndexRecords(delRecs...)
	if err != nil && !errors.Is(err, errors.ErrNotExist) {
		return res, errors.GRPCWrap(fmt.Errorf("index records patch(delete) failed: %w", err))
	}
	res.Deleted = n
	_ = mtx.Commit()
	return res, nil
}

func (s *Service) createFormat(ctx context.Context, req *format.Format) (*format.Format, error) {
	s.logger.Infof("createFormat(): request=%s", req)
	if req == nil {
		return &format.Format{}, errors.GRPCWrap(fmt.Errorf("invalid nil request: %w", errors.ErrInvalid))
	}
	mtx := s.Db.NewModelTx(ctx)
	frmt, err := mtx.CreateFormat(toModelFormat(req))
	if err != nil {
		return &format.Format{}, errors.GRPCWrap(fmt.Errorf("could not create format=%v: %w", req, err))
	}
	return toApiFormat(frmt), nil
}

func (s *Service) getFormat(ctx context.Context, id *format.Id) (*format.Format, error) {
	s.logger.Debugf("getFormat(): id=%s", id)
	if id == nil {
		return &format.Format{}, errors.GRPCWrap(fmt.Errorf("invalid nil request: %w", errors.ErrInvalid))
	}
	mtx := s.Db.NewModelTx(ctx)
	frmt, err := mtx.GetFormat((*id).Id)
	if err != nil {
		return &format.Format{}, errors.GRPCWrap(fmt.Errorf("could not get format with ID=%v: %w", (*id).Id, err))
	}
	return toApiFormat(frmt), nil
}

func (s *Service) deleteFormat(ctx context.Context, id *format.Id) (*emptypb.Empty, error) {
	s.logger.Infof("deleteFormat(): id=%s", id)
	if id == nil {
		return &emptypb.Empty{}, errors.GRPCWrap(fmt.Errorf("invalid nil request: %w", errors.ErrInvalid))
	}
	mtx := s.Db.NewModelTx(ctx)
	if err := mtx.DeleteFormat((*id).Id); err != nil {
		return &emptypb.Empty{}, errors.GRPCWrap(fmt.Errorf("could not delete format with ID=%v: %w", (*id).Id, err))
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

func (ids idxService) Create(ctx context.Context, request *index.CreateRecordsRequest) (*index.CreateRecordsResult, error) {
	return ids.s.createRecords(ctx, request, nil)
}

func (ids idxService) UpdateNode(ctx context.Context, request *index.UpdateNodeRequest) (*emptypb.Empty, error) {
	return ids.s.updateNode(ctx, request)
}

func (ids idxService) DeleteNodes(ctx context.Context, dnr *index.DeleteNodesRequest) (*emptypb.Empty, error) {
	return ids.s.deleteNodes(ctx, dnr)
}

func (ids idxService) ListNodes(ctx context.Context, path *index.Path) (*index.Nodes, error) {
	return ids.s.listNodes(ctx, path)
}

func (ids idxService) ListRecords(ctx context.Context, request *index.ListRequest) (*index.ListRecordsResult, error) {
	return ids.s.listRecords(ctx, request)
}

func (ids idxService) Search(ctx context.Context, request *index.SearchRecordsRequest) (*index.SearchRecordsResult, error) {
	return ids.s.search(ctx, request)
}

func (ids idxService) PatchRecords(ctx context.Context, request *index.PatchRecordsRequest) (*index.PatchRecordsResult, error) {
	return ids.s.patchIndexRecords(ctx, request)
}

func (ids idxService) CreateWithStreamData(server index.Service_CreateWithStreamDataServer) error {
	req, err := server.Recv()
	if err != nil {
		return err
	}
	if req.Meta == nil {
		return errors.GRPCWrap(fmt.Errorf("CreateWithStreamData(): first packet of the grpc stream must contain meta: %w", errors.ErrInvalid))
	}
	r, w := io.Pipe()
	defer r.Close()
	buf := req.Data
	go func() {
		var err error
		defer func() {
			defer w.CloseWithError(err)
		}()
		for err == nil {
			if len(buf) > 0 {
				n := 0
				n, err = w.Write(buf)
				buf = buf[n:]
			} else {
				var req *index.CreateIndexStreamRequest
				req, err = server.Recv()
				if req != nil {
					buf = req.Data
				}
			}
		}
		if err != nil && err != io.EOF {
			ids.s.logger.Warnf("CreateWithStreamData(): could not write data: %s", err.Error())
		} else {
			err = nil
		}
	}()
	idx, err := ids.s.createRecords(server.Context(), req.Meta, r)
	if err == nil {
		server.SendAndClose(idx)
	}
	return err
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
