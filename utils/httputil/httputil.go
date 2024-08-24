package httputil

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	ContentTypeJSON = "application/json"
	ContentTypeForm = "application/x-www-form-urlencoded"
)

// Client 定义了 HTTP 客户端的接口
type Client interface {
	Do(ctx context.Context, req *http.Request) (*http.Response, error)
}

// myClient 实现了 Client 接口
type myClient struct {
	cli *http.Client
}

// NewClient 创建一个新的 Client 实例
func NewClient(timeout time.Duration) Client {
	return &myClient{
		cli: &http.Client{
			Timeout: timeout,
		},
	}
}

// Do 发送单个 HTTP 请求
func (c *myClient) Do(ctx context.Context, req *http.Request) (*http.Response, error) {
	setTraceIDInHeader(ctx, req)
	req = req.WithContext(ctx)
	resp, err := c.cli.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%s %s%s: %w", req.Method, req.URL.Host, req.URL.Path, err)
	}
	if resp.StatusCode != http.StatusOK {
		return resp, fmt.Errorf("%s %s%s: 意外的状态码 %d", req.Method, req.URL.Host, req.URL.Path, resp.StatusCode)
	}
	return resp, nil
}

// setTraceIDInHeader 将 trace ID 添加到请求头中
func setTraceIDInHeader(ctx context.Context, req *http.Request) {
	if traceID := ctx.Value("traceID"); traceID != nil {
		if tid, ok := traceID.(string); ok && tid != "" {
			req.Header.Set("X-TRACE-ID", tid)
		}
	}
}

// NewGetRequest 创建一个新的 GET 请求
func NewGetRequest(url string, params map[string]interface{}, headers ...map[string]string) (*http.Request, error) {
	if params != nil {
		u, err := buildURLWithParams(url, params)
		if err != nil {
			return nil, err
		}
		url = u
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	if len(headers) > 0 {
		setHeaders(req, headers[0])
	}
	return req, nil
}

// NewFormPostRequest 创建一个新的 POST 请求，使用表单数据
func NewFormPostRequest(urlStr string, params map[string]interface{}, headers ...map[string]string) (*http.Request, error) {
	formData := make([]byte, 0)
	for key, value := range params {
		switch v := value.(type) {
		case string:
			formData = append(formData, []byte(key+"="+url.QueryEscape(v)+"&")...)
		case []string:
			for _, s := range v {
				formData = append(formData, []byte(key+"="+url.QueryEscape(s)+"&")...)
			}
		default:
			formData = append(formData, []byte(key+"="+url.QueryEscape(fmt.Sprint(v))+"&")...)
		}
	}
	// 移除最后一个 '&'
	if len(formData) > 0 {
		formData = formData[:len(formData)-1]
	}

	req, err := http.NewRequest(http.MethodPost, urlStr, bytes.NewReader(formData))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", ContentTypeForm)

	if len(headers) > 0 {
		setHeaders(req, headers[0])
	}
	return req, nil
}

// NewJSONPostRequest 创建一个新的 POST 请求，使用 JSON 数据
func NewJSONPostRequest(url string, params map[string]interface{}, headers ...map[string]string) (*http.Request, error) {
	body, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, url, strings.NewReader(string(body)))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", ContentTypeJSON)

	if len(headers) > 0 {
		setHeaders(req, headers[0])
	}
	return req, nil
}

// Get 执行 GET 请求
func Get(ctx context.Context, url string, params map[string]interface{}, options ...map[string]interface{}) (*http.Response, error) {
	timeout := getTimeout(options...)
	client := NewClient(timeout)
	req, err := NewGetRequest(url, params)
	if err != nil {
		return nil, err
	}
	return client.Do(ctx, req)
}

// Post 执行 POST 请求，使用表单数据
func Post(ctx context.Context, url string, params map[string]interface{}, options ...map[string]interface{}) (*http.Response, error) {
	timeout := getTimeout(options...)
	client := NewClient(timeout)
	req, err := NewFormPostRequest(url, params)
	if err != nil {
		return nil, err
	}
	return client.Do(ctx, req)
}

// PostJSON 执行 POST 请求，使用 JSON 数据
func PostJSON(ctx context.Context, url string, params map[string]interface{}, options ...map[string]interface{}) (*http.Response, error) {
	timeout := getTimeout(options...)
	client := NewClient(timeout)
	req, err := NewJSONPostRequest(url, params)
	if err != nil {
		return nil, err
	}
	return client.Do(ctx, req)
}

// Request 执行指定方法的 HTTP 请求
func Request(ctx context.Context, method, url string, params map[string]interface{}, options ...map[string]interface{}) (*http.Response, error) {
	timeout := getTimeout(options...)
	client := NewClient(timeout)

	var req *http.Request
	var err error

	switch strings.ToUpper(method) {
	case http.MethodPost:
		req, err = NewFormPostRequest(url, params)
	case "POST/JSON":
		req, err = NewJSONPostRequest(url, params)
	default:
		req, err = NewGetRequest(url, params)
	}

	if err != nil {
		return nil, err
	}
	return client.Do(ctx, req)
}

// DealResponse 读取并关闭响应体
func DealResponse(resp *http.Response) ([]byte, error) {
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

// setHeaders 为请求设置头部
func setHeaders(req *http.Request, headers map[string]string) {
	for k, v := range headers {
		req.Header.Set(k, v)
	}
}

// buildURLWithParams 构建带查询参数的 URL
func buildURLWithParams(baseURL string, params map[string]interface{}) (string, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return "", err
	}

	q := u.Query()
	for k, v := range params {
		q.Set(k, fmt.Sprint(v))
	}
	u.RawQuery = q.Encode()

	return u.String(), nil
}

// getTimeout 从选项中提取超时时间
func getTimeout(options ...map[string]interface{}) time.Duration {
	var timeout int
	if len(options) > 0 {
		if t, ok := options[0]["timeout"]; ok {
			timeout, _ = t.(int)
		}
	}
	if timeout <= 0 {
		timeout = 30
	}
	return time.Duration(timeout) * time.Second
}
