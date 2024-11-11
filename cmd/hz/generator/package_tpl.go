/*
 * Copyright 2022 CloudWeGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package generator

var (
	routerTplName           = "router.go"
	middlewareTplName       = "middleware.go"
	middlewareSingleTplName = "middleware_single.go"
	handlerTplName          = "handler.go"
	handlerSingleTplName    = "handler_single.go"
	modelTplName            = "model.go"
	registerTplName         = "register.go"
	clientTplName           = "client.go"       // generate a default client for server
	hertzClientTplName      = "hertz_client.go" // underlying client for client command
	idlClientName           = "idl_client.go"   // client of service for quick call

	insertPointNew        = "//INSERT_POINT: DO NOT DELETE THIS LINE!"
	insertPointPatternNew = `//INSERT_POINT\: DO NOT DELETE THIS LINE\!`
)

var templateNameSet = map[string]string{
	routerTplName:           routerTplName,
	middlewareTplName:       middlewareTplName,
	middlewareSingleTplName: middlewareSingleTplName,
	handlerTplName:          handlerTplName,
	handlerSingleTplName:    handlerSingleTplName,
	modelTplName:            modelTplName,
	registerTplName:         registerTplName,
	clientTplName:           clientTplName,
	hertzClientTplName:      hertzClientTplName,
	idlClientName:           idlClientName,
}

func IsDefaultPackageTpl(name string) bool {
	if _, exist := templateNameSet[name]; exist {
		return true
	}

	return false
}

var defaultPkgConfig = TemplateConfig{
	Layouts: []Template{
		{
			Path:   defaultHandlerDir + sp + handlerTplName,
			Delims: [2]string{"{{", "}}"},
			Body: `// Code generated by hertz generator.

package {{.PackageName}}

import (
	"context"

	"github.com/weightwave/gocommon/errs"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/weightwave/gocommon/logs"

{{- range $k, $v := .Imports}}
	{{$k}} "{{$v.Package}}"
{{- end}}
)

{{range $_, $MethodInfo := .Methods}}
{{$MethodInfo.Comment}}
func {{$MethodInfo.Name}}(ctx context.Context, c *app.RequestContext) { 
	var err error
	{{if ne $MethodInfo.RequestTypeName "" -}}
	var req {{$MethodInfo.RequestTypeName}}
	defer func(){
		if err != nil {
			logs.CtxErrorf(ctx, "%s", errs.Wrap(err, "{{$MethodInfo.Name}} error").Error())
			c.Error(err)
		}
	}()
	err = c.BindAndValidate(&req)
	if err != nil {
		c.String(consts.StatusBadRequest, err.Error())
		return
	}
	{{end}}

	// TODO params check

	var resp *{{$MethodInfo.ReturnTypeName}}
	// resp, err = _service.{{$MethodInfo.Name}}(ctx, &req)
	if err != nil {
		return
	}
	resp.Status = &base.ResponseStatus{}
	c.{{.Serializer}}(consts.StatusOK, resp)
}
{{end}}
			`,
		},
		{
			Path:   defaultRouterDir + sp + routerTplName,
			Delims: [2]string{"{{", "}}"},
			Body: `// Code generated by hertz generator. DO NOT EDIT.

package {{$.PackageName}}

import (
	"github.com/cloudwego/hertz/pkg/app/server"

    {{- range $k, $v := .HandlerPackages}}
        {{$k}} "{{$v}}"
    {{- end}}
)

/*
 This file will register all the routes of the services in the master idl.
 And it will update automatically when you use the "update" command for the idl.
 So don't modify the contents of the file, or your code will be deleted when it is updated.
 */

{{define "g"}}
{{- if eq .Path "/"}}r
{{- else}}{{.GroupName}}{{end}}
{{- end}}

{{define "G"}}
{{- if ne .Handler ""}}
	{{- .GroupName}}.{{.HttpMethod}}("{{.Path}}", append({{.HandlerMiddleware}}Mw(), {{.Handler}})...)
{{- end}}
{{- if ne (len .Children) 0}}
{{.MiddleWare}} := {{template "g" .}}.Group("{{.Path}}", {{.GroupMiddleware}}Mw()...)
{{- end}}
{{- range $_, $router := .Children}}
{{- if ne .Handler ""}}
	{{template "G" $router}}
{{- else}}
	{	{{template "G" $router}}
	}
{{- end}}
{{- end}}
{{- end}}

