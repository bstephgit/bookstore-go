package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"

	"github.com/bookstore-go/utils"
)

type StorageClient interface {
	Auth(url string)
	OnCode(code int)
	OnToken(token string)
}

type GoogleClient struct {
	Connected bool
}

type myHandler struct {
	Ch chan<- string
}

func (handler *myHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {

	if req.URL.Path == "/" {
		fmt.Printf("Got request %v\n", req.URL)

		f, _ := os.Create("resp.html")
		io.Copy(rw, f)
		f.Close()

	}

	handler.Ch <- req.URL.RawQuery
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

func (goog *GoogleClient) Auth() {

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
	s := <-ch

	fmt.Printf("response received %s\n", s)
}
