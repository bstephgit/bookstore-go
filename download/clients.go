package download

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/bookstore-go/utils"
)

type StorageClient interface {
	Auth(url string) string
	//OnCode(code int)
	//OnToken(token string)
	DownloadFile(fileId, file_name, token string)
}

var Clinets map[string]StorageClient

type GoogleClient struct {
	Connected bool
	Token     string
}

type myHandler struct {
	Ch chan<- string
}

func DownloadFile(book *utils.BookDownload) {

}

func (handler *myHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {

	if req.URL.Path == "/" {
		fmt.Printf("Got request %v\n", req.URL)

		f, _ := os.Open("resp.html")
		io.Copy(rw, f)
		f.Close()

	}
	if req.URL.Path == "/token" {
		splitted := strings.Split(req.URL.RawQuery, "&")
		const access_token = "access_token"
		var token string

		for _, s := range splitted {
			if s[:len(access_token)] == access_token {
				token = strings.Split(s, "=")[1]
			}
		}

		handler.Ch <- token

	}
}

func StartServer(ch chan<- string) {
	handler := &myHandler{ch}
	mux := http.NewServeMux()
	mux.Handle("/", handler)
	mux.Handle("/token", handler)
	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		log.Fatal(err)
	}
}

func (goog *GoogleClient) Auth() string {

	const BaseUrl = "https://accounts.google.com/o/oauth2/v2/auth"
	const ClientId = "76824108658-qopibc57hedf4k4he7rlateis2bkoigv.apps.googleusercontent.com"
	const ClientSecret = "444KL-z0e_JZR8_PBu_vXvKH"

	args := map[string]string{"client_id": ClientId, "redirect_uri": "http://localhost:8080", "response_type": "token",
		"scope": "https://www.googleapis.com/auth/drive.readonly", "login_hint": "tcn75323@gmail.com"}

	url := utils.AddUrlArguments(BaseUrl, args)

	req := utils.CreateRequest(url)

	_, err := utils.GetUrlContent(req)
	if err != nil {
		fmt.Printf("%v\n", err)
	}

	//fmt.Println(string(content))
	cmd := exec.Command("xdg-open", url)
	cmd.Start()

	ch := make(chan string)
	go StartServer(ch)
	token := <-ch

	fmt.Printf("token received %s\n", token)

	return token
}

func (goog *GoogleClient) DownloadFile(fileId, file_name, token string) {

	const URL = "https://www.googleapis.com/drive/v3/files/"
	const SCOPE = "https://www.googleapis.com/auth/drive.readonly"

	fileurl := URL + fileId

	resp, err := http.Get(fileurl)

	if err != nil {
		fmt.Printf("%v\n", err)
	}

	out, _ := os.Create(file_name)

	_, err = io.Copy(out, resp.Body)
	out.Close()
}
