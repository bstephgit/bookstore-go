package download

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/bookstore-go/utils"
)

type StorageClient interface {
	AuthUrl() string
	DownloadUrl(fileId string) string
	RefreshToken() string
	OnToken(token string)
	//OnCode(code int)
	//DownloadFile(fileId, file_name, token string)

}

type StorageData struct {
	VendorId      int
	VendorCode    string
	Connected     bool
	Token         string
	RefreshToken  string
	TokenValidity int
	Client        StorageClient
}

var _StorageData map[string]*StorageData

type GoogleClient struct{}

type myHandler struct {
	Ch chan<- string
}

func DownloadFile(book *utils.BookDownload) error {

	data, ok := _StorageData[book.VendorCode]

	if !ok {
		return fmt.Errorf("Vendor %s not found", book.VendorCode)
	}

	if data.Client == nil {
		return errors.New("Client not implemented")
	}

	if len(data.Token) == 0 {
		err := Auth(data.Client)
		if err != nil {
			return err
		}
	}

	fileurl := data.Client.DownloadUrl(book.FileId)

	client := &http.Client{}
	req, err := http.NewRequest("GET", fileurl, nil)
	req.Header.Add("Authorization", "Bearer "+data.Token)

	resp, err := client.Do(req)

	if err != nil {
		fmt.Printf("%v\n", err)
		return err
	}

	if resp.StatusCode >= 200 && resp.StatusCode < 400 {

		out, _ := os.Create(book.FileName)
		_, err = io.Copy(out, resp.Body)
		out.Close()

		fmt.Println("File downloaded")
	} else {
		return fmt.Errorf("Error file download Code=%d, status=%s", resp.StatusCode, resp.Status)
	}
	return nil
}

func InitVendorsData() {
	vendors, _ := utils.GetVendors()
	_StorageData = make(map[string]*StorageData)

	for _, v := range vendors {
		data := &StorageData{}
		data.Connected = false
		data.VendorId = v.Id
		data.VendorCode = v.VendorCode
		data.TokenValidity = 0

		if v.VendorCode == "GOOG" {
			data.Client = &GoogleClient{}
		}
		_StorageData[v.VendorCode] = data
	}
}

func (handler *myHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {

	if req.URL.Path == "/" {
		fmt.Printf("Got request %v\n", req.URL)

		f, err := os.Open("download/resp.html")
		if err == nil {
			io.Copy(rw, f)
			f.Close()
		} else {
			io.WriteString(rw, "resp.html error")
			close(handler.Ch)
		}

	}
	if req.URL.Path == "/token" {
		handler.Ch <- req.URL.RawQuery
		io.WriteString(rw, "Ok")
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

func Auth(client StorageClient) error {

	url := client.AuthUrl()

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
	token, ok := <-ch

	if !ok {
		return errors.New("Error retrieving token.")
	} else {
		client.OnToken(token)
	}

	return nil
}

func (goog *GoogleClient) AuthUrl() string {
	const BaseUrl = "https://accounts.google.com/o/oauth2/v2/auth"
	const ClientId = "76824108658-qopibc57hedf4k4he7rlateis2bkoigv.apps.googleusercontent.com"
	const ClientSecret = "444KL-z0e_JZR8_PBu_vXvKH"

	args := map[string]string{"client_id": ClientId, "redirect_uri": "http://localhost:8080", "response_type": "token",
		"scope": "https://www.googleapis.com/auth/drive.readonly", "login_hint": "tcn75323@gmail.com"}

	return utils.AddUrlArguments(BaseUrl, args)
}

func (goog *GoogleClient) DownloadUrl(fileId string) string {

	const URL = "https://www.googleapis.com/drive/v3/files/"
	const SCOPE = "https://www.googleapis.com/auth/drive.readonly"

	return URL + fileId + "?alt=media"
}

func (goog *GoogleClient) RefreshToken() string {
	const code = "GOOG"
	data, _ := _StorageData[code]

	return data.RefreshToken

}

func (goog *GoogleClient) OnToken(token string) {
	const code = "GOOG"

	//access_token=ya29.a0ARrdaM-sVhN8knzKB9QXXOgn_Z_TIdbffDWnTapzSH0_zDI7SL-CQjza_tg15MhzScp8HUFcOVF-YbSRm5BiOThl57RsmihjpZcJHj7ERpXdSNXXKy-9-uwRxBpyA0rCDg7-7kDXu4NvouG0W2tob9xZzLoW
	//&token_type=Bearer&expires_in=3599&scope=https://www.googleapis.com/auth/drive.readonly
	data, _ := _StorageData[code]

	splitted := strings.Split(token, "&")
	const access_token = "access_token"
	const expiry = "expires_in"

	for _, s := range splitted {
		if s[:len(access_token)] == access_token {
			data.Token = strings.Split(s, "=")[1]
		}
		if s[:len(expiry)] == expiry {
			tokens := strings.Split(s, "=")
			if len(tokens) > 1 {
				var err error
				data.TokenValidity, err = strconv.Atoi(tokens[1])
				if err != nil {
					data.TokenValidity = 0
				}
			}
		}
	}
}
