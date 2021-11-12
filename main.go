package main

import (
	"log"
	"os"

	"github.com/bookstore-go/config"
	"github.com/bookstore-go/console"
	"github.com/bookstore-go/download"
	"github.com/bookstore-go/utils"
)

func main() {

	os.Setenv("LANG", "$LANG.UTF-8")

	err := config.LoadConfig()

	if err != nil {
		panic(err)
	}

	err = utils.DbConnect(config.GetConfig().Database)

	if err != nil {
		log.Fatal("Cannot connect to database " + err.Error())
	}

	// connect to db to get file storage vendors
	download.InitVendorsData()

	console.TerminalLoop()

	err = utils.DbClose()

	if err != nil {
		log.Fatal("Error closing database " + err.Error())
	}

}
