// Package similapi provides primitives to interact with the openapi HTTP API.
//
// Code generated by github.com/deepmap/oapi-codegen version v1.12.4 DO NOT EDIT.
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

	"github.com/deepmap/oapi-codegen/pkg/runtime"
	openapi_types "github.com/deepmap/oapi-codegen/pkg/types"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/gin-gonic/gin"
)

// Defines values for NodeType.
const (
	Document NodeType = "document"
	Folder   NodeType = "folder"
)

// CreateRecordsRequest The object is used for records creation.
type CreateRecordsRequest struct {
	// Document The binary data for the document of the specified format.
	Document []byte `json:"document"`

	// NodeType The object describes the index node type.
	NodeType NodeType `json:"nodeType"`

	// Parser The parser name (format name) to be used for the document body.
	Parser string `json:"parser"`

	// RankMultiplier The priority coefficient (must be >= 1.0) of the records within a search result set, the value is overridden by the rankMultiplier value specified for an individual record.
	RankMultiplier float32 `json:"rankMultiplier"`

	// Records The list of records that must be added to the node.
	Records []Record `json:"records"`

	// Tags The object describes the node tags.
	Tags Tags `json:"tags"`
}

// CreateRecordsResult The object is used as a response of the records creation request.
type CreateRecordsResult struct {
	// NodesCreated The list of nodes created.
	NodesCreated []Node `json:"nodesCreated"`

	// RecordsCreated The number of records created.
	RecordsCreated int `json:"recordsCreated"`
}

// Format The object describes a data format.
type Format struct {
	// Name The format name, it is used as the format identifier.
	Name string `json:"name"`
}

// Formats The object is used as a response of the formats list request.
type Formats struct {
	// Formats Contains a list of formats.
	Formats []Format `json:"formats"`
}

// ListNodesResult The object is used as a response of the nodes list request.
type ListNodesResult struct {
	// Items The list of nodes.
	Items []Node `json:"items"`
}

// ListRecordsResult The object is used a response to the list records request.
type ListRecordsResult struct {
	// NextPageId The id of the next page for getting the rest of the records.
	NextPageId *string `json:"nextPageId,omitempty"`

	// Records The list of found records.
	Records *[]Record `json:"records,omitempty"`

	// Total The total number of found records.
	Total int `json:"total"`
}

// Node The object describes the index node.
type Node struct {
	// Name The node name, must be unique among the siblings in the tree.
	Name string `json:"name"`

	// Path The node path (identifier) within the index.
	Path string `json:"path"`

	// Tags The object describes the node tags.
	Tags Tags `json:"tags"`

	// Type The object describes the index node type.
	Type NodeType `json:"type"`
}

// NodeType The object describes the index node type.
type NodeType string

// PatchRecordsRequest The object is used to upsert and delete the node records.
type PatchRecordsRequest struct {
	// DeleteRecords The records to be deleted for the node.
	DeleteRecords []Record `json:"deleteRecords"`

	// UpsertRecords The records to be upserted for the node.
	UpsertRecords []Record `json:"upsertRecords"`
}

// PatchRecordsResult The object is used as a response to the patch records request.
type PatchRecordsResult struct {
	// Deleted The number of deleted records.
	Deleted int `json:"deleted"`

	// Upserted The number of upserted records.
	Upserted int `json:"upserted"`
}

// Record The object contains information about the index record.
type Record struct {
	// Format The format of the record.
	Format string `json:"format"`

	// Id The record identifier within the node.
	Id string `json:"id"`

	// RankMultiplier The priority coefficient (must be >= 1.0) of the record within a search result set.
	RankMultiplier float32 `json:"rankMultiplier"`

	// Segment The searchable text for the record.
	Segment string `json:"segment"`

	// Vector The vector data for the segment.
	Vector []byte `json:"vector"`
}

// SearchRecordsRequest The object is used to search text across all the index records.
type SearchRecordsRequest struct {
	// Limit The maximum number of records per page.
	Limit int `json:"limit"`

	// Offset The number of records to skip before start returning results.
	Offset int `json:"offset"`

	// Path The node path where the text to be searched for, by default, this includes the node subtrees too.
	Path string `json:"path"`

	// Strict The strict flag limits the search scope strictly to the node located by the provided path, i.e. the node subtrees are not included.
	Strict bool `json:"strict"`

	// Tags The object describes the node tags.
	Tags Tags `json:"tags"`

	// Text The text to be found in the records.
	Text string `json:"text"`
}

// SearchRecordsResult The object is used as a response to the search records request.
type SearchRecordsResult struct {
	// Items The found index records.
	Items []SearchRecordsResultItem `json:"items"`

	// Total The total number of found records.
	Total int `json:"total"`
}

