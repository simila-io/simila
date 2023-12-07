package api

import (
	"fmt"
	"github.com/acquirecloud/golibs/cast"
	"github.com/acquirecloud/golibs/errors"
	"github.com/acquirecloud/golibs/logging"
	"github.com/gin-gonic/gin"
	"github.com/simila-io/simila/api/gen/format/v1"
	"github.com/simila-io/simila/api/gen/index/v2"
	similapi "github.com/simila-io/simila/api/genpublic/v1"
	"net/http"
	"path"
	"strings"
)

type Rest struct {
	svc    *Service
	logger logging.Logger
}

type ErrorMsg struct {
	Error string `json:"error"`
}

var _ similapi.ServerInterface = (*Rest)(nil)

func NewRest(svc *Service) *Rest {
	return &Rest{svc: svc, logger: logging.NewLogger("api.rest")}
}

func (r *Rest) RegisterEPs(g *gin.Engine) error {
	similapi.RegisterHandlersWithOptions(g, r, similapi.GinServerOptions{BaseURL: "v1"})
	return nil
}

func (r *Rest) ListNodes(c *gin.Context, params similapi.ListNodesParams) {
	nodes, err := r.svc.IndexServiceServer().ListNodes(c, &index.Path{Path: cast.Value(params.Path, "")})
	if r.errorRespnse(c, err, "") {
		return
	}
	c.JSON(http.StatusOK, similapi.ListNodesResult{Items: nodes2Rest(nodes.Nodes)})
}

func (r *Rest) DeleteNode(c *gin.Context, path similapi.Path) {
	//TODO implement me
	panic("implement me")
}

func (r *Rest) ListNodeRecords(c *gin.Context, path similapi.Path, params similapi.ListNodeRecordsParams) {
	//TODO implement me
	panic("implement me")
}

func (r *Rest) PatchNodeRecords(c *gin.Context, path similapi.Path) {
	//TODO implement me
	panic("implement me")
}

func (r *Rest) CreateNodeRecords(c *gin.Context, path similapi.Path) {
	//TODO implement me
	panic("implement me")
}

func (r *Rest) Search(c *gin.Context) {
	var sr similapi.SearchRecordsRequest
	if r.errorRespnse(c, BindAppJson(c, &sr), "") {
		return
	}
	res, err := r.svc.IndexServiceServer().Search(c, searchRecordsRequest2Proto(sr))
	if r.errorRespnse(c, err, "") {
		return
	}
	c.JSON(http.StatusOK, searchRecordsResult2Rest(res))
}

func (r *Rest) ListFormats(c *gin.Context) {
	fmts, err := r.svc.FormatServiceServer().List(c, nil)
	if r.errorRespnse(c, err, "") {
		return
	}
	c.JSON(http.StatusOK, formats2Rest(fmts))
}

func (r *Rest) CreateFormat(c *gin.Context) {
	var f similapi.Format
	if r.errorRespnse(c, BindAppJson(c, &f), "") {
		return
	}
	f1, err := r.svc.FormatServiceServer().Create(c, &format.Format{Name: f.Name})
	if r.errorRespnse(c, err, "") {
		return
	}
	c.Header("Location", ComposeURI(c.Request, f1.Name))
	c.JSON(http.StatusCreated, format2Rest(f1))
}

func (r *Rest) DeleteFormat(c *gin.Context, formatId similapi.FormatId) {
	_, err := r.svc.FormatServiceServer().Delete(c, &format.Id{Id: formatId})
	if r.errorRespnse(c, err, "") {
		return
	}
	c.Status(http.StatusNoContent)
}

func (r *Rest) GetFormat(c *gin.Context, formatId similapi.FormatId) {
	f, err := r.svc.FormatServiceServer().Get(c, &format.Id{Id: formatId})
	if r.errorRespnse(c, err, "") {
		return
	}
	c.JSON(http.StatusOK, format2Rest(f))
}