// Register register routes based on the IDL 'api.${HTTP Method}' annotation.
func Register(r *server.Hertz) {
{{template "G" .Router}}
}

		`,
		},
		{
			Path: defaultRouterDir + sp + registerTplName,
			Body: `// Code generated by hertz generator. DO NOT EDIT.

package {{.PackageName}}

import (
	"github.com/cloudwego/hertz/pkg/app/server"
	{{$.DepPkgAlias}} "{{$.DepPkg}}"
)

// GeneratedRegister registers routers generated by IDL.
func GeneratedRegister(r *server.Hertz){
	` + insertPointNew + `
	{{$.DepPkgAlias}}.Register(r)
}
`,
		},
		// Model tpl is imported by model generator. Here only decides model directory.
		{
			Path: defaultModelDir + sp + modelTplName,
			Body: ``,
		},
		{
			Path:   defaultRouterDir + sp + middlewareTplName,
			Delims: [2]string{"{{", "}}"},
			Body: `// Code generated by hertz generator.

package {{$.PackageName}}

import (
	"github.com/cloudwego/hertz/pkg/app"
)

{{define "M"}}
{{- if ne .Children.Len 0}}
func {{.GroupMiddleware}}Mw() []app.HandlerFunc {
	// your code...
	return nil
}
{{end}}
{{- if ne .Handler ""}}
func {{.HandlerMiddleware}}Mw() []app.HandlerFunc {
	// your code...
	return nil
}
{{end}}
{{range $_, $router := $.Children}}{{template "M" $router}}{{end}}
{{- end}}

{{template "M" .Router}}

		`,
		},
		{
			Path:   defaultClientDir + sp + clientTplName,
			Delims: [2]string{"{{", "}}"},
			Body: `// Code generated by hertz generator.

package {{$.PackageName}}

import (
    "github.com/cloudwego/hertz/pkg/app/client"
	"github.com/cloudwego/hertz/pkg/common/config"
)

type {{.ServiceName}}Client struct {
	client * client.Client
}

func New{{.ServiceName}}Client(opt ...config.ClientOption) (*{{.ServiceName}}Client, error) {
	c, err := client.NewClient(opt...)
	if err != nil {
		return nil, err
	}

	return &{{.ServiceName}}Client{
		client: c,
	}, nil
}
		`,
		},
		{
			Path:   defaultHandlerDir + sp + handlerSingleTplName,
			Delims: [2]string{"{{", "}}"},
			Body: `
{{.Comment}}
func {{.Name}}(ctx context.Context, c *app.RequestContext) { 
	var err error
	{{if ne .RequestTypeName "" -}}
	var req {{.RequestTypeName}}
	defer func() {
		if err != nil {
			logs.CtxErrorf(ctx, "%s", errs.Wrap(err, "{{.Name}} error").Error())
			c.Error(err)
		}
	}()
	err = c.BindAndValidate(&req)
	if err != nil {
		c.String(consts.StatusBadRequest, err.Error())
		return
	}
	{{end}}

	//TODO params check

	var resp *{{.ReturnTypeName}}
	// resp, err = _service.{{.Name}}(ctx, &req)
	if err != nil {
		return
	}
	resp.Status = &base.ResponseStatus{}
	c.{{.Serializer}}(consts.StatusOK, resp)
}
`,
		},
		{
			Path:   defaultRouterDir + sp + middlewareSingleTplName,
			Delims: [2]string{"{{", "}}"},
			Body: `
func {{.MiddleWare}}Mw() []app.HandlerFunc {
	// your code...
	return nil
}
`,
		},
		{
			Path:   defaultRouterDir + sp + hertzClientTplName,
			Delims: [2]string{"{{", "}}"},
			Body:   hertzClientTpl,
		},
		{
			Path:   defaultRouterDir + sp + idlClientName,
			Delims: [2]string{"{{", "}}"},
			Body:   idlClientTpl,
		},
	},
}

var hertzClientTpl = `// Code generated by hz.

