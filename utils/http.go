package utils

import (
	"bytes"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

// Interfaces Types
type IHttpResponse interface {
	Headers() http.Header
	Content() []byte
	StatusCode() int
}
type IHttpRequest interface {
	Get() IHttpResponse
	Head() IHttpResponse
	Error() error
	Url() string
	AddHeaders(map[string]string)
}
type IHttpRequestFactory interface {
	CreateRequest(url_str string) IHttpRequest
}

type HttpResponse struct {
	Content_    []byte
	Headers_    http.Header
	StatusCode_ int
}

func (resp *HttpResponse) Content() []byte {
	return resp.Content_
}

func (resp *HttpResponse) Headers() http.Header {
	return resp.Headers_
}

func (resp *HttpResponse) StatusCode() int {
	return resp.StatusCode_
}

type HttpRequest struct {
	Url_     string
	Headers_ map[string]string
	Error_   error
}

func (req *HttpRequest) Get() IHttpResponse {
	res, err := http.Get(req.Url_)
	var response HttpResponse
	if err == nil {
		response.Content_, err = ioutil.ReadAll(res.Body)
		response.Headers_ = res.Header
		response.StatusCode_ = res.StatusCode
		res.Body.Close()
		return &response
	}
	req.Error_ = err
	return nil
}

func (req *HttpRequest) Head() IHttpResponse {
	var response HttpResponse
	res, err := http.Head(req.Url_)
	if err == nil {
		response.Headers_ = res.Header
		response.StatusCode_ = res.StatusCode
		return &response
	}
	req.Error_ = err
	return nil
}

func (req *HttpRequest) Error() error {
	return req.Error_
}

func (req *HttpRequest) Url() string {
	return req.Url_
}

func (req *HttpRequest) AddHeaders(headers map[string]string) {
	for k, v := range headers {
		req.Headers_[k] = v
	}
}

type HttpRequestFactoryImpl struct {
}

func (f *HttpRequestFactoryImpl) CreateRequest(url_str string) IHttpRequest {
	var req HttpRequest
	req.Url_ = url_str
	return &req
}

//------------------------------------------------------------------
//						global variables
//------------------------------------------------------------------
var factory_impl HttpRequestFactoryImpl

var factory IHttpRequestFactory = &factory_impl

func CreateRequest(url_str string) IHttpRequest {
	return factory.CreateRequest(url_str)
}

//for tests
func SetHttpRequestFactory(new_factory IHttpRequestFactory) IHttpRequestFactory {
	var old_factory IHttpRequestFactory = factory
	factory = new_factory
	return old_factory
}

//------------------------------------------------------------------
//                        Functions
//------------------------------------------------------------------
func UnescapeUrlString(url_str string) (string, error) {
	return url.PathUnescape(url_str)
}

func GetUrlContent(req IHttpRequest) ([]byte, error) {
	var response IHttpResponse
	response = req.Get()
	if response != nil {
		return response.Content(), nil
	}
	return nil, req.Error()
}

func GetUrlHeaders(req IHttpRequest) (http.Header, error) {
	var response IHttpResponse
	response = req.Head()
	if response != nil {
		return response.Headers(), nil
	}
	return nil, req.Error()
}

func GetImageUrl(page_content []byte) (img_url string, err error) {
	var pattern string = "<table class='book_record'><tr><td width='210px' valign='top' rowspan='2'><img src=\"(.+)\"\\s+class='book'>"
	match, err := Extract_bytes(pattern, page_content)

	if err != nil {
		err = errors.New("HTML img source not found")
	} else {
		img_url = string(match)
		/*
			for index := 0; index < len(match); index += 1 {
				fmt.Printf("match[%d]: %+v\n", index, string(match[index]))
			} */
	}
	return
}

func AddUrlArguments(url string, args map[string]string) string {
	var buf bytes.Buffer
	var sep string = "&"

	buf.WriteString(url)
	if strings.Index(url, "?") == -1 {
		buf.WriteString("?")
		sep = ""
	}
	for k, v := range args {

		buf.WriteString(sep)
		buf.WriteString(k)
		buf.WriteString("=")
		buf.WriteString(v)
		sep = "&"
	}
	return buf.String()
}
