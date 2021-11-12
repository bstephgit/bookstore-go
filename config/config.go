package config

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"strings"

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

	conf.Dirs.Workdir = ExpandPath(conf.Dirs.Workdir)
	err = CreateFolderIfNotExist(conf.Dirs.Workdir)
	if err != nil {
		log.Fatal(err)
	}

	conf.Dirs.Download = ExpandPath(conf.Dirs.Download)
	err = CreateFolderIfNotExist(conf.Dirs.Download)
	if err != nil {
		log.Fatal(err)
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

func ExpandPath(dirpath string) string {

	if dirpath[0:2] == "~/" {
		return path.Join(os.Getenv("HOME"), dirpath[2:])

	} else if dirpath[0] != '/' {
		return path.Join(os.Getenv("PWD"), dirpath)
	}

	return dirpath
}

func CreateFolderIfNotExist(path string) error {

	_, err := os.Stat(path)

	if os.IsNotExist(err) {
		folders := strings.Split(path, "/")
		built_path := ""

		for _, f := range folders {
			built_path += f + "/"
			_, err = os.Stat(built_path)
			if os.IsNotExist(err) {
				err = os.Mkdir(built_path, 0700)
				if err != nil {
					return err
				}
			}
		}
	} else {
		err = nil
	}

	return err
}
