package main

import (
	"fmt"
	"log"
	"os"

	"github.com/BurntSushi/toml"
	"github.com/bookstore-go/console"
	"github.com/bookstore-go/download"
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

	os.Setenv("LANG", "$LANG.UTF-8")

	var conf config
	md, err := toml.DecodeFile("config.toml", &conf)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Undecoded keys: %q\n", md.Undecoded())

	err = utils.DbConnect(conf.Database)

	if err != nil {
		log.Fatal("Cannot connect to database " + err.Error())
		panic(err)
	}

	// connect to db to get file storage vendors
	download.InitVendorsData()

	console.TerminalLoop()

	err = utils.DbClose()

	if err != nil {
		log.Fatal("Error closing database " + err.Error())
	}

}