// SearchRecordsResultItem The object is used as an item in the search records response.
type SearchRecordsResultItem struct {
	// MatchedKeywords The matched keywords within the record.
	MatchedKeywords []string `json:"matchedKeywords"`

	// Path The path of the record.
	Path string `json:"path"`

	// Record The object contains information about the index record.
	Record Record `json:"record"`

	// Score The relevancy score of the record.
	Score float32 `json:"score"`
}

// Tags The object describes the node tags.
type Tags map[string]string

// CreatedAfterFilter defines model for CreatedAfterFilter.
type CreatedAfterFilter = time.Time

// CreatedBeforeFilter defines model for CreatedBeforeFilter.
type CreatedBeforeFilter = time.Time

// FormatFilter defines model for FormatFilter.
type FormatFilter = string

// FormatId defines model for FormatId.
type FormatId = string

// Limit defines model for Limit.
type Limit = int

// PageId defines model for PageId.
type PageId = string

// Path defines model for Path.
type Path = string

// PathFilter defines model for PathFilter.
type PathFilter = string

// TagsFilter The object describes the node tags.
type TagsFilter = Tags

// ListNodesParams defines parameters for ListNodes.
type ListNodesParams struct {
	// Path The path specifies the path to filter by.
	Path *PathFilter `form:"path,omitempty" json:"path,omitempty"`
}

// ListNodeRecordsParams defines parameters for ListNodeRecords.
type ListNodeRecordsParams struct {
	// Format The format specifies the format to filter the records by.
	Format *FormatFilter `form:"format,omitempty" json:"format,omitempty"`

	// CreatedAfter The createdAfter specifies the lowest creation time (exclusive) the resulting records can have.
	CreatedAfter *CreatedAfterFilter `form:"createdAfter,omitempty" json:"createdAfter,omitempty"`

	// CreatedBefore The createdBefore specifies the greatest creation time (exclusive) the resulting records can have.
	CreatedBefore *CreatedBeforeFilter `form:"createdBefore,omitempty" json:"createdBefore,omitempty"`

	// PageId The pageId specifies from which page to start return results.
	PageId *PageId `form:"pageId,omitempty" json:"pageId,omitempty"`

	// Limit The limit defines the max number of objects returned per page.
	Limit *Limit `form:"limit,omitempty" json:"limit,omitempty"`
}

// CreateNodeRecordsMultipartBody defines parameters for CreateNodeRecords.
type CreateNodeRecordsMultipartBody struct {
	// File The document binary data in the specified format.
	File *openapi_types.File `json:"file,omitempty"`

	// Meta The object is used for records creation.
	Meta *CreateRecordsRequest `json:"meta,omitempty"`
}

// CreateFormatJSONRequestBody defines body for CreateFormat for application/json ContentType.
type CreateFormatJSONRequestBody = Format

// UpdateNodeJSONRequestBody defines body for UpdateNode for application/json ContentType.
type UpdateNodeJSONRequestBody = Node

// PatchNodeRecordsJSONRequestBody defines body for PatchNodeRecords for application/json ContentType.
type PatchNodeRecordsJSONRequestBody = PatchRecordsRequest

// CreateNodeRecordsJSONRequestBody defines body for CreateNodeRecords for application/json ContentType.
type CreateNodeRecordsJSONRequestBody = CreateRecordsRequest

// CreateNodeRecordsMultipartRequestBody defines body for CreateNodeRecords for multipart/form-data ContentType.
type CreateNodeRecordsMultipartRequestBody CreateNodeRecordsMultipartBody

// SearchJSONRequestBody defines body for Search for application/json ContentType.
type SearchJSONRequestBody = SearchRecordsRequest

