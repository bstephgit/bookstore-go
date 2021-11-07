package utils_test

import (
	"github.com/bookstore-go/utils"
	"bytes"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

//-------------------------------------------------------------------------------------
//										MOCKS
//-------------------------------------------------------------------------------------

type HttpRequestTest struct {
	NbGetCall          int
	NbHeadCall         int
	ResponseContent    []byte
	ResponseStatusCode int
	ResponseHeaders    http.Header
	Request            utils.HttpRequest
}

type HttpRequestFactoryTest struct {
	Req       HttpRequestTest
	NbCreated int
}

func (factory *HttpRequestFactoryTest) CreateRequest(url_str string) utils.IHttpRequest {
	factory.Req.Request.Url_ = url_str
	return &factory.Req
}

func (req *HttpRequestTest) Get() utils.IHttpResponse {
	var response utils.HttpResponse
	response.Content_ = req.ResponseContent
	response.Headers_ = req.ResponseHeaders
	response.StatusCode_ = req.ResponseStatusCode
	req.NbGetCall += 1
	return &response
}

func (req *HttpRequestTest) Head() utils.IHttpResponse {
	var response utils.HttpResponse
	response.Headers_ = req.ResponseHeaders
	response.StatusCode_ = req.ResponseStatusCode
	return &response
}

func (req *HttpRequestTest) Error() error {
	return req.Request.Error_
}

func (req *HttpRequestTest) Url() string {
	return req.Request.Url_
}

const Test_Url string = "https://httpbin.org/get?param1=val1&param2=val2"
const Bad_Url string = "https://not_a_valid_url.null/no_way_path/page.html?exists=false"

func get_Expected_JSON() string {
	return `{
		"args": {
		  "param1": "val1", 
		  "param2": "val2"
		}, 
		"headers": {
		  "Accept-Encoding": "gzip", 
		  "Host": "httpbin.org", 
		  "User-Agent": "Go-http-client/1.1"
		}, 
		"origin": "78.241.157.28, 78.241.157.28", 
		"url": "https://httpbin.org/get?param1=val1&param2=val2"
	  }`
}

type JData struct {
	Args    map[string]string
	Headers map[string]string
	Origin  string
	Url     string
}

func do_TestExpectedJSON(t *testing.T, json_content []byte) {
	var want JData
	var got JData

	decoder := json.NewDecoder(strings.NewReader(get_Expected_JSON()))
	decoder.Decode(&want)

	if !json.Valid(json_content) {
		t.Fatalf("JSON response content not valid")
	}
	decoder = json.NewDecoder(bytes.NewReader(json_content))
	decoder.Decode(&got)

	/*
		JData{
			Args:map[string]string{"param2":"val2", "param1":"val1"},
			Headers:map[string]string{"Accept-Encoding":"gzip", "Host":"httpbin.org", "User-Agent":"Go-http-client/1.1"},
			Origin:"78.241.157.28, 78.241.157.28",
			Url:"https://httpbin.org/get?param1=val1&param2=val2"}
	*/

	if len(want.Args) != len(want.Args) {
		t.Fatalf("Args: not same number (Want %d, Got %d)\n", len(want.Args), len(got.Args))
	}
	for k, v := range want.Args {
		a, ok := got.Args[k]
		if !ok {
			t.Fatalf("Args[%s] not found.", k)
		}
		if a != v {
			t.Fatalf("Args[%s] have different values ('%s'!='%s')", k, v, a)
		}
	}
	if len(want.Headers) != len(want.Headers) {
		t.Fatalf("Headers: not same number (Want %d, Got %d)\n", len(want.Headers), len(got.Headers))
	}
	for k, v := range want.Headers {
		a, ok := got.Headers[k]
		if !ok {
			t.Fatalf("Headers[%s] not found.", k)
		}
		if a != v {
			t.Fatalf("Headers[%s] have different values ('%s'!='%s')", k, v, a)
		}
	}
	if want.Origin != got.Origin {
		t.Fatalf("Origin have different values ('%s'!='%s')\n", want.Origin, got.Origin)
	}
	if want.Url != got.Url {
		t.Fatalf("Url have different values ('%s'!='%s')\n", want.Origin, got.Origin)
	}

}

func do_Test_Headers(t *testing.T, expected_headers map[string][]string, response_headers map[string][]string) {

	for k, v := range expected_headers {

		hv, ok := response_headers[k]

		if !ok {
			t.Fatalf("Headers[%s] not found\n", k)
		}
		if len(hv) != len(v) {
			t.Fatalf("Headers have different length (expected '%d'!='%d')\n", len(v), len(hv))
		}
		for i := 0; i < len(v); i += 1 {
			var found bool = false
			for j := 0; !found && j < len(hv); j += 1 {
				if v[i] == hv[i] {
					found = true
				}
			}
			if !found {
				t.Fatalf("Header[%s] value '%s' not found\n", k, v[i])
			}
		}
	}
}

//-------------------------------------------------------------------------------------
//										TESTS
//-------------------------------------------------------------------------------------
func TestRequestUrl(t *testing.T) {
	var want string = "http://url.test"
	var req utils.IHttpRequest = utils.CreateRequest(want)

	got := req.Url()
	if got != want {
		t.Fatalf("Url should be '%s' not '%s'\n", want, got)
	}

}

func TestGetRequest(t *testing.T) {

	var req utils.IHttpRequest = utils.CreateRequest(Test_Url)
	var response utils.IHttpResponse = req.Get()

	if response == nil {
		t.Fatal("response should not be null")
	}

	do_TestExpectedJSON(t, response.Content())
	/*
		http.Header{
			"Access-Control-Allow-Credentials":[]string{"true"},
			"Access-Control-Allow-Origin":[]string{"*"},
			"Content-Type":[]string{"application/json"},
			"Date":[]string{"Fri, 12 Apr 2019 06:37:09 GMT"},
			"Server":[]string{"nginx"},
			"Connection":[]string{"keep-alive"}}
	*/
	expected_headers := make(map[string][]string)
	expected_headers["Access-Control-Allow-Credentials"] = []string{"true"}
	expected_headers["Access-Control-Allow-Origin"] = []string{"*"}
	expected_headers["Content-Type"] = []string{"application/json"}
	expected_headers["Server"] = []string{"nginx"}
	expected_headers["Connection"] = []string{"keep-alive"}

	do_Test_Headers(t, expected_headers, response.Headers())

}

func TestHeadRequest(t *testing.T) {
	/*
		Head() resp: &utils.HttpResponse{
			Content_:[]uint8(nil),
			Headers_:http.Header{
				"Content-Length":[]string{"258"},
				"Content-Type":[]string{"application/json"},
				"Date":[]string{"Fri, 12 Apr 2019 06:37:09 GMT"},
				"Server":[]string{"nginx"}, "Connection":[]string{"keep-alive"},
				"Access-Control-Allow-Credentials":[]string{"true"},
				"Access-Control-Allow-Origin":[]string{"*"}},
				StatusCode_:200}
	*/
	var req utils.IHttpRequest = utils.CreateRequest(Test_Url)
	var response utils.IHttpResponse = req.Head()

	expected_headers := make(map[string][]string)
	expected_headers["Content-Length"] = []string{"258"}
	expected_headers["Access-Control-Allow-Origin"] = []string{"*"}
	expected_headers["Content-Type"] = []string{"application/json"}
	expected_headers["Server"] = []string{"nginx"}
	expected_headers["Connection"] = []string{"keep-alive"}

	do_Test_Headers(t, expected_headers, response.Headers())

	if response.StatusCode() != 200 {
		t.Fatalf("Unexpected status code %d\n", response.StatusCode())
	}
	if response.Content() != nil {
		t.Fatalf("Content should be nil (got %d)\n", response.Content())
	}
}

func TestRequestFailure(t *testing.T) {
	req := utils.CreateRequest(Bad_Url)
	response := req.Get()

	if response != nil {
		t.Fatalf("response should be null (got %#v)\n", response)
	}

	if req.Error() == nil {
		t.Fatal("Error should not be null\n")
	}
}

func TestGetUrlContent(t *testing.T) {

	req := utils.CreateRequest(Test_Url)

	bt, err := utils.GetUrlContent(req)

	if err != nil {
		t.Fatalf("Error should be null (got %s)\n", err.Error())
	}

	if bt == nil {
		t.Fatal("[]byte should not be null\n")
	}

	do_TestExpectedJSON(t, bt)

	req = utils.CreateRequest(Bad_Url)

	bt, err = utils.GetUrlContent(req)

	if err == nil {
		t.Fatal("Error should not be null\n")
	}

	if bt != nil {
		t.Fatalf("[]byte should be null (got %s)\n", string(bt))
	}
}

func TestGetUrlHeaders(t *testing.T) {

	req := utils.CreateRequest(Test_Url)

	headers, err := utils.GetUrlHeaders(req)

	if err != nil {
		t.Fatalf("Error should be null (got %s)\n", err.Error())
	}

	if headers == nil {
		t.Fatal("headers should not be null\n")
	}

	expected_headers := make(map[string][]string)
	expected_headers["Content-Length"] = []string{"258"}
	expected_headers["Access-Control-Allow-Origin"] = []string{"*"}
	expected_headers["Content-Type"] = []string{"application/json"}
	expected_headers["Server"] = []string{"nginx"}
	expected_headers["Connection"] = []string{"keep-alive"}

	do_Test_Headers(t, expected_headers, headers)

	req = utils.CreateRequest(Bad_Url)

	headers, err = utils.GetUrlHeaders(req)

	if err == nil {
		t.Fatal("Error should not be null\n")
	}

	if headers != nil {
		t.Fatalf("[]byte should be null (got %#v)\n", headers)
	}
}

func TestGetImageUrl(t *testing.T) {
	html_src := []byte(`<html><body>
	<div class="main">
    <div class="banner"><h1 class="title">BOOK STORE</h1></div>

        <div class='internal'>
        <div class="nav_elements" style="border-radius: 15px">
        <div class='book_record'><table class='book_record'><tr><td width='210px' valign='top' rowspan='2'><img src="img/%5Bmicrosoft-excel-2016-programming-example%5D/img.png" class='book'><div class='book'>Tags:</div><div class='tags display'><span class="tag label label-info"><span style="cursor: pointer" onclick="window.location='home.php?subject=186';">PROGRAMMING</span></span><span class="tag label label-info"><span style="cursor: pointer" onclick="window.location='home.php?subject=134';">XL</span></span></div></td><td><div class='book'><span class='book_title'>Microsoft Excel 2016 Programming by Example: with VBA, XML, and ASP</span></div></td></tr><tr><td><div class='book'>Auteur: <span>Julitta Korol</span></div><div class='book'>Parution: <span>2017</span></div><div class='book'>File size: <span>57.9 Mb</span></div><div class='book'><span class='book_descr'><div class='scrollable'><h3>EPUB Format</h3><hr><br/><p>Updated for Excel 2016 and based on the bestselling editions from previous versions, Microsoft Excel 2016 Programming by Example with VBA, XML and ASP is a practical, how-to book on Excel programming, suitable for readers already proficient with the Excel user interface (UI). If you are looking to automate Excel routine tasks, this book will progressively introduce you to programming concepts via numerous, illustrated, hands-on exercises. Includes a comprehensive disc with source code, supplemental files, and color screen captures (Also available from the publisher for download by writing to [email protected]). More advanced topics are demonstrated via custom projects. From recording and editing a macro and writing VBA code to working with XML documents and using Classic ASP pages to access and display data on the Web, this book takes you on a programming journey that will change the way you work with Excel. The book provides information on performing automatic operations on files, folders, and other Microsoft Office applications. It also covers proper use of event procedures, testing and debugging, and guides you through programming advanced Excel features such as PivotTables, PivotCharts, and the Ribbon interface. </p><br/></div></span></div><div class='book'><button id='btnDownload' class='nav_element' onclick='downloadFile(6363);'>Download</button> 60739945 octets (GOOGLE)</div></td></tr><tr><td>&nbsp;</td><td colspan='2'><div class='upload-out' id='upload-out' style='visibility: hidden; margin-left: 1cm'><div class='upload-in' id='upl-in1'><div></div></td></tr></table></div><a class="nav_element" href="/books/home.php?page=4&search=image&order=DESC">back</a><br>        </div>
    </div>
  <span style="text-align: right; color: white"><h5 style="padding-right: 15px"><i>&#169; Stéphane Samara</i></h5></span>
  </div></body></html>`)

	got, err := utils.GetImageUrl(html_src)

	if err != nil {
		t.Fatalf("Error should be null (got %s) \n", err)
	}
	if len(got) == 0 {
		t.Fatal("Image url length should not be zero")
	}
	want := "img/%5Bmicrosoft-excel-2016-programming-example%5D/img.png"
	if got != want {
		t.Fatalf("Url image mismatch (want'%s', got'%s'\n", want, got)
	}
}
