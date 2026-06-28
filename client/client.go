// Package client 提供面向 API 集成测试的通用 HTTP 客户端。
//
// 特性：
//   - 自动拼接 baseURL 与请求路径
//   - 支持 Token 自动注入 Authorization 头
//   - 函数式 RequestOption 配置查询参数、请求体、自定义头等
//   - DoJSON 封装「请求 + 状态码校验 + JSON 反序列化」
package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Client 封装 HTTP 客户端，持有 baseURL、超时配置及认证 Token。
type Client struct {
	baseURL    string       // 默认接口基地址
	router     Router       // 可选路径路由
	httpClient *http.Client // 底层 HTTP 客户端，带超时
	authHeader string       // 认证头名称，默认 Authorization
	authToken  string       // 当前认证 Token（含前缀，如 Bearer xxx）
}

// Router 根据请求路径解析目标 base URL。
type Router interface {
	Resolve(path string) string
}

// Response 封装 HTTP 响应，包含状态码、原始响应体及响应头。
type Response struct {
	StatusCode int
	Body       []byte
	Header     http.Header
}

// RequestOption 函数式选项，用于配置单次请求的参数。
type RequestOption func(*requestOptions)

// multipartFile 单个 multipart 文件字段。
type multipartFile struct {
	fieldName string
	filename  string
	content   []byte
}

// requestOptions 单次请求的内部配置。
type requestOptions struct {
	withAuth      bool              // 是否自动附加认证头，默认 true
	query         url.Values        // URL 查询参数
	body          any               // 请求体，将序列化为 JSON
	headers       map[string]string // 额外自定义请求头
	multipartFile *multipartFile    // multipart 文件上传（与 body 互斥）
}

// WithAuth 显式启用认证头注入（默认已启用）。
func WithAuth() RequestOption {
	return func(o *requestOptions) {
		o.withAuth = true
	}
}

// WithoutAuth 禁用认证头注入，用于登录等无需 Token 的请求。
func WithoutAuth() RequestOption {
	return func(o *requestOptions) {
		o.withAuth = false
	}
}

// WithQuery 设置 URL 查询参数。
func WithQuery(query url.Values) RequestOption {
	return func(o *requestOptions) {
		o.query = query
	}
}

// WithBody 设置 JSON 请求体（任意可序列化类型）。
func WithBody(body any) RequestOption {
	return func(o *requestOptions) {
		o.body = body
	}
}

// WithHeader 添加自定义请求头。
func WithHeader(key, value string) RequestOption {
	return func(o *requestOptions) {
		if o.headers == nil {
			o.headers = make(map[string]string)
		}
		o.headers[key] = value
	}
}

// WithMultipartFile 设置 multipart/form-data 文件字段（与 WithBody 互斥）。
func WithMultipartFile(fieldName, filename string, content []byte) RequestOption {
	return func(o *requestOptions) {
		o.multipartFile = &multipartFile{
			fieldName: fieldName,
			filename:  filename,
			content:   content,
		}
	}
}

// BaseURL 返回客户端配置的接口基地址。
func (c *Client) BaseURL() string {
	return c.baseURL
}

// New 创建 HTTP 客户端实例。baseURL 末尾斜杠会被去除。
func New(baseURL string, timeout time.Duration) *Client {
	return NewWithRouter(baseURL, timeout, nil)
}

