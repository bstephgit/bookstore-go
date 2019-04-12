package download

import (
	"bookstore/utils"
	"fmt"
	"net/url"
	"path"
)

func SaveImage(url string, content []byte) {

	var folder string
	var err error
	pattern := fmt.Sprintf("img/(.+)/.+")

	folder, err = utils.Extract_string(pattern, url)

	if err != nil {
		fmt.Printf("No directory found for url: %s\n", url)
		return
	}

	folder, err = utils.UnescapeUrlString(folder)
	if err != nil {
		fmt.Printf("error unescaping url %s", folder)
		return
	}

	err = utils.Create_folder(folder)
	if err != nil {
		fmt.Printf("create folder failed %s\n", folder)
	}

	file_path := path.Join(folder, "img.png")
	fmt.Printf("image file= %s\n", file_path)

	n, err := utils.Save_file(file_path, content)
	if err == nil {
		fmt.Printf("File %s OK: %d bytes written\n", file_path, n)
	} else {
		fmt.Printf("%s\n", err)
	}
}

func DownloadImage(url_page string) {
	var err error

	req := utils.CreateRequest(url_page)
	resp := req.Get()
	if resp == nil {
		fmt.Printf("Content not found at %s\n%s\n", url_page, req.Error())
		return
	}

	url_img, err := utils.GetImageUrl(resp.Content())

	if err != nil {
		fmt.Printf("Image url cannot be extracted from page. Request page got status %d\n", resp.StatusCode())
		return
	}
	oUrl, err := url.Parse(url_page)
	if err == nil {
		oUrlImg, err := url.Parse(url_img)

		if err == nil {

			oUrl.Path = "/books/" + oUrlImg.Path
			url_img = oUrl.String()

		} else {
			fmt.Println("Image url cannot be parsed " + url_img)
			return
		}
	} else {
		fmt.Printf("Cannot parse url imaget: %s -- err: %s\n", url_img, err)
		return
	}
	req = utils.CreateRequest(url_img)
	resp = req.Get()
	if resp == nil {
		fmt.Printf("Content not found at: %s\n%s\n", url_img, req.Error())
		return
	}
	SaveImage(url_img, resp.Content())
}