//func (r *Rest) GetIndexes(c *gin.Context, params similapi.GetIndexesParams) {
//	req := &index.ListRequest{}
//	if params.CreatedAfter != nil {
//		req.CreatedAfter = timestamppb.New(*params.CreatedAfter)
//	}
//	if params.CreatedBefore != nil {
//		req.CreatedBefore = timestamppb.New(*params.CreatedBefore)
//	}
//	if params.Limit != nil {
//		req.Limit = cast.Ptr(int64(*params.Limit))
//	}
//	if params.StartIndexId != nil {
//		req.StartIndexId = *params.StartIndexId
//	}
//	if params.Tags != nil {
//		req.Tags = *params.Tags
//	}
//	if params.Format != nil {
//		req.Format = params.Format
//	}
//	idxs, err := r.svc.IndexServiceServer().List(c, req)
//	if r.errorRespnse(c, err, "") {
//		return
//	}
//	c.JSON(http.StatusOK, idxs)
//}

//func (r *Rest) CreateIndex(c *gin.Context) {
//	var cir similapi.CreateIndexRequest
//	var idx *index.Index
//	if err := BindAppJson(c, &cir); err == nil {
//		r.logger.Infof("creating new index from the json object %s", cir)
//		idx, err = r.svc.createIndex(c, createIndexRequest2Proto(cir), nil)
//		if r.errorRespnse(c, err, "") {
//			return
//		}
//	} else {
//		form, err := c.MultipartForm()
//		if err != nil {
//			r.errorRespnse(c, errors.ErrUnimplemented, "expecting either application/json or multipart/form-data with valid request")
//			return
//		}
//
//		metas := form.Value["meta"]
//		if len(metas) > 0 {
//			meta := metas[0]
//			err = json.Unmarshal(cast.StringToByteArray(meta), &cir)
//			if err != nil {
//				r.errorRespnse(c, errors.ErrInvalid, fmt.Sprintf("could not parse meta value: %s", err.Error()))
//				return
//			}
//		}
//		fh, err := c.FormFile("file")
//		if err != nil {
//			r.errorRespnse(c, errors.ErrInvalid, fmt.Sprintf("for multipar/form-data file field must be provided: %s", err.Error()))
//			return
//		}
//
//		file, err := fh.Open()
//		if r.errorRespnse(c, err, "") {
//			return
//		}
//		defer file.Close()
//
//		idx, err = r.svc.createIndex(c, createIndexRequest2Proto(cir), file)
//		if r.errorRespnse(c, err, "") {
//			return
//		}
//	}
//	c.Header("Location", ComposeURI(c.Request, idx.Id))
//	c.JSON(http.StatusCreated, index2Rest(idx))
//
//}

//func (r *Rest) DeleteIndex(c *gin.Context, indexId similapi.IndexId) {
//	_, err := r.svc.idxService.Delete(c, &index.Id{Id: indexId})
//	if r.errorRespnse(c, err, "") {
//		return
//	}
//	c.Status(http.StatusNoContent)
//}
//
//func (r *Rest) GetIndex(c *gin.Context, indexId similapi.IndexId) {
//	idx, err := r.svc.IndexServiceServer().Get(c, &index.Id{Id: indexId})
//	if r.errorRespnse(c, err, "") {
//		return
//	}
//	c.JSON(http.StatusOK, index2Rest(idx))
//}
//
//func (r *Rest) PutIndex(c *gin.Context, indexId similapi.IndexId) {
//	var idx similapi.Index
//	if r.errorRespnse(c, BindAppJson(c, &idx), "") {
//		return
//	}
//	idx.Id = indexId
//	updated, err := r.svc.IndexServiceServer().Put(c, index2Proto(idx))
//	if r.errorRespnse(c, err, "") {
//		return
//	}
//	c.JSON(http.StatusOK, index2Rest(updated))
//}

