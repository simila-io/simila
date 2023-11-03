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
	"encoding/json"
	"fmt"
	"github.com/acquirecloud/golibs/cast"
	"github.com/acquirecloud/golibs/errors"
	"github.com/acquirecloud/golibs/logging"
	"github.com/gin-gonic/gin"
	"github.com/simila-io/simila/api/gen/format/v1"
	"github.com/simila-io/simila/api/gen/index/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
	"net/http"
	"path"
	"strconv"
	"strings"
	"time"
)

// HttpEP provides the api endpoints for the HTTP interface
type HttpEP struct {
	svc    *Service
	logger logging.Logger
}

type ErrorMsg struct {
	Error string `json:"error"`
}

func NewHttpEP(svc *Service) *HttpEP {
	return &HttpEP{svc: svc, logger: logging.NewLogger("api.rest")}
}

func (hep *HttpEP) RegisterEPs(g *gin.Engine) error {
	g.GET("/v1/ping", hep.hGetPing)
	g.POST("/v1/indexes", hep.hPostIndexes)
	g.GET("/v1/indexes", hep.hGetIndexes)
	g.DELETE("/v1/indexes/:id", hep.hDeleteIndexesId)
	g.GET("/v1/indexes/:id", hep.hGetIndexesId)
	g.PUT("/v1/indexes/:id", hep.hPutIndexesId)
	g.PATCH("/v1/indexes/:id/records", hep.hPatchIndexesIdRecords)
	g.GET("/v1/indexes/:id/records", hep.hGetIndexesIdRecords)
	g.POST("/v1/search", hep.hPostSearch)
	g.POST("/v1/formats", hep.hPostFormats)
	g.GET("/v1/formats", hep.hGetFormats)
	g.DELETE("/v1/formats/:id", hep.hDeleteFormatsId)
	g.GET("/v1/formats/:id", hep.hGetFormatsId)
	return nil
}

func (hep *HttpEP) hGetPing(c *gin.Context) {
	hep.logger.Debugf("ping")
	c.String(http.StatusOK, "pong")
}

// hPostIndexes accepts a multipart form for a file data together with a json in meta field
// example:
// curl -i -X POST -H "Content-Type: multipart/form-data" -F"file=@/Users/user/Downloads/fr_9782_size1024.jpg" -F "meta={\"aa\": \"bbb\"};type=application/json" http://localhost:8080/v1/indexes
func (hep *HttpEP) hPostIndexes(c *gin.Context) {
	var cir CreateIndexRequest
	var idx *index.Index
	if err := BindAppJson(c, &cir); err == nil {
		hep.logger.Infof("creating new index from the json object %s", cir)
		idx, err = hep.svc.createIndex(c, CreateIndexRequest2Proto(cir), nil)
		if hep.errorRespnse(c, err, "") {
			return
		}
	} else {
		form, err := c.MultipartForm()
		if err != nil {
			hep.errorRespnse(c, errors.ErrUnimplemented, "expecting either application/json or multipart/form-data with valid request")
			return
		}

		metas := form.Value["meta"]
		if len(metas) > 0 {
			meta := metas[0]
			err = json.Unmarshal(cast.StringToByteArray(meta), &cir)
			if err != nil {
				hep.errorRespnse(c, errors.ErrInvalid, fmt.Sprintf("could not parse meta value: %s", err.Error()))
				return
			}
		}
		fh, err := c.FormFile("file")
		if err != nil {
			hep.errorRespnse(c, errors.ErrInvalid, fmt.Sprintf("for multipar/form-data file field must be provided: %s", err.Error()))
			return
		}

		file, err := fh.Open()
		if hep.errorRespnse(c, err, "") {
			return
		}
		defer file.Close()

		idx, err = hep.svc.createIndex(c, CreateIndexRequest2Proto(cir), file)
		if hep.errorRespnse(c, err, "") {
			return
		}
	}
	c.Header("Location", ComposeURI(c.Request, idx.Id))
	c.JSON(http.StatusCreated, Index2Rest(idx))
}

func (hep *HttpEP) hDeleteIndexesId(c *gin.Context) {
	iid := c.Param("id")
	_, err := hep.svc.IndexServiceServer().Delete(c, &index.Id{Id: iid})
	if hep.errorRespnse(c, err, "") {
		return
	}
	c.Status(http.StatusNoContent)
}

func (hep *HttpEP) hGetIndexesId(c *gin.Context) {
	iid := c.Param("id")
	idx, err := hep.svc.IndexServiceServer().Get(c, &index.Id{Id: iid})
	if hep.errorRespnse(c, err, "") {
		return
	}
	c.JSON(http.StatusOK, idx)
}

func (hep *HttpEP) hPutIndexesId(c *gin.Context) {
	iid := c.Param("id")
	var idx index.Index
	if hep.errorRespnse(c, BindAppJson(c, &idx), "") {
		return
	}
	idx.Id = iid
	updated, err := hep.svc.IndexServiceServer().Put(c, &idx)
	if hep.errorRespnse(c, err, "") {
		return
	}
	c.JSON(http.StatusOK, updated)
}

