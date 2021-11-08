package main

import (
	"fmt"
	"log"
	"os"

	"github.com/BurntSushi/toml"
	"github.com/bookstore-go/utils"
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

type config struct {
	Url      string
	Workdir  string
	Database utils.Database
}

// USE PTERM TO DISPLAY DATA?
// https://github.com/pterm/pterm

var Base_url_folder string = "https://vps665513.ovh.net/books/"
var Base_url string = Base_url_folder + "home.php"

func main() {

	/**if len(os.Args) < 2 {
		fmt.Printf("missing arguments: provide at least 1 number\n")
		os.Exit(1)
	}**/
	//var url_img string

	var conf config
	md, err := toml.DecodeFile("config.toml", &conf)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Undecoded keys: %q\n", md.Undecoded())

	err = utils.DbConnect(conf.Database)

	os.Setenv("LANG", "$LANG.UTF-8")
	if err != nil {
		log.Fatal("Cannot connect to database " + err.Error())
		panic(err)
	}

	var goog GoogleClient
	goog.Auth()
	//console.TerminalLoop()

	err = utils.DbClose()

	if err != nil {
		log.Fatal("Error closing database " + err.Error())
	}

	/*for index := 1; index < len(os.Args); index += 1 {
		var err error

		_, err = strconv.Atoi(os.Args[index])

		if err == nil {

			//url_img = Base_url + "?bookid=" + os.Args[index]
			//fmt.Printf("Page url => %s\n", url_img)
			//download.DownloadImage(url_img)

		} else {
			fmt.Printf("Argument bad format:%s. Not a number", os.Args[index])
		}
	}*/

}
