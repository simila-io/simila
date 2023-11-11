// Package similapi provides primitives to interact with the openapi HTTP API.
//
// Code generated by github.com/deepmap/oapi-codegen version v1.16.2 DO NOT EDIT.
package similapi

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/gin-gonic/gin"
	"github.com/oapi-codegen/runtime"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

// CreateIndexRequest The object describes a new index request.
type CreateIndexRequest struct {
	// Document contains the binary data for the document of the specified format
	Document []byte `json:"document"`

	// Format the index format name
	Format string `json:"format"`

	// Id the new index identifier. It must not be more than 64 bytes long
	Id string `json:"id"`

	// Records the list of records that must be added to the new index.
	Records []Record `json:"records"`
	Tags    Tags     `json:"tags"`
}

// Format The object describes a data format.
type Format struct {
	// Name The format name. It is used as the format identifier
	Name string `json:"name"`
}

// Formats The object is used as response of contact objects query request.
type Formats struct {
	// Formats The list of all known formats
	Formats []Format `json:"formats"`
}

// Index An index description
type Index struct {
	CreatedAt time.Time `json:"createdAt"`
	Format    string    `json:"format"`
	Id        string    `json:"id"`
	Tags      Tags      `json:"tags"`
}

// Indexes The object contains information about an index record.
type Indexes struct {
	// Indexes The list of indexes
	Indexes []Index `json:"indexes"`

	// NextPageId the index Id for the next page, if presents
	NextPageId *string `json:"nextPageId,omitempty"`

	// Total total number of indexes that match the initial criteria
	Total int `json:"total"`
}

// PatchRecordsRequest defines model for PatchRecordsRequest.
type PatchRecordsRequest struct {
	DeleteRecords []Record `json:"deleteRecords"`
	Id            string   `json:"id"`
	UpsertRecords []Record `json:"upsertRecords"`
}

// PatchRecordsResult defines model for PatchRecordsResult.
type PatchRecordsResult struct {
	Deleted  int `json:"deleted"`
	Upserted int `json:"upserted"`
}

// Record The object contains information about an index record.
type Record struct {
	// Id the record identifier within the index. The value must be unique for the index and it is defined by the format parser.
	Id string `json:"id"`

	// Segment contains the searchable text for the record.
	Segment string `json:"segment"`

	// Vector contains the vector data for the record in the format basis. The format parser defines the basis and the field structure.
	Vector []byte `json:"vector"`
}

// RecordsResult defines model for RecordsResult.
type RecordsResult struct {
	NextPageId *string  `json:"nextPageId,omitempty"`
	Records    []Record `json:"records"`
	Total      int      `json:"total"`
}

// SearchRecord defines model for SearchRecord.
type SearchRecord struct {
	IndexId string `json:"indexId"`

	// IndexRecord The object contains information about an index record.
	IndexRecord     Record   `json:"indexRecord"`
	MatchedKeywords []string `json:"matchedKeywords"`
	Score           int      `json:"score"`
}

// SearchRequest defines model for SearchRequest.
type SearchRequest struct {
	Distinct     bool     `json:"distinct"`
	IndexIDs     []string `json:"indexIDs"`
	Limit        int      `json:"limit"`
	Offset       int      `json:"offset"`
	OrderByScore bool     `json:"orderByScore"`
	PageId       string   `json:"pageId"`
	Tags         Tags     `json:"tags"`
	Text         string   `json:"text"`
}

// SearchResult defines model for SearchResult.
type SearchResult struct {
	NextPageId *string        `json:"nextPageId,omitempty"`
	Records    []SearchRecord `json:"records"`
	Total      int            `json:"total"`
}

// Tags defines model for Tags.
type Tags map[string]string

// CreatedAfter defines model for CreatedAfter.
type CreatedAfter = time.Time

// CreatedBefore defines model for CreatedBefore.
type CreatedBefore = time.Time

// FormatId defines model for FormatId.
type FormatId = string

// FormatParam defines model for FormatParam.
type FormatParam = string

// IndexId defines model for IndexId.
type IndexId = string

// Limit defines model for Limit.
type Limit = int

// PageId defines model for PageId.
type PageId = string

// StartIndexId defines model for StartIndexId.
type StartIndexId = string

// TagsParam defines model for TagsParam.
type TagsParam = Tags