package {{.PackageName}}

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"regexp"
	"strings"

	hertz_client "github.com/cloudwego/hertz/pkg/app/client"
	"github.com/cloudwego/hertz/pkg/common/config"
	"github.com/cloudwego/hertz/pkg/common/errors"
	"github.com/cloudwego/hertz/pkg/protocol"
	"github.com/cloudwego/hertz/pkg/protocol/client"
)

type use interface {
	Use(mws ...hertz_client.Middleware)
}

// Definition of global data and types.
type ResponseResultDecider func(statusCode int, rawResponse *protocol.Response) (isError bool)

type (
	bindRequestBodyFunc func(c *cli, r *request) (contentType string, body io.Reader, err error)
	beforeRequestFunc   func(*cli, *request) error
	afterResponseFunc   func(*cli, *response) error
)

var (
	hdrContentTypeKey     = http.CanonicalHeaderKey("Content-Type")
	hdrContentEncodingKey = http.CanonicalHeaderKey("Content-Encoding")

	plainTextType   = "text/plain; charset=utf-8"
	jsonContentType = "application/json; charset=utf-8"
	formContentType = "multipart/form-data"

	jsonCheck = regexp.MustCompile(` + "`(?i:(application|text)/(json|.*\\+json|json\\-.*)(; |$))`)\n" +
	`xmlCheck  = regexp.MustCompile(` + "`(?i:(application|text)/(xml|.*\\+xml)(; |$))`)\n" +
	`
)

// Configuration of client
type Option struct {
	f func(*Options)
}

type Options struct {
	hostUrl               string
	doer                  client.Doer
	header                http.Header
	requestBodyBind       bindRequestBodyFunc
	responseResultDecider ResponseResultDecider
	middlewares           []hertz_client.Middleware
	clientOption          []config.ClientOption
}

func getOptions(ops ...Option) *Options {
	opts := &Options{}
	for _, do := range ops {
		do.f(opts)
	}
	return opts
}

// WithHertzClientOption is used to pass configuration for the hertz client
func WithHertzClientOption(opt ...config.ClientOption) Option {
	return Option{func(op *Options) {
		op.clientOption = append(op.clientOption, opt...)
	}}
}

// WithHertzClientMiddleware is used to register the middleware for the hertz client
func WithHertzClientMiddleware(mws ...hertz_client.Middleware) Option {
	return Option{func(op *Options) {
		op.middlewares = append(op.middlewares, mws...)
	}}
}

// WithHertzClient is used to register a custom hertz client
func WithHertzClient(client client.Doer) Option {
	return Option{func(op *Options) {
		op.doer = client
	}}
}

// WithHeader is used to add the default header, which is carried by every request
func WithHeader(header http.Header) Option {
	return Option{func(op *Options) {
		op.header = header
	}}
}

// WithResponseResultDecider configure custom deserialization of http response to response struct
func WithResponseResultDecider(decider ResponseResultDecider) Option {
	return Option{func(op *Options) {
		op.responseResultDecider = decider
	}}
}

func withHostUrl(HostUrl string) Option {
	return Option{func(op *Options) {
		op.hostUrl = HostUrl
	}}
}

// underlying client
type cli struct {
	hostUrl               string
	doer                  client.Doer
	header                http.Header
	bindRequestBody       bindRequestBodyFunc
	responseResultDecider ResponseResultDecider

	beforeRequest []beforeRequestFunc
	afterResponse []afterResponseFunc
}

func (c *cli) Use(mws ...hertz_client.Middleware) error {
	u, ok := c.doer.(use)
	if !ok {
		return errors.NewPublic("doer does not support middleware, choose the right doer.")
	}
	u.Use(mws...)
	return nil
}

func newClient(opts *Options) (*cli, error) {
	if opts.requestBodyBind == nil {
		opts.requestBodyBind = defaultRequestBodyBind
	}
	if opts.responseResultDecider == nil {
		opts.responseResultDecider = defaultResponseResultDecider
	}
	if opts.doer == nil {
		cli, err := hertz_client.NewClient(opts.clientOption...)
		if err != nil {
			return nil, err
		}
		opts.doer = cli
	}

	c := &cli{
		hostUrl:               opts.hostUrl,
		doer:                  opts.doer,
		header:                opts.header,
		bindRequestBody:       opts.requestBodyBind,
		responseResultDecider: opts.responseResultDecider,
		beforeRequest: []beforeRequestFunc{
			parseRequestURL,
			parseRequestHeader,
			createHTTPRequest,
		},
		afterResponse: []afterResponseFunc{
			parseResponseBody,
		},
	}

	if len(opts.middlewares) != 0 {
		if err := c.Use(opts.middlewares...); err != nil {
			return nil, err
		}
	}
	return c, nil
}

func (c *cli) execute(req *request) (*response, error) {
	var err error
	for _, f := range c.beforeRequest {
		if err = f(c, req); err != nil {
			return nil, err
		}
	}

	if hostHeader := req.header.Get("Host"); hostHeader != "" {
		req.rawRequest.Header.SetHost(hostHeader)
	}

	resp := protocol.Response{}

	err = c.doer.Do(req.ctx, req.rawRequest, &resp)

	response := &response{
		request:     req,
		rawResponse: &resp,
	}

	if err != nil {
		return response, err
	}

	body, err := resp.BodyE()
	if err != nil {
		return nil, err
	}

	if strings.EqualFold(resp.Header.Get(hdrContentEncodingKey), "gzip") && resp.Header.ContentLength() != 0 {
		body, err = resp.BodyGunzip()
		if err != nil {
			return nil, err
		}
	}

	response.bodyByte = body

	response.size = int64(len(response.bodyByte))

	// Apply Response middleware
	for _, f := range c.afterResponse {
		if err = f(c, response); err != nil {
			break
		}
	}

	return response, err
}

// r get request
func (c *cli) r() *request {
	return &request{
		queryParam:     url.Values{},
		header:         http.Header{},
		pathParam:      map[string]string{},
		formParam:      map[string]string{},
		fileParam:      map[string]string{},
		client:         c,
		queryEnumAsInt: {{.Config.QueryEnumAsInt}},
	}
}

type response struct {
	request     *request
	rawResponse *protocol.Response

	bodyByte []byte
	size     int64
}

// statusCode method returns the HTTP status code for the executed request.
func (r *response) statusCode() int {
	if r.rawResponse == nil {
		return 0
	}

	return r.rawResponse.StatusCode()
}

// body method returns HTTP response as []byte array for the executed request.
func (r *response) body() []byte {
	if r.rawResponse == nil {
		return []byte{}
	}
	return r.bodyByte
}

// Header method returns the response headers
func (r *response) header() http.Header {
	if r.rawResponse == nil {
		return http.Header{}
	}
	h := http.Header{}
	r.rawResponse.Header.VisitAll(func(key, value []byte) {
		h.Add(string(key), string(value))
	})

	return h
}

type request struct {
	client         *cli
	url            string
	method         string
	queryEnumAsInt bool
	queryParam     url.Values
	header         http.Header
	pathParam      map[string]string
	formParam      map[string]string
	fileParam      map[string]string
	bodyParam      interface{}
	rawRequest     *protocol.Request
	ctx            context.Context
	requestOptions []config.RequestOption
	result         interface{}
	Error          interface{}
}

func (r *request) setContext(ctx context.Context) *request {
	r.ctx = ctx
	return r
}

func (r *request) context() context.Context {
	return r.ctx
}

func (r *request) setHeader(header, value string) *request {
	r.header.Set(header, value)
	return r
}

func (r *request) addHeader(header, value string) *request {
	r.header.Add(header, value)
	return r
}

func (r *request) addHeaders(params map[string]string) *request {
	for k, v := range params {
		r.addHeader(k, v)
	}
	return r
}


func (r *request) setQueryParam(param string, value interface{}) *request {
	v := reflect.ValueOf(value)
	switch v.Kind() {
	case reflect.Slice, reflect.Array:
		for index := 0; index < v.Len(); index++ {
			if r.queryEnumAsInt && (v.Index(index).Kind() == reflect.Int32 || v.Index(index).Kind() == reflect.Int64) {
				r.queryParam.Add(param, fmt.Sprintf("%d", v.Index(index).Interface()))
			} else {
				r.queryParam.Add(param, fmt.Sprint(v.Index(index).Interface()))
			}
		}
	case reflect.Int32, reflect.Int64:
		if r.queryEnumAsInt {
			r.queryParam.Add(param, fmt.Sprintf("%d", v.Interface()))
		} else {
			r.queryParam.Add(param, fmt.Sprint(v))
		}
	default:
		r.queryParam.Set(param, fmt.Sprint(v))
	}
	return r
}

func (r *request) setResult(res interface{}) *request {
	r.result = res
	return r
}

func (r *request) setError(err interface{}) *request {
	r.Error = err
	return r
}

func (r *request) setHeaders(headers map[string]string) *request {
	for h, v := range headers {
		r.setHeader(h, v)
	}

	return r
}

func (r *request) setQueryParams(params map[string]interface{}) *request {
	for p, v := range params {
		r.setQueryParam(p, v)
	}

	return r
}

func (r *request) setPathParams(params map[string]string) *request {
	for p, v := range params {
		r.pathParam[p] = v
	}
	return r
}

func (r *request) setFormParams(params map[string]string) *request {
	for p, v := range params {
		r.formParam[p] = v
	}
	return r
}

func (r *request) setFormFileParams(params map[string]string) *request {
	for p, v := range params {
		r.fileParam[p] = v
	}
	return r
}

func (r *request) setBodyParam(body interface{}) *request {
	r.bodyParam = body
	return r
}

func (r *request) setRequestOption(option ...config.RequestOption) *request {
	r.requestOptions = append(r.requestOptions, option...)
	return r
}

func (r *request) execute(method, url string) (*response, error) {
	r.method = method
	r.url = url
	return r.client.execute(r)
}

func parseRequestURL(c *cli, r *request) error {
	if len(r.pathParam) > 0 {
		for p, v := range r.pathParam {
			r.url = strings.Replace(r.url, ":"+p, url.PathEscape(v), -1)
		}
	}

	// Parsing request URL
	reqURL, err := url.Parse(r.url)
	if err != nil {
		return err
	}

	// If request.URL is relative path then added c.HostURL into
	// the request URL otherwise request.URL will be used as-is
	if !reqURL.IsAbs() {
		r.url = reqURL.String()
		if len(r.url) > 0 && r.url[0] != '/' {
			r.url = "/" + r.url
		}

		reqURL, err = url.Parse(c.hostUrl + r.url)
		if err != nil {
			return err
		}
	}

	// Adding Query Param
	query := make(url.Values)

	for k, v := range r.queryParam {
		// remove query param from client level by key
		// since overrides happens for that key in the request
		query.Del(k)
		for _, iv := range v {
			query.Add(k, iv)
		}
	}

	if len(query) > 0 {
		if isStringEmpty(reqURL.RawQuery) {
			reqURL.RawQuery = query.Encode()
		} else {
			reqURL.RawQuery = reqURL.RawQuery + "&" + query.Encode()
		}
	}

	r.url = reqURL.String()

	return nil
}

func isStringEmpty(str string) bool {
	return len(strings.TrimSpace(str)) == 0
}

func parseRequestHeader(c *cli, r *request) error {
	hdr := make(http.Header)
	if c.header != nil {
		for k := range c.header {
			hdr[k] = append(hdr[k], c.header[k]...)
		}
	}

	for k := range r.header {
		hdr.Del(k)
		hdr[k] = append(hdr[k], r.header[k]...)
	}

	if len(r.formParam) != 0 || len(r.fileParam) != 0 {
		hdr.Add(hdrContentTypeKey, formContentType)
	}

	r.header = hdr
	return nil
}

// detectContentType method is used to figure out "request.Body" content type for request header
func detectContentType(body interface{}) string {
	contentType := plainTextType
	kind := reflect.Indirect(reflect.ValueOf(body)).Kind()
	switch kind {
	case reflect.Struct, reflect.Map:
		contentType = jsonContentType
	case reflect.String:
		contentType = plainTextType
	default:
		if b, ok := body.([]byte); ok {
			contentType = http.DetectContentType(b)
		} else if kind == reflect.Slice {
			contentType = jsonContentType
		}
	}

	return contentType
}

func defaultRequestBodyBind(c *cli, r *request) (contentType string, body io.Reader, err error) {
	if !isPayloadSupported(r.method) {
		return
	}
	var bodyBytes []byte
	contentType = r.header.Get(hdrContentTypeKey)
	if isStringEmpty(contentType) {
		contentType = detectContentType(r.bodyParam)
		r.header.Set(hdrContentTypeKey, contentType)
	}
	kind := reflect.Indirect(reflect.ValueOf(r.bodyParam)).Kind()
	if isJSONType(contentType) &&
		(kind == reflect.Struct || kind == reflect.Map || kind == reflect.Slice) {
		bodyBytes, err = json.Marshal(r.bodyParam)
	} else if isXMLType(contentType) && (kind == reflect.Struct) {
		bodyBytes, err = xml.Marshal(r.bodyParam)
	}
	if err != nil {
		return
	}
	return contentType, strings.NewReader(string(bodyBytes)), nil
}

func isPayloadSupported(m string) bool {
	return !(m == http.MethodHead || m == http.MethodOptions || m == http.MethodGet || m == http.MethodDelete)
}

func createHTTPRequest(c *cli, r *request) (err error) {
	contentType, body, err := c.bindRequestBody(c, r)
	if !isStringEmpty(contentType) {
		r.header.Set(hdrContentTypeKey, contentType)
	}
	if err == nil {
		r.rawRequest = protocol.NewRequest(r.method, r.url, body)
		if contentType == formContentType && isPayloadSupported(r.method) {
			if r.rawRequest.IsBodyStream() {
				r.rawRequest.ResetBody()
			}
			r.rawRequest.SetMultipartFormData(r.formParam)
			r.rawRequest.SetFiles(r.fileParam)
		}
		for key, values := range r.header {
			for _, val := range values {
				r.rawRequest.Header.Add(key, val)
			}
		}
		r.rawRequest.SetOptions(r.requestOptions...)
	}
	return err
}

func silently(_ ...interface{}) {}

// defaultResponseResultDecider method returns true if HTTP status code >= 400 otherwise false.
func defaultResponseResultDecider(statusCode int, rawResponse *protocol.Response) bool {
	return statusCode > 399
}

// IsJSONType method is to check JSON content type or not
func isJSONType(ct string) bool {
	return jsonCheck.MatchString(ct)
}

// IsXMLType method is to check XML content type or not
func isXMLType(ct string) bool {
	return xmlCheck.MatchString(ct)
}

func parseResponseBody(c *cli, res *response) (err error) {
	if res.statusCode() == http.StatusNoContent {
		return
	}
	// Handles only JSON or XML content type
	ct := res.header().Get(hdrContentTypeKey)

	isError := c.responseResultDecider(res.statusCode(), res.rawResponse)
	if isError {
		if res.request.Error != nil {
			if isJSONType(ct) || isXMLType(ct) {
				err = unmarshalContent(ct, res.bodyByte, res.request.Error)
			}
		} else {
			jsonByte, jsonErr := json.Marshal(map[string]interface{}{
				"status_code": res.rawResponse.StatusCode(),
				"body":        string(res.bodyByte),
			})
			if jsonErr != nil {
				return jsonErr
			}
			err = fmt.Errorf(string(jsonByte))
		}
	} else if res.request.result != nil {
		if isJSONType(ct) || isXMLType(ct) {
			err = unmarshalContent(ct, res.bodyByte, res.request.result)
			return
		}
	}
	return
}

// unmarshalContent content into object from JSON or XML
func unmarshalContent(ct string, b []byte, d interface{}) (err error) {
	if isJSONType(ct) {
		err = json.Unmarshal(b, d)
	} else if isXMLType(ct) {
		err = xml.Unmarshal(b, d)
	}

	return
}

`

var idlClientTpl = `// Code generated by hertz generator.

package {{.PackageName}}

import (
	"context"
	"fmt"
	"sync"

	"github.com/cloudwego/hertz/pkg/common/config"
	"github.com/cloudwego/hertz/pkg/protocol"
	"github.com/weightwave/gocommon/envs"
	"github.com/weightwave/gocommon/errs"
	"github.com/weightwave/gocommon/hertz_mw"
	"github.com/weightwave/gocommon/logs"
	"github.com/weightwave/gocommon/json_utils"

{{- range $k, $v := .Imports}}
	{{$k}} "{{$v.Package}}"
{{- end}}
)

// unused protection
var (
	_ = fmt.Formatter(nil)
	nacosDisable bool
)

type Client interface {
	{{range $_, $MethodInfo := .ClientMethods}}
		{{$MethodInfo.Name}}(context context.Context, req *{{$MethodInfo.RequestTypeName}}, reqOpt ...config.RequestOption) (resp *{{$MethodInfo.ReturnTypeName}}, rawResponse *protocol.Response, err error)
	{{end}}
}

type {{.ServiceName}}Client struct {
	client *cli
}

func New{{.ServiceName}}Client(hostUrl string, ops ...Option) (Client, error) {
	for _, mw := range hertz_mw.ClientMws() {
		ops = append(ops, WithHertzClientMiddleware(mw))
	}
	ops = append(ops, withHostUrl(hostUrl))
	opts := getOptions(ops...)
	cli, err := newClient(opts)
	if err != nil {
		return nil, err
	}
	return &{{.ServiceName}}Client{
		client: cli,
	}, nil
}

{{range $_, $MethodInfo := .ClientMethods}}
func (s *{{$.ServiceName}}Client) {{$MethodInfo.Name}}(context context.Context, req *{{$MethodInfo.RequestTypeName}}, reqOpt ...config.RequestOption) (resp *{{$MethodInfo.ReturnTypeName}}, rawResponse *protocol.Response, err error) {
	logs.CtxDebugf(context, "{{$.ServiceName}} {{$MethodInfo.Name}} req: %s", json_utils.GetJsonStr(req))
	if !nacosDisable {
		reqOpt = append(reqOpt, config.WithSD(true))
	}
	httpResp := &{{$MethodInfo.ReturnTypeName}}{}
	ret, err := s.client.r().
		setContext(context).
		setQueryParams(map[string]interface{}{
			{{$MethodInfo.QueryParamsCode}}
		}).
		setPathParams(map[string]string{
			{{$MethodInfo.PathParamsCode}}
		}).
		addHeaders(map[string]string{
			{{$MethodInfo.HeaderParamsCode}}
		}).
		setFormParams(map[string]string{
			{{$MethodInfo.FormValueCode}}
		}).
		setFormFileParams(map[string]string{
			{{$MethodInfo.FormFileCode}}
		}).
		{{$MethodInfo.BodyParamsCode}}
		setRequestOption(reqOpt...).
		setResult(httpResp).
		execute("{{if EqualFold $MethodInfo.HTTPMethod "Any"}}POST{{else}}{{ $MethodInfo.HTTPMethod }}{{end}}", "{{$MethodInfo.Path}}")
	if err != nil {
		return nil, nil, err
	}
    
	resp = httpResp
	rawResponse = ret.rawResponse
	if resp.Status == nil {
		return nil, nil, errs.New(500, "Invalid Response, No Status")
	}
	logs.CtxDebugf(context, "{{$.ServiceName}} {{$MethodInfo.Name}} resp: %s", json_utils.GetJsonStr(resp))
	if resp.Status.Code != 0 {
		return nil, nil, errs.New(int(resp.GetStatus().GetCode()), resp.GetStatus().GetMessage())
	}
	return resp, rawResponse, nil
}
{{end}}

var defaultClient Client
var once sync.Once

func Init(){
	once.Do(func() {
		envSuffix := envs.GetEnv()
		if envSuffix == "local" || envSuffix == "" {
			envSuffix = "dev"
		}
		defaultClient, _ = New{{.ServiceName}}Client("http://nacos-{{.BaseDomain}}-" + envSuffix)
	})
}

{{range $_, $MethodInfo := .ClientMethods}}
func {{$MethodInfo.Name}}(context context.Context, req *{{$MethodInfo.RequestTypeName}}, reqOpt ...config.RequestOption) (resp *{{$MethodInfo.ReturnTypeName}}, rawResponse *protocol.Response, err error) {
	Init()
	return defaultClient.{{$MethodInfo.Name}}(context, req, reqOpt...)
}
{{end}}
`