func (hep *HttpEP) hGetIndexes(c *gin.Context) {
	req := &index.ListRequest{}
	if sLimit := c.DefaultQuery("limit", ""); sLimit != "" {
		limit, err := strconv.ParseInt(sLimit, 10, 64)
		if hep.errorRespnse(c, err, "limit must be a number") {
			return
		}
		req.Limit = cast.Ptr(limit)
	}

	req.StartIndexId = c.DefaultQuery("start-index-id", "")
	if format := c.DefaultQuery("format", ""); format != "" {
		req.Format = cast.Ptr(format)
	}
	if tags := c.DefaultQuery("tags", ""); tags != "" {
		if hep.errorRespnse(c, json.Unmarshal(cast.StringToByteArray(tags), &req.Tags), fmt.Sprintf("the tags parameter should be a json map values like tags={\"key1\": \"val1\"..}")) {
			return
		}
	}
	if ca := c.DefaultQuery("created-after", ""); ca != "" {
		st, err := ParseTime(ca)
		if hep.errorRespnse(c, err, fmt.Sprintf("created-after must be formatted like %s", timeLayout)) {
			return
		}
		req.CreatedAfter = timestamppb.New(st)
	}
	if cb := c.DefaultQuery("created-before", ""); cb != "" {
		st, err := ParseTime(cb)
		if hep.errorRespnse(c, err, fmt.Sprintf("created-before must be formatted like %s", timeLayout)) {
			return
		}
		req.CreatedAfter = timestamppb.New(st)
	}
	idxs, err := hep.svc.IndexServiceServer().List(c, req)
	if hep.errorRespnse(c, err, "") {
		return
	}
	c.JSON(http.StatusOK, idxs)
}

func (hep *HttpEP) hPatchIndexesIdRecords(c *gin.Context) {
	iid := c.Param("id")
	var req PatchRecordsRequest
	if hep.errorRespnse(c, BindAppJson(c, &req), "") {
		return
	}
	req.Id = iid
	updated, err := hep.svc.IndexServiceServer().PatchRecords(c, PatchIndexRecordsRequest2Proto(req))
	if hep.errorRespnse(c, err, "") {
		return
	}
	c.JSON(http.StatusOK, updated)
}

func (hep *HttpEP) hPostSearch(c *gin.Context) {
	var sr index.SearchRecordsRequest
	if hep.errorRespnse(c, BindAppJson(c, &sr), "") {
		return
	}
	res, err := hep.svc.IndexServiceServer().SearchRecords(c, &sr)
	if hep.errorRespnse(c, err, "") {
		return
	}
	c.JSON(http.StatusOK, SearchRecordsResult2Rest(res))
}

func (hep *HttpEP) hGetIndexesIdRecords(c *gin.Context) {
	var req index.ListRecordsRequest
	req.Id = c.Param("id")
	if sLimit := c.DefaultQuery("limit", ""); sLimit != "" {
		limit, err := strconv.ParseInt(sLimit, 10, 64)
		if hep.errorRespnse(c, err, "limit must be a number") {
			return
		}
		req.Limit = cast.Ptr(limit)
	}
	if sri := c.DefaultQuery("start-record-id", ""); sri != "" {
		req.StartRecordId = cast.Ptr(sri)
	}
	res, err := hep.svc.IndexServiceServer().ListRecords(c, &req)
	if hep.errorRespnse(c, err, "") {
		return
	}
	c.JSON(http.StatusOK, ListRecordsResult2Proto(res))
}

func (hep *HttpEP) hPostFormats(c *gin.Context) {
	var f format.Format
	if hep.errorRespnse(c, BindAppJson(c, &f), "") {
		return
	}
	f1, err := hep.svc.FormatServiceServer().Create(c, &f)
	if hep.errorRespnse(c, err, "") {
		return
	}
	c.Header("Location", ComposeURI(c.Request, f1.Name))
	c.JSON(http.StatusCreated, f1)
}

func (hep *HttpEP) hGetFormats(c *gin.Context) {
	fmts, err := hep.svc.FormatServiceServer().List(c, nil)
	if hep.errorRespnse(c, err, "") {
		return
	}
	c.JSON(http.StatusOK, fmts)
}

func (hep *HttpEP) hDeleteFormatsId(c *gin.Context) {
	fid := c.Param("id")
	_, err := hep.svc.FormatServiceServer().Delete(c, &format.Id{Id: fid})
	if hep.errorRespnse(c, err, "") {
		return
	}
	c.Status(http.StatusNoContent)
}

func (hep *HttpEP) hGetFormatsId(c *gin.Context) {
	fid := c.Param("id")
	f, err := hep.svc.FormatServiceServer().Get(c, &format.Id{Id: fid})
	if hep.errorRespnse(c, err, "") {
		return
	}
	c.JSON(http.StatusOK, f)
}

func (hep *HttpEP) errorRespnse(c *gin.Context, err error, msg string) bool {
	if err == nil {
		return false
	}
	if msg == "" {
		msg = err.Error()
	}
	status := http.StatusInternalServerError
	defer func() {
		c.JSON(status, ErrorMsg{Error: msg})
		hep.logger.Warnf("%s %s -> %d %s", c.Request.Method, c.Request.URL, status, msg)
		hep.logger.Debugf("original error: %s", err)
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

var timeLayout = "2006-01-02T15:04:05-07:00"

func ParseTime(s string) (time.Time, error) {
	if s != "" {
		return time.Parse(timeLayout, s)
	}
	return time.Time{}, nil
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