// GetIndexesParams defines parameters for GetIndexes.
type GetIndexesParams struct {
	// CreatedAfter start of time interval in which items are queried
	CreatedAfter *CreatedAfter `form:"createdAfter,omitempty" json:"createdAfter,omitempty"`

	// CreatedBefore end of time interval in which items are queried
	CreatedBefore *CreatedBefore `form:"createdBefore,omitempty" json:"createdBefore,omitempty"`

	// StartIndexId The indexId for the first record
	StartIndexId *StartIndexId `form:"startIndexId,omitempty" json:"startIndexId,omitempty"`
	Format       *FormatParam  `form:"format,omitempty" json:"format,omitempty"`
	Tags         *TagsParam    `form:"tags,omitempty" json:"tags,omitempty"`

	// Limit The limit defines the max number of objects returned per page.
	Limit *Limit `form:"limit,omitempty" json:"limit,omitempty"`
}

// CreateIndexMultipartBody defines parameters for CreateIndex.
type CreateIndexMultipartBody struct {
	// File contains the binary data for the document of the specified format
	File *openapi_types.File `json:"file,omitempty"`

	// Meta The object describes a new index request.
	Meta *CreateIndexRequest `json:"meta,omitempty"`
}

// GetIndexRecordsParams defines parameters for GetIndexRecords.
type GetIndexRecordsParams struct {
	// PageId The pageId for the paging request
	PageId *PageId `form:"pageId,omitempty" json:"pageId,omitempty"`

	// Limit The limit defines the max number of objects returned per page.
	Limit *Limit `form:"limit,omitempty" json:"limit,omitempty"`
}

// CreateFormatJSONRequestBody defines body for CreateFormat for application/json ContentType.
type CreateFormatJSONRequestBody = Format

// CreateIndexJSONRequestBody defines body for CreateIndex for application/json ContentType.
type CreateIndexJSONRequestBody = CreateIndexRequest

// CreateIndexMultipartRequestBody defines body for CreateIndex for multipart/form-data ContentType.
type CreateIndexMultipartRequestBody CreateIndexMultipartBody

// PutIndexJSONRequestBody defines body for PutIndex for application/json ContentType.
type PutIndexJSONRequestBody = Index

// PatchIndexRecordsJSONRequestBody defines body for PatchIndexRecords for application/json ContentType.
type PatchIndexRecordsJSONRequestBody = PatchRecordsRequest

// SearchJSONRequestBody defines body for Search for application/json ContentType.
type SearchJSONRequestBody = SearchRequest

// ServerInterface represents all server handlers.
type ServerInterface interface {
	// Retreive all known formats
	// (GET /formats)
	GetFormats(c *gin.Context)
	// Create new format
	// (POST /formats)
	CreateFormat(c *gin.Context)

	// (DELETE /formats/{formatId})
	DeleteFormat(c *gin.Context, formatId FormatId)

	// (GET /formats/{formatId})
	GetFormat(c *gin.Context, formatId FormatId)
	// Retreive indexes
	// (GET /indexes)
	GetIndexes(c *gin.Context, params GetIndexesParams)
	// Create new index
	// (POST /indexes)
	CreateIndex(c *gin.Context)

	// (DELETE /indexes/{indexId})
	DeleteIndex(c *gin.Context, indexId IndexId)

	// (GET /indexes/{indexId})
	GetIndex(c *gin.Context, indexId IndexId)

	// (PUT /indexes/{indexId})
	PutIndex(c *gin.Context, indexId IndexId)

	// (GET /indexes/{indexId}/records)
	GetIndexRecords(c *gin.Context, indexId IndexId, params GetIndexRecordsParams)

	// (PATCH /indexes/{indexId}/records)
	PatchIndexRecords(c *gin.Context, indexId IndexId)

	// (GET /ping)
	Ping(c *gin.Context)

	// (POST /search)
	Search(c *gin.Context)
}

// ServerInterfaceWrapper converts contexts to parameters.
type ServerInterfaceWrapper struct {
	Handler            ServerInterface
	HandlerMiddlewares []MiddlewareFunc
	ErrorHandler       func(*gin.Context, error, int)
}

type MiddlewareFunc func(c *gin.Context)

// GetFormats operation middleware
func (siw *ServerInterfaceWrapper) GetFormats(c *gin.Context) {

	for _, middleware := range siw.HandlerMiddlewares {
		middleware(c)
		if c.IsAborted() {
			return
		}
	}

	siw.Handler.GetFormats(c)
}

// CreateFormat operation middleware
func (siw *ServerInterfaceWrapper) CreateFormat(c *gin.Context) {

	for _, middleware := range siw.HandlerMiddlewares {
		middleware(c)
		if c.IsAborted() {
			return
		}
	}

	siw.Handler.CreateFormat(c)
}

