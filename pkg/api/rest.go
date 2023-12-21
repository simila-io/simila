package api

import (
	"encoding/json"
	"fmt"
	"github.com/acquirecloud/golibs/cast"
	"github.com/acquirecloud/golibs/errors"
	"github.com/acquirecloud/golibs/logging"
	"github.com/gin-gonic/gin"
	"github.com/simila-io/simila/api/gen/format/v1"
	"github.com/simila-io/simila/api/gen/index/v1"
	similapi "github.com/simila-io/simila/api/genpublic/v1"
	"github.com/simila-io/simila/pkg/indexer/persistence"
	"google.golang.org/protobuf/types/known/timestamppb"
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
	nodes, err := r.svc.IndexServiceServer().ListNodes(c, &index.ListNodesRequest{FilterConditions: cast.Value(params.Condition, ""),
		Offset: int64(cast.Value(params.Offset, 0)), Limit: int64(cast.Value(params.Limit, 100))})
	if r.errorRespnse(c, err, "") {
		return
	}
	c.JSON(http.StatusOK, similapi.ListNodesResult{Items: nodes2Rest(nodes.Nodes)})
}

func (r *Rest) UpdateNode(c *gin.Context, path similapi.Path) {
	var n similapi.Node
	if r.errorRespnse(c, BindAppJson(c, &n), "") {
		return
	}
	_, err := r.svc.IndexServiceServer().UpdateNode(c,
		&index.UpdateNodeRequest{Path: persistence.ConcatPath(path, ""), Node: &index.Node{Tags: n.Tags}})
	if r.errorRespnse(c, err, "") {
		return
	}
	c.Status(http.StatusOK)
}

func (r *Rest) DeleteNodes(c *gin.Context) {
	var dnr similapi.DeleteNodesRequest
	if r.errorRespnse(c, BindAppJson(c, &dnr), "") {
		return
	}
	_, err := r.svc.IndexServiceServer().DeleteNodes(c, &index.DeleteNodesRequest{FilterConditions: dnr.FilterConditions, Force: cast.Ptr(dnr.Force)})
	if r.errorRespnse(c, err, "") {
		return
	}
	c.Status(http.StatusNoContent)
}

func (r *Rest) DeleteNode(c *gin.Context, path similapi.Path, params similapi.DeleteNodeParams) {
	_, err := r.svc.IndexServiceServer().DeleteNodes(c, &index.DeleteNodesRequest{FilterConditions: fmt.Sprintf("path like '%s%%'", path), Force: params.Force})
	if r.errorRespnse(c, err, "") {
		return
	}
	c.Status(http.StatusNoContent)
}

func (r *Rest) ListNodeRecords(c *gin.Context, path similapi.Path, params similapi.ListNodeRecordsParams) {
	lrr := &index.ListRequest{}
	lrr.Path = persistence.ConcatPath(path, "")
	if params.Limit != nil {
		lrr.Limit = cast.Ptr(int64(*params.Limit))
	}
	lrr.Format = params.Format
	if params.CreatedAfter != nil {
		lrr.CreatedAfter = timestamppb.New(*params.CreatedAfter)
	}
	if params.CreatedBefore != nil {
		lrr.CreatedBefore = timestamppb.New(*params.CreatedBefore)
	}
	lrr.PageId = params.PageId
	lr, err := r.svc.IndexServiceServer().ListRecords(c, lrr)
	if r.errorRespnse(c, err, "") {
		return
	}
	c.JSON(http.StatusOK, similapi.ListRecordsResult{Records: cast.Ptr(records2Rest(lr.Records)), Total: int(lr.Total), NextPageId: lr.NextPageId})
}

func (r *Rest) PatchNodeRecords(c *gin.Context, path similapi.Path) {
	var pr similapi.PatchRecordsRequest
	if r.errorRespnse(c, BindAppJson(c, &pr), "") {
		return
	}
	prr, err := r.svc.IndexServiceServer().PatchRecords(c, &index.PatchRecordsRequest{
		Path:          path,
		DeleteRecords: rest2Records(pr.DeleteRecords),
		UpsertRecords: rest2Records(pr.UpsertRecords),
	})
	if r.errorRespnse(c, err, "") {
		return
	}
	c.JSON(http.StatusOK, similapi.PatchRecordsResult{Deleted: int(prr.Deleted), Upserted: int(prr.Upserted)})
}

func (r *Rest) CreateNodeRecords(c *gin.Context, path similapi.Path) {
	var crr similapi.CreateRecordsRequest
	var res *index.CreateRecordsResult
	if err := BindAppJson(c, &crr); err == nil {
		r.logger.Infof("creating new node records %v", crr)
		res, err = r.svc.createRecords(c, rest2CreateRecordsRequest(path, crr), nil)
		if r.errorRespnse(c, err, "") {
			return
		}
	} else {
		form, err := c.MultipartForm()
		if err != nil {
			r.errorRespnse(c, errors.ErrUnimplemented, "expecting either application/json or multipart/form-data with valid request")
			return
		}

		metas := form.Value["meta"]
		if len(metas) > 0 {
			meta := metas[0]
			err = json.Unmarshal(cast.StringToByteArray(meta), &crr)
			if err != nil {
				r.errorRespnse(c, errors.ErrInvalid, fmt.Sprintf("could not parse meta value: %s", err.Error()))
				return
			}
		}
		fh, err := c.FormFile("file")
		if err != nil {
			r.errorRespnse(c, errors.ErrInvalid, fmt.Sprintf("for multipar/form-data file field must be provided: %s", err.Error()))
			return
		}

		file, err := fh.Open()
		if r.errorRespnse(c, err, "") {
			return
		}
		defer file.Close()

		res, err = r.svc.createRecords(c, rest2CreateRecordsRequest(path, crr), file)
		if r.errorRespnse(c, err, "") {
			return
		}
	}
	var nc []similapi.Node
	if res.NodesCreated != nil {
		nc = nodes2Rest(res.NodesCreated.Nodes)
	}
	c.JSON(http.StatusCreated, similapi.CreateRecordsResult{RecordsCreated: int(res.RecordsCreated), NodesCreated: nc})
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
	f1, err := r.svc.FormatServiceServer().Create(c, &format.Format{Name: f.Name, Basis: f.Basis})
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
