package main

import (
	"bookstore/download"
	"fmt"
	"os"
	"strconv"
)

type Action interface {
	Do()
	Result() error
	SetError(err error)
	Data() []byte
}

type DownloadImg struct {
	bookId int32
	res    error
}

var Base_url_folder string = "https://vps665513.ovh.net/books/"
var Base_url string = Base_url_folder + "home.php"

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("missing arguments: provide at least 1 number\n")
		os.Exit(1)
	}
	var url_img string

	for index := 1; index < len(os.Args); index += 1 {
		var err error

		_, err = strconv.Atoi(os.Args[index])

		if err == nil {

			url_img = Base_url + "?bookid=" + os.Args[index]
			fmt.Printf("Page url => %s\n", url_img)
			download.DownloadImage(url_img)

		} else {
			fmt.Printf("Argument bad format:%s. Not a number", os.Args[index])
		}
	}

}