// DeleteFormat operation middleware
func (siw *ServerInterfaceWrapper) DeleteFormat(c *gin.Context) {

	var err error

	// ------------- Path parameter "formatId" -------------
	var formatId FormatId

	err = runtime.BindStyledParameter("simple", false, "formatId", c.Param("formatId"), &formatId)
	if err != nil {
		siw.ErrorHandler(c, fmt.Errorf("Invalid format for parameter formatId: %w", err), http.StatusBadRequest)
		return
	}

	for _, middleware := range siw.HandlerMiddlewares {
		middleware(c)
		if c.IsAborted() {
			return
		}
	}

	siw.Handler.DeleteFormat(c, formatId)
}

// GetFormat operation middleware
func (siw *ServerInterfaceWrapper) GetFormat(c *gin.Context) {

	var err error

	// ------------- Path parameter "formatId" -------------
	var formatId FormatId

	err = runtime.BindStyledParameter("simple", false, "formatId", c.Param("formatId"), &formatId)
	if err != nil {
		siw.ErrorHandler(c, fmt.Errorf("Invalid format for parameter formatId: %w", err), http.StatusBadRequest)
		return
	}

	for _, middleware := range siw.HandlerMiddlewares {
		middleware(c)
		if c.IsAborted() {
			return
		}
	}

	siw.Handler.GetFormat(c, formatId)
}

// GetIndexes operation middleware
func (siw *ServerInterfaceWrapper) GetIndexes(c *gin.Context) {

	var err error

	// Parameter object where we will unmarshal all parameters from the context
	var params GetIndexesParams

	// ------------- Optional query parameter "createdAfter" -------------

	err = runtime.BindQueryParameter("form", true, false, "createdAfter", c.Request.URL.Query(), &params.CreatedAfter)
	if err != nil {
		siw.ErrorHandler(c, fmt.Errorf("Invalid format for parameter createdAfter: %w", err), http.StatusBadRequest)
		return
	}

	// ------------- Optional query parameter "createdBefore" -------------

	err = runtime.BindQueryParameter("form", true, false, "createdBefore", c.Request.URL.Query(), &params.CreatedBefore)
	if err != nil {
		siw.ErrorHandler(c, fmt.Errorf("Invalid format for parameter createdBefore: %w", err), http.StatusBadRequest)
		return
	}

	// ------------- Optional query parameter "startIndexId" -------------

	err = runtime.BindQueryParameter("form", true, false, "startIndexId", c.Request.URL.Query(), &params.StartIndexId)
	if err != nil {
		siw.ErrorHandler(c, fmt.Errorf("Invalid format for parameter startIndexId: %w", err), http.StatusBadRequest)
		return
	}

	// ------------- Optional query parameter "format" -------------

	err = runtime.BindQueryParameter("form", true, false, "format", c.Request.URL.Query(), &params.Format)
	if err != nil {
		siw.ErrorHandler(c, fmt.Errorf("Invalid format for parameter format: %w", err), http.StatusBadRequest)
		return
	}

	// ------------- Optional query parameter "tags" -------------

	err = runtime.BindQueryParameter("form", true, false, "tags", c.Request.URL.Query(), &params.Tags)
	if err != nil {
		siw.ErrorHandler(c, fmt.Errorf("Invalid format for parameter tags: %w", err), http.StatusBadRequest)
		return
	}

	// ------------- Optional query parameter "limit" -------------

	err = runtime.BindQueryParameter("form", true, false, "limit", c.Request.URL.Query(), &params.Limit)
	if err != nil {
		siw.ErrorHandler(c, fmt.Errorf("Invalid format for parameter limit: %w", err), http.StatusBadRequest)
		return
	}

	for _, middleware := range siw.HandlerMiddlewares {
		middleware(c)
		if c.IsAborted() {
			return
		}
	}

	siw.Handler.GetIndexes(c, params)
}

// CreateIndex operation middleware
func (siw *ServerInterfaceWrapper) CreateIndex(c *gin.Context) {

	for _, middleware := range siw.HandlerMiddlewares {
		middleware(c)
		if c.IsAborted() {
			return
		}
	}

	siw.Handler.CreateIndex(c)
}