// ServerInterface represents all server handlers.
type ServerInterface interface {
	// List formats
	// (GET /formats)
	ListFormats(c *gin.Context)
	// Create new format
	// (POST /formats)
	CreateFormat(c *gin.Context)
	// Delete format
	// (DELETE /formats/{formatId})
	DeleteFormat(c *gin.Context, formatId FormatId)
	// Get format
	// (GET /formats/{formatId})
	GetFormat(c *gin.Context, formatId FormatId)
	// List nodes
	// (GET /nodes)
	ListNodes(c *gin.Context, params ListNodesParams)
	// Delete node
	// (DELETE /nodes/{path})
	DeleteNode(c *gin.Context, path Path)
	// Update node
	// (PUT /nodes/{path})
	UpdateNode(c *gin.Context, path Path)
	// List node records
	// (GET /nodes/{path}/records)
	ListNodeRecords(c *gin.Context, path Path, params ListNodeRecordsParams)
	// Patch node records
	// (PATCH /nodes/{path}/records)
	PatchNodeRecords(c *gin.Context, path Path)
	// Create node records
	// (POST /nodes/{path}/records)
	CreateNodeRecords(c *gin.Context, path Path)
	// Health check
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

// ListFormats operation middleware
func (siw *ServerInterfaceWrapper) ListFormats(c *gin.Context) {

	for _, middleware := range siw.HandlerMiddlewares {
		middleware(c)
	}

	siw.Handler.ListFormats(c)
}

// CreateFormat operation middleware
func (siw *ServerInterfaceWrapper) CreateFormat(c *gin.Context) {

	for _, middleware := range siw.HandlerMiddlewares {
		middleware(c)
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
		siw.ErrorHandler(c, fmt.Errorf("Invalid format for parameter formatId: %s", err), http.StatusBadRequest)
		return
	}

	for _, middleware := range siw.HandlerMiddlewares {
		middleware(c)
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
		siw.ErrorHandler(c, fmt.Errorf("Invalid format for parameter formatId: %s", err), http.StatusBadRequest)
		return
	}

	for _, middleware := range siw.HandlerMiddlewares {
		middleware(c)
	}

	siw.Handler.GetFormat(c, formatId)
}

// ListNodes operation middleware
func (siw *ServerInterfaceWrapper) ListNodes(c *gin.Context) {

	var err error

	// Parameter object where we will unmarshal all parameters from the context
	var params ListNodesParams

	// ------------- Optional query parameter "path" -------------

	err = runtime.BindQueryParameter("form", true, false, "path", c.Request.URL.Query(), &params.Path)
	if err != nil {
		siw.ErrorHandler(c, fmt.Errorf("Invalid format for parameter path: %s", err), http.StatusBadRequest)
		return
	}

	for _, middleware := range siw.HandlerMiddlewares {
		middleware(c)
	}

	siw.Handler.ListNodes(c, params)
}

// DeleteNode operation middleware
func (siw *ServerInterfaceWrapper) DeleteNode(c *gin.Context) {

	var err error

	// ------------- Path parameter "path" -------------
	var path Path

	err = runtime.BindStyledParameter("simple", false, "path", c.Param("path"), &path)
	if err != nil {
		siw.ErrorHandler(c, fmt.Errorf("Invalid format for parameter path: %s", err), http.StatusBadRequest)
		return
	}

	for _, middleware := range siw.HandlerMiddlewares {
		middleware(c)
	}

	siw.Handler.DeleteNode(c, path)
}

// UpdateNode operation middleware
func (siw *ServerInterfaceWrapper) UpdateNode(c *gin.Context) {

	var err error

	// ------------- Path parameter "path" -------------
	var path Path

	err = runtime.BindStyledParameter("simple", false, "path", c.Param("path"), &path)
	if err != nil {
		siw.ErrorHandler(c, fmt.Errorf("Invalid format for parameter path: %s", err), http.StatusBadRequest)
		return
	}

	for _, middleware := range siw.HandlerMiddlewares {
		middleware(c)
	}

	siw.Handler.UpdateNode(c, path)
}

// ListNodeRecords operation middleware
func (siw *ServerInterfaceWrapper) ListNodeRecords(c *gin.Context) {

	var err error

	// ------------- Path parameter "path" -------------
	var path Path

	err = runtime.BindStyledParameter("simple", false, "path", c.Param("path"), &path)
	if err != nil {
		siw.ErrorHandler(c, fmt.Errorf("Invalid format for parameter path: %s", err), http.StatusBadRequest)
		return
	}

	// Parameter object where we will unmarshal all parameters from the context
	var params ListNodeRecordsParams

	// ------------- Optional query parameter "format" -------------

	err = runtime.BindQueryParameter("form", true, false, "format", c.Request.URL.Query(), &params.Format)
	if err != nil {
		siw.ErrorHandler(c, fmt.Errorf("Invalid format for parameter format: %s", err), http.StatusBadRequest)
		return
	}

	// ------------- Optional query parameter "createdAfter" -------------

	err = runtime.BindQueryParameter("form", true, false, "createdAfter", c.Request.URL.Query(), &params.CreatedAfter)
	if err != nil {
		siw.ErrorHandler(c, fmt.Errorf("Invalid format for parameter createdAfter: %s", err), http.StatusBadRequest)
		return
	}

	// ------------- Optional query parameter "createdBefore" -------------

	err = runtime.BindQueryParameter("form", true, false, "createdBefore", c.Request.URL.Query(), &params.CreatedBefore)
	if err != nil {
		siw.ErrorHandler(c, fmt.Errorf("Invalid format for parameter createdBefore: %s", err), http.StatusBadRequest)
		return
	}

	// ------------- Optional query parameter "pageId" -------------

	err = runtime.BindQueryParameter("form", true, false, "pageId", c.Request.URL.Query(), &params.PageId)
	if err != nil {
		siw.ErrorHandler(c, fmt.Errorf("Invalid format for parameter pageId: %s", err), http.StatusBadRequest)
		return
	}

	// ------------- Optional query parameter "limit" -------------

	err = runtime.BindQueryParameter("form", true, false, "limit", c.Request.URL.Query(), &params.Limit)
	if err != nil {
		siw.ErrorHandler(c, fmt.Errorf("Invalid format for parameter limit: %s", err), http.StatusBadRequest)
		return
	}

	for _, middleware := range siw.HandlerMiddlewares {
		middleware(c)
	}

	siw.Handler.ListNodeRecords(c, path, params)
}

// PatchNodeRecords operation middleware
func (siw *ServerInterfaceWrapper) PatchNodeRecords(c *gin.Context) {

	var err error

	// ------------- Path parameter "path" -------------
	var path Path

	err = runtime.BindStyledParameter("simple", false, "path", c.Param("path"), &path)
	if err != nil {
		siw.ErrorHandler(c, fmt.Errorf("Invalid format for parameter path: %s", err), http.StatusBadRequest)
		return
	}

	for _, middleware := range siw.HandlerMiddlewares {
		middleware(c)
	}

	siw.Handler.PatchNodeRecords(c, path)
}

// CreateNodeRecords operation middleware
func (siw *ServerInterfaceWrapper) CreateNodeRecords(c *gin.Context) {

	var err error

	// ------------- Path parameter "path" -------------
	var path Path

	err = runtime.BindStyledParameter("simple", false, "path", c.Param("path"), &path)
	if err != nil {
		siw.ErrorHandler(c, fmt.Errorf("Invalid format for parameter path: %s", err), http.StatusBadRequest)
		return
	}

	for _, middleware := range siw.HandlerMiddlewares {
		middleware(c)
	}

	siw.Handler.CreateNodeRecords(c, path)
}

// Ping operation middleware
func (siw *ServerInterfaceWrapper) Ping(c *gin.Context) {

	for _, middleware := range siw.HandlerMiddlewares {
		middleware(c)
	}

	siw.Handler.Ping(c)
}

// Search operation middleware
func (siw *ServerInterfaceWrapper) Search(c *gin.Context) {

	for _, middleware := range siw.HandlerMiddlewares {
		middleware(c)
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
func RegisterHandlers(router *gin.Engine, si ServerInterface) *gin.Engine {
	return RegisterHandlersWithOptions(router, si, GinServerOptions{})
}

// RegisterHandlersWithOptions creates http.Handler with additional options
func RegisterHandlersWithOptions(router *gin.Engine, si ServerInterface, options GinServerOptions) *gin.Engine {

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

	router.GET(options.BaseURL+"/formats", wrapper.ListFormats)

	router.POST(options.BaseURL+"/formats", wrapper.CreateFormat)

	router.DELETE(options.BaseURL+"/formats/:formatId", wrapper.DeleteFormat)

	router.GET(options.BaseURL+"/formats/:formatId", wrapper.GetFormat)

	router.GET(options.BaseURL+"/nodes", wrapper.ListNodes)

	router.DELETE(options.BaseURL+"/nodes/:path", wrapper.DeleteNode)

	router.PUT(options.BaseURL+"/nodes/:path", wrapper.UpdateNode)

	router.GET(options.BaseURL+"/nodes/:path/records", wrapper.ListNodeRecords)

	router.PATCH(options.BaseURL+"/nodes/:path/records", wrapper.PatchNodeRecords)

	router.POST(options.BaseURL+"/nodes/:path/records", wrapper.CreateNodeRecords)

	router.GET(options.BaseURL+"/ping", wrapper.Ping)

	router.POST(options.BaseURL+"/search", wrapper.Search)

	return router
}

// Base64 encoded, gzipped, json marshaled Swagger object
var swaggerSpec = []string{

	"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",
}

// GetSwagger returns the content of the embedded swagger specification file
// or error if failed to decode
func decodeSpec() ([]byte, error) {
	zipped, err := base64.StdEncoding.DecodeString(strings.Join(swaggerSpec, ""))
	if err != nil {
		return nil, fmt.Errorf("error base64 decoding spec: %s", err)
	}
	zr, err := gzip.NewReader(bytes.NewReader(zipped))
	if err != nil {
		return nil, fmt.Errorf("error decompressing spec: %s", err)
	}
	var buf bytes.Buffer
	_, err = buf.ReadFrom(zr)
	if err != nil {
		return nil, fmt.Errorf("error decompressing spec: %s", err)
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
	var res = make(map[string]func() ([]byte, error))
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
	var resolvePath = PathToRawSpec("")

	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true
	loader.ReadFromURIFunc = func(loader *openapi3.Loader, url *url.URL) ([]byte, error) {
		var pathToFile = url.String()
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
