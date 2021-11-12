package config

import (
	"flag"
	"fmt"

	"github.com/BurntSushi/toml"
	"github.com/bookstore-go/utils"
)

type DownloadImg struct {
	bookId int32
	res    error
}

type FileStore struct {
	ClientId    string
	Secret      string
	RedirectUri string
}

type Config struct {
	Database utils.Database
	Dirs     struct {
		Workdir  string
		Download string
	}
	Storage struct {
		Google   FileStore
		OneDrive FileStore
		Box      FileStore
	}
}

var GetConfig func() Config

func LoadConfig() error {
	var conf Config
	md, err := toml.DecodeFile("config.toml", &conf)

	if err != nil {
		return err
	}

	fmt.Printf("Undecoded keys: %q\n", md.Undecoded())

	GetConfig = func() Config {
		return conf
	}
	return nil
}

func ProcessArgs() error {

	var decrypt_file string
	var encrypt_file string
	var conf_file string

	flag.StringVar(&decrypt_file, "d", "decrypt", "File containing key to decrypt configuration")
	flag.StringVar(&encrypt_file, "e", "encrypt", "File containing key to encrypt configuration")
	flag.StringVar(&conf_file, "c", "config", "Configuration file. Default is conf.tml")

	flag.Parse()

	return nil
}