// DeleteIndex operation middleware
func (siw *ServerInterfaceWrapper) DeleteIndex(c *gin.Context) {

	var err error

	// ------------- Path parameter "indexId" -------------
	var indexId IndexId

	err = runtime.BindStyledParameter("simple", false, "indexId", c.Param("indexId"), &indexId)
	if err != nil {
		siw.ErrorHandler(c, fmt.Errorf("Invalid format for parameter indexId: %w", err), http.StatusBadRequest)
		return
	}

	for _, middleware := range siw.HandlerMiddlewares {
		middleware(c)
		if c.IsAborted() {
			return
		}
	}

	siw.Handler.DeleteIndex(c, indexId)
}

// GetIndex operation middleware
func (siw *ServerInterfaceWrapper) GetIndex(c *gin.Context) {

	var err error

	// ------------- Path parameter "indexId" -------------
	var indexId IndexId

	err = runtime.BindStyledParameter("simple", false, "indexId", c.Param("indexId"), &indexId)
	if err != nil {
		siw.ErrorHandler(c, fmt.Errorf("Invalid format for parameter indexId: %w", err), http.StatusBadRequest)
		return
	}

	for _, middleware := range siw.HandlerMiddlewares {
		middleware(c)
		if c.IsAborted() {
			return
		}
	}

	siw.Handler.GetIndex(c, indexId)
}

// PutIndex operation middleware
func (siw *ServerInterfaceWrapper) PutIndex(c *gin.Context) {

	var err error

	// ------------- Path parameter "indexId" -------------
	var indexId IndexId

	err = runtime.BindStyledParameter("simple", false, "indexId", c.Param("indexId"), &indexId)
	if err != nil {
		siw.ErrorHandler(c, fmt.Errorf("Invalid format for parameter indexId: %w", err), http.StatusBadRequest)
		return
	}

	for _, middleware := range siw.HandlerMiddlewares {
		middleware(c)
		if c.IsAborted() {
			return
		}
	}

	siw.Handler.PutIndex(c, indexId)
}

// GetIndexRecords operation middleware
func (siw *ServerInterfaceWrapper) GetIndexRecords(c *gin.Context) {

	var err error

	// ------------- Path parameter "indexId" -------------
	var indexId IndexId

	err = runtime.BindStyledParameter("simple", false, "indexId", c.Param("indexId"), &indexId)
	if err != nil {
		siw.ErrorHandler(c, fmt.Errorf("Invalid format for parameter indexId: %w", err), http.StatusBadRequest)
		return
	}

	// Parameter object where we will unmarshal all parameters from the context
	var params GetIndexRecordsParams

	// ------------- Optional query parameter "pageId" -------------

	err = runtime.BindQueryParameter("form", true, false, "pageId", c.Request.URL.Query(), &params.PageId)
	if err != nil {
		siw.ErrorHandler(c, fmt.Errorf("Invalid format for parameter pageId: %w", err), http.StatusBadRequest)
		return
	}

	// ------------- Optional query parameter "limit" -------------

	err = runtime.BindQueryParameter("form", true, false, "limit", c.Request.URL.Query(), &params.Limit)
	if err != nil {
		siw.ErrorHandler(c, fmt.Errorf("Invalid format for parameter limit: %w", err), http.StatusBadRequest)
		return
	}

	for _, middleware := range siw.HandlerMiddlewares {
		middleware(c)
		if c.IsAborted() {
			return
		}
	}

	siw.Handler.GetIndexRecords(c, indexId, params)
}

// PatchIndexRecords operation middleware
func (siw *ServerInterfaceWrapper) PatchIndexRecords(c *gin.Context) {

	var err error

	// ------------- Path parameter "indexId" -------------
	var indexId IndexId

	err = runtime.BindStyledParameter("simple", false, "indexId", c.Param("indexId"), &indexId)
	if err != nil {
		siw.ErrorHandler(c, fmt.Errorf("Invalid format for parameter indexId: %w", err), http.StatusBadRequest)
		return
	}

	for _, middleware := range siw.HandlerMiddlewares {
		middleware(c)
		if c.IsAborted() {
			return
		}
	}

	siw.Handler.PatchIndexRecords(c, indexId)
}

// Ping operation middleware
func (siw *ServerInterfaceWrapper) Ping(c *gin.Context) {

	for _, middleware := range siw.HandlerMiddlewares {
		middleware(c)
		if c.IsAborted() {
			return
		}
	}

	siw.Handler.Ping(c)
}

// Search operation middleware
func (siw *ServerInterfaceWrapper) Search(c *gin.Context) {

	for _, middleware := range siw.HandlerMiddlewares {
		middleware(c)
		if c.IsAborted() {
			return
		}
	}

	siw.Handler.Search(c)
}