//func (r *Rest) GetIndexRecords(c *gin.Context, indexId similapi.IndexId, params similapi.GetIndexRecordsParams) {
//	var req index.ListRecordsRequest
//	req.Id = indexId
//	if params.PageId != nil {
//		req.StartRecordId = params.PageId
//	}
//	if params.Limit != nil {
//		req.Limit = cast.Ptr(int64(*params.Limit))
//	}
//	res, err := r.svc.IndexServiceServer().ListRecords(c, &req)
//	if r.errorRespnse(c, err, "") {
//		return
//	}
//	c.JSON(http.StatusOK, listRecordsResult2Rest(res))
//}

//func (r *Rest) PatchIndexRecords(c *gin.Context, indexId similapi.IndexId) {
//	var req similapi.PatchRecordsRequest
//	if r.errorRespnse(c, BindAppJson(c, &req), "") {
//		return
//	}
//	req.Id = indexId
//	updated, err := r.svc.IndexServiceServer().PatchRecords(c, patchIndexRecordsRequest2Proto(req))
//	if r.errorRespnse(c, err, "") {
//		return
//	}
//	c.JSON(http.StatusOK, patchIndexRecordsResult2Rest(updated))
//}

func (r *Rest) Ping(c *gin.Context) {
	r.logger.Debugf("ping")
	c.String(http.StatusOK, "pong")
}

func (r *Rest) errorRespnse(c *gin.Context, err error, msg string) bool {
	if err == nil {
		return false
	}
	if msg == "" {
		msg = err.Error()
	}
	status := http.StatusInternalServerError
	defer func() {
		c.JSON(status, ErrorMsg{Error: msg})
		r.logger.Warnf("%s %s -> %d %s", c.Request.Method, c.Request.URL, status, msg)
		r.logger.Debugf("original error: %s", err)
	}()

	if errors.Is(err, errors.ErrNotExist) {
		status = http.StatusNotFound
	} else if errors.Is(err, errors.ErrInvalid) {
		status = http.StatusBadRequest
	} else if errors.Is(err, errors.ErrExist) {
		status = http.StatusConflict
	} else if errors.Is(err, errors.ErrUnimplemented) {
		status = http.StatusUnsupportedMediaType
	}
	return true
}

// BindAppJson turns the request body to inf, but for "application/json" contents only
func BindAppJson(c *gin.Context, inf interface{}) error {
	ct := c.ContentType()
	if ct != "application/json" {
		return fmt.Errorf("invalid content type %s, \"application/json\" expected. err=%w", ct, errors.ErrInternal)
	}
	err := c.Bind(inf)
	if err != nil {
		err = fmt.Errorf("%s: %w", err.Error(), errors.ErrInternal)
	}
	return err
}

// ComposeURI helper function which composes URI, adding ID to the request path
func ComposeURI(r *http.Request, id string) string {
	var resURL string
	if r.URL.IsAbs() {
		resURL = path.Join(r.URL.String(), id)
	} else {
		resURL = ResolveScheme(r) + "://" + path.Join(ResolveHost(r), r.URL.String(), id)
	}
	return resURL
}

// ResolveScheme resolves initial request type by r
func ResolveScheme(r *http.Request) string {
	switch {
	case r.Header.Get("X-Forwarded-Proto") == "https":
		return "https"
	case r.URL.Scheme == "https":
		return "https"
	case r.TLS != nil:
		return "https"
	case strings.HasPrefix(r.Proto, "HTTPS"):
		return "https"
	default:
		return "http"
	}
}

// ResolveHost returns host part of r
func ResolveHost(r *http.Request) (host string) {
	switch {
	case r.Header.Get("X-Forwarded-For") != "":
		return r.Header.Get("X-Forwarded-For")
	case r.Header.Get("X-Host") != "":
		return r.Header.Get("X-Host")
	case r.Host != "":
		return r.Host
	case r.URL.Host != "":
		return r.URL.Host
	default:
		return ""
	}
}
