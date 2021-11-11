package download

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/bookstore-go/utils"
)

type StorageClient interface {
	AuthUrl() string
	DownloadUrl(fileId string) string
	RefreshToken() string
	OnToken(token string)
	RedirectUri() string
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
type OneDriveClient struct{}
type BoxComClient struct{}

type myHandler struct {
	Ch      chan<- string
	ChNotif chan<- bool
	Server  *http.Server
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

	go StartServer(ch, client.RedirectUri())
	token, ok := <-ch

	if !ok {
		return errors.New("Error retrieving token.")
	} else {
		client.OnToken(token)
	}

	return nil
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

	//log.Println("url ", fileurl)
	//log.Println("Set token", data.Token)

	req.Header.Add("Authorization", "Bearer "+data.Token)

	resp, err := client.Do(req)

	log.Printf("%v\n%v\n%v\n", req, resp, resp.Body)

	if err != nil {
		log.Printf("%v\n", err)
		return err
	}

	if resp.StatusCode >= 200 && resp.StatusCode < 400 {

		out, _ := os.Create(book.FileName)
		_, err = io.Copy(out, resp.Body)
		out.Close()

		log.Println("File downloaded", resp.StatusCode, resp.Status)
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
		if v.VendorCode == "MSOD" {
			data.Client = &OneDriveClient{}
		}
		_StorageData[v.VendorCode] = data
	}
}

func OnToken(token, code string) error {
	data, _ := _StorageData[code]

	splitted := strings.Split(token, "&")
	const access_token = "access_token"
	const expiry = "expires_in"
	const error_token = "error"
	const error_desc = "error_description"

	var error_str string
	var error_desc_str string

	for _, s := range splitted {
		if len(s) > len(access_token) && s[:len(access_token)] == access_token {
			data.Token = strings.Split(s, "=")[1]
		}
		if len(s) > len(expiry) && s[:len(expiry)] == expiry {
			tokens := strings.Split(s, "=")
			if len(tokens) > 1 {
				var err error
				data.TokenValidity, err = strconv.Atoi(tokens[1])
				if err != nil {
					data.TokenValidity = 0
				}
			}
		}
		if len(s) > len(error_token) && s[:len(error_token)] == error_token {
			error_str = strings.Split(s, "=")[1]
		}
		if len(s) > len(error_desc) && s[:len(error_desc)] == error_desc {
			error_desc_str = strings.Split(s, "=")[1]
		}

	}

	if len(error_str) > 0 {
		if len(error_desc_str) > 0 {
			err_msg := fmt.Sprintf("%s: %s\n", error_str, error_desc_str)
			return errors.New(err_msg)
		}
		return errors.New(error_str)
	}
	return nil
}

//
/// HTTP SERVER for API REDIRECTION
//
func (handler *myHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	//log.Printf("Got request %v\n", req.URL)

	if req.URL.Path == "/" {

		f, err := os.Open("download/resp.html")
		if err == nil {
			io.Copy(rw, f)
			f.Close()
		} else {
			io.WriteString(rw, "resp.html error")
			close(handler.Ch)
		}

	}
	if req.URL.Path == "/token" || req.URL.Path == "/error" {
		handler.Ch <- req.URL.RawQuery
		io.WriteString(rw, "Ok")
		handler.ChNotif <- true
	}

}

func StartServer(ch chan string, url string) {
	handler := &myHandler{ch, nil, nil}
	server := &http.Server{
		Addr:    ":8080",
		Handler: handler,
	}
	handler.Server = server
	mux := http.NewServeMux()

	mux.Handle("/", handler)
	mux.Handle("/token", handler)
	mux.Handle("/error", handler)

	chnotif := make(chan bool)
	handler.ChNotif = chnotif

	go func() {
		timer := time.NewTimer(time.Second * 30)
		select {
		case <-timer.C:
			fmt.Println("\nServer timeout")
			timer.Stop()
			close(ch) // unlock channel for receiver
		case <-chnotif:
			timer.Stop()
		}
		defer server.Close()
	}()

	var err error
	if url[:5] == "https" {
		err = server.ListenAndServeTLS("download/localhost.cert", "download/localhost.key")
	} else {
		err = server.ListenAndServe()
	}
	if err != nil {
		fmt.Printf("%v\n", err)
	}
}

//
/// Storage API methods
//

// ----------- Google Drive  -----------

func (goog *GoogleClient) RedirectUri() string {
	return "http://localhost:8080"
}

func (goog *GoogleClient) AuthUrl() string {
	const BaseUrl = "https://accounts.google.com/o/oauth2/v2/auth"
	const ClientId = "76824108658-qopibc57hedf4k4he7rlateis2bkoigv.apps.googleusercontent.com"
	const ClientSecret = "444KL-z0e_JZR8_PBu_vXvKH"

	redirect_uri := goog.RedirectUri()

	args := map[string]string{"client_id": ClientId, "redirect_uri": redirect_uri, "response_type": "token",
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
	OnToken(token, code)
}

// -----------  OneDrive -----------

func (od *OneDriveClient) AuthUrl() string {
	//"https://login.microsoftonline.com/common/oauth2/v2.0/authorize?client_id={client_id}&scope={scope}&response_type=token&redirect_uri={redirect_uri}"
	const BASE_URL = "https://login.microsoftonline.com/common/oauth2/v2.0/authorize?client_id=%s&scope=%s&response_type=token&redirect_uri=%s"

	const CLIENT_ID = "ffd664c2-aba8-4284-ba2e-2300b5320387"
	const SCOPE = "Files.Read"
	//"https://myapp.com/auth-redirect#access_token=EwC...EB&authentication_token=eyJ...3EM&token_type=bearer&expires_in=3600&scope=onedrive.readwrite&user_id=3626...1d"
	url := fmt.Sprintf(BASE_URL, CLIENT_ID, SCOPE, url.QueryEscape(od.RedirectUri()))
	return url
}

func (od *OneDriveClient) DownloadUrl(fileId string) string {

	const URL = "https://graph.microsoft.com/v1.0/me/drive/items/%s/content"
	const SCOPE = ""

	splitted := strings.Split(fileId, ".")
	if len(splitted) == 3 {
		fileId = splitted[2]
	}
	return fmt.Sprintf(URL, fileId)
}

func (od *OneDriveClient) RedirectUri() string {
	const REDIRECT = "https://localhost:8080"
	return REDIRECT
}

func (msod *OneDriveClient) OnToken(token string) {
	const code = "MSOD"

	//access_token=ya29.a0ARrdaM-sVhN8knzKB9QXXOgn_Z_TIdbffDWnTapzSH0_zDI7SL-CQjza_tg15MhzScp8HUFcOVF-YbSRm5BiOThl57RsmihjpZcJHj7ERpXdSNXXKy-9-uwRxBpyA0rCDg7-7kDXu4NvouG0W2tob9xZzLoW
	//&token_type=Bearer&expires_in=3599&scope=File.Read
	err := OnToken(token, code)
	if err != nil {
		log.Printf("%v\n", err)
	}
}

func (msod *OneDriveClient) RefreshToken() string {
	return "N/A"
}

// ----------- BOX.COM -----------

func (bx *BoxComClient) AuthUrl() string {
	const CLIENT_ID = "2en9g8pt7jgu5kgvyss7qbrxgk783212"
	const CLIENT_SECRET = "t0nY1UF8AkmKZZp7qPEHWU8i2OG2pZwD"
	//curl -i -X GET "https://account.box.com/api/oauth2/authorize?response_type=code&client_id=ly1nj6n11vionaie65emwzk575hnnmrk&redirect_uri=http://example.com/auth/callback"
	const BASE_URL = "https://account.box.com/api/oauth2/authorize?response_type=code&client_id=ly1nj6n11vionaie65emwzk575hnnmrk&redirect_uri=http://example.com/auth/callback"

	return BASE_URL
}