// NewWithRouter 创建带路径路由的 HTTP 客户端。
func NewWithRouter(baseURL string, timeout time.Duration, router Router) *Client {
	return &Client{
		baseURL:    strings.TrimRight(baseURL, "/"),
		router:     router,
		authHeader: "Authorization",
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

// SetToken 设置认证 Token，后续请求将自动写入 authHeader。
func (c *Client) SetToken(token string) {
	c.authToken = token
}

// Token 返回当前持有的认证 Token。
func (c *Client) Token() string {
	return c.authToken
}

// Clone 复制客户端（共享 Token 与路由，独立 HTTP 连接池），供压测 worker 使用。
func (c *Client) Clone() *Client {
	if c == nil {
		return nil
	}
	nc := NewWithRouter(c.baseURL, c.httpClient.Timeout, c.router)
	nc.authHeader = c.authHeader
	nc.authToken = c.authToken
	return nc
}

// SetAuthHeader 自定义认证头名称（如 X-Api-Key）。
func (c *Client) SetAuthHeader(name string) {
	if name != "" {
		c.authHeader = name
	}
}

// Get 发送 GET 请求。
func (c *Client) Get(ctx context.Context, path string, opts ...RequestOption) (*Response, error) {
	return c.Do(ctx, http.MethodGet, path, opts...)
}

// Post 发送 POST 请求。
func (c *Client) Post(ctx context.Context, path string, opts ...RequestOption) (*Response, error) {
	return c.Do(ctx, http.MethodPost, path, opts...)
}

// Put 发送 PUT 请求。
func (c *Client) Put(ctx context.Context, path string, opts ...RequestOption) (*Response, error) {
	return c.Do(ctx, http.MethodPut, path, opts...)
}

// Delete 发送 DELETE 请求。
func (c *Client) Delete(ctx context.Context, path string, opts ...RequestOption) (*Response, error) {
	return c.Do(ctx, http.MethodDelete, path, opts...)
}

// Do 发送 HTTP 请求的核心方法，支持所有 HTTP 动词。
func (c *Client) Do(ctx context.Context, method, path string, opts ...RequestOption) (*Response, error) {
	options := requestOptions{withAuth: true}
	for _, opt := range opts {
		opt(&options)
	}

	fullURL, err := c.buildURL(path, options.query)
	if err != nil {
		return nil, err
	}

	var bodyReader io.Reader
	contentType := "application/json;charset=utf-8"
	if options.multipartFile != nil {
		var buf bytes.Buffer
		writer := multipart.NewWriter(&buf)
		part, err := writer.CreateFormFile(options.multipartFile.fieldName, options.multipartFile.filename)
		if err != nil {
			return nil, fmt.Errorf("create multipart form file: %w", err)
		}
		if _, err := part.Write(options.multipartFile.content); err != nil {
			return nil, fmt.Errorf("write multipart file content: %w", err)
		}
		if err := writer.Close(); err != nil {
			return nil, fmt.Errorf("close multipart writer: %w", err)
		}
		bodyReader = &buf
		contentType = writer.FormDataContentType()
	} else if options.body != nil {
		payload, err := json.Marshal(options.body)
		if err != nil {
			return nil, fmt.Errorf("marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(payload)
	}

	req, err := http.NewRequestWithContext(ctx, method, fullURL, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	if bodyReader != nil {
		req.Header.Set("Content-Type", contentType)
	}
	for key, value := range options.headers {
		req.Header.Set(key, value)
	}
	// 有 Token 且未禁用认证时，自动注入认证头
	if options.withAuth && c.authToken != "" {
		req.Header.Set(c.authHeader, c.authToken)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%s %s: %w", method, fullURL, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}

	return &Response{
		StatusCode: resp.StatusCode,
		Body:       body,
		Header:     resp.Header,
	}, nil
}

// DoJSON 发送请求并在状态码为 200 时将响应体反序列化到 dest。
// dest 为 nil 时仅校验状态码，不解析响应体。
func (c *Client) DoJSON(ctx context.Context, method, path string, dest any, opts ...RequestOption) error {
	resp, err := c.Do(ctx, method, path, opts...)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%s %s: unexpected status %d, body=%s", method, path, resp.StatusCode, string(resp.Body))
	}
	if dest == nil {
		return nil
	}
	if err := json.Unmarshal(resp.Body, dest); err != nil {
		return fmt.Errorf("decode json: %w, body=%s", err, string(resp.Body))
	}
	return nil
}

// buildURL 拼接 baseURL、路径及查询参数；若配置了路由则按 path 选择目标 base。
func (c *Client) buildURL(path string, query url.Values) (string, error) {
	baseURL := c.baseURL
	if c.router != nil {
		if resolved := c.router.Resolve(path); resolved != "" {
			baseURL = resolved
		}
	}

	normalizedPath := strings.TrimLeft(path, "/")
	endpoint, err := url.JoinPath(baseURL, normalizedPath)
	if err != nil {
		return "", fmt.Errorf("join url: %w", err)
	}
	if len(query) == 0 {
		return endpoint, nil
	}
	return endpoint + "?" + query.Encode(), nil
}