// GinServerOptions provides options for the Gin server.
type GinServerOptions struct {
	BaseURL      string
	Middlewares  []MiddlewareFunc
	ErrorHandler func(*gin.Context, error, int)
}

// RegisterHandlers creates http.Handler with routing matching OpenAPI spec.
func RegisterHandlers(router gin.IRouter, si ServerInterface) {
	RegisterHandlersWithOptions(router, si, GinServerOptions{})
}

// RegisterHandlersWithOptions creates http.Handler with additional options
func RegisterHandlersWithOptions(router gin.IRouter, si ServerInterface, options GinServerOptions) {
	errorHandler := options.ErrorHandler
	if errorHandler == nil {
		errorHandler = func(c *gin.Context, err error, statusCode int) {
			c.JSON(statusCode, gin.H{"msg": err.Error()})
		}
	}

	wrapper := ServerInterfaceWrapper{
		Handler:            si,
		HandlerMiddlewares: options.Middlewares,
		ErrorHandler:       errorHandler,
	}

	router.GET(options.BaseURL+"/formats", wrapper.GetFormats)
	router.POST(options.BaseURL+"/formats", wrapper.CreateFormat)
	router.DELETE(options.BaseURL+"/formats/:formatId", wrapper.DeleteFormat)
	router.GET(options.BaseURL+"/formats/:formatId", wrapper.GetFormat)
	router.GET(options.BaseURL+"/indexes", wrapper.GetIndexes)
	router.POST(options.BaseURL+"/indexes", wrapper.CreateIndex)
	router.DELETE(options.BaseURL+"/indexes/:indexId", wrapper.DeleteIndex)
	router.GET(options.BaseURL+"/indexes/:indexId", wrapper.GetIndex)
	router.PUT(options.BaseURL+"/indexes/:indexId", wrapper.PutIndex)
	router.GET(options.BaseURL+"/indexes/:indexId/records", wrapper.GetIndexRecords)
	router.PATCH(options.BaseURL+"/indexes/:indexId/records", wrapper.PatchIndexRecords)
	router.GET(options.BaseURL+"/ping", wrapper.Ping)
	router.POST(options.BaseURL+"/search", wrapper.Search)
}

// Base64 encoded, gzipped, json marshaled Swagger object
var swaggerSpec = []string{

	"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",
	"1liDKQAA",
}

// GetSwagger returns the content of the embedded swagger specification file
// or error if failed to decode
func decodeSpec() ([]byte, error) {
	zipped, err := base64.StdEncoding.DecodeString(strings.Join(swaggerSpec, ""))
	if err != nil {
		return nil, fmt.Errorf("error base64 decoding spec: %w", err)
	}
	zr, err := gzip.NewReader(bytes.NewReader(zipped))
	if err != nil {
		return nil, fmt.Errorf("error decompressing spec: %w", err)
	}
	var buf bytes.Buffer
	_, err = buf.ReadFrom(zr)
	if err != nil {
		return nil, fmt.Errorf("error decompressing spec: %w", err)
	}

	return buf.Bytes(), nil
}

var rawSpec = decodeSpecCached()

// a naive cached of a decoded swagger spec
func decodeSpecCached() func() ([]byte, error) {
	data, err := decodeSpec()
	return func() ([]byte, error) {
		return data, err
	}
}

// Constructs a synthetic filesystem for resolving external references when loading openapi specifications.
func PathToRawSpec(pathToFile string) map[string]func() ([]byte, error) {
	res := make(map[string]func() ([]byte, error))
	if len(pathToFile) > 0 {
		res[pathToFile] = rawSpec
	}

	return res
}

// GetSwagger returns the Swagger specification corresponding to the generated code
// in this file. The external references of Swagger specification are resolved.
// The logic of resolving external references is tightly connected to "import-mapping" feature.
// Externally referenced files must be embedded in the corresponding golang packages.
// Urls can be supported but this task was out of the scope.
func GetSwagger() (swagger *openapi3.T, err error) {
	resolvePath := PathToRawSpec("")

	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true
	loader.ReadFromURIFunc = func(loader *openapi3.Loader, url *url.URL) ([]byte, error) {
		pathToFile := url.String()
		pathToFile = path.Clean(pathToFile)
		getSpec, ok := resolvePath[pathToFile]
		if !ok {
			err1 := fmt.Errorf("path not found: %s", pathToFile)
			return nil, err1
		}
		return getSpec()
	}
	var specData []byte
	specData, err = rawSpec()
	if err != nil {
		return
	}
	swagger, err = loader.LoadFromData(specData)
	if err != nil {
		return
	}
	return
}
