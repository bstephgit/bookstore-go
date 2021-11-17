package config

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/bookstore-go/utils"
	"golang.org/x/crypto/scrypt"
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
	Ssl struct {
		Key    string
		PubKey string `toml:"pub-key"`
	}
	Storage struct {
		Google   FileStore
		OneDrive FileStore
		Box      FileStore
	}
	ConfigFile string
}

var GetConfig func() Config

func LoadConfig() error {

	var conf Config
	err := ProcessArgs(&conf)

	if err != nil {
		return err
	}

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

func ProcessArgs(conf *Config) error {

	//var encrypt_file bool
	var password string

	encrypt := flag.Bool("e", false, "Do encrypt configuration file (output = <config-file> encrypt).")
	flag.StringVar(&conf.ConfigFile, "c", "config.toml", "Configuration file. Default is conf.toml")
	flag.StringVar(&password, "p", "", "Password to decrypt/encrypt configuration file")

	flag.Parse()

	if len(conf.ConfigFile) == 0 {
		_, err := os.Stat(conf.ConfigFile)
		if err != nil {
			return fmt.Errorf("Config file %s not found", conf.ConfigFile)
		}
	}

	if *encrypt {
		file_encr := conf.ConfigFile + ".encrypted"
		log.Printf("Encrypt file to %s: password '%s'\n", file_encr, password)
		if len(password) > 0 {
			err := EncryptFile(password, conf.ConfigFile, file_encr)
			if err != nil {
				return err
			}
		} else {
			return errors.New("Cannot encrypt configuration file: no password provided.")
		}
	}

	md, err := toml.DecodeFile(conf.ConfigFile, &conf)
	fmt.Printf("Undecoded keys: %q\n", md.Undecoded())

	if err != nil { // if error try decrypt file
		log.Println("Configuration is not valid Toml file. Try to decrypt file...")
		if len(password) > 0 {

			var sb strings.Builder
			err = DecryptFile(password, conf.ConfigFile, &sb)
			if err != nil {
				return err
			}
			md, err = toml.Decode(sb.String(), &conf)
			if err != nil {
				return err
			}

		} else {
			return fmt.Errorf("Cannot decrpyt configuration file: no password provided.")
		}
	}

	return nil
}

func EncryptFile(password, inputfile, outputfile string) error {

	log.Println("Encrypt file...")
	// encrypt file
	passdata := []byte(password)

	log.Println("Generate key from password...")

	c := 8
	salt := make([]byte, c)
	_, err := rand.Read(salt)
	if err != nil {
		return err
	}

	log.Printf("Generating key with salt=%s...\n", hex.EncodeToString(salt))
	key, err := scrypt.Key(passdata, salt, 1<<15, 8, 1, 32)

	if err != nil {
		return err
	}

	log.Printf("Create new cipher block key=%s...\n", hex.EncodeToString(key))

	block, err := aes.NewCipher(key)

	if err != nil {
		return err
	}

	log.Println("Create new Galois cipher block...")

	aesgcm, err := cipher.NewGCM(block)

	log.Println("Read input file...")

	content, err := ioutil.ReadFile(inputfile)

	if err != nil {
		return err
	}

	nonce, _ := hex.DecodeString("64a9433eae7ccceee2fc0eda")

	//var dst []byte
	log.Println("Encrypting...")

	dst := aesgcm.Seal(nil, nonce, content, nil)

	log.Println("Create output file...")
	ioutil.WriteFile(outputfile, []byte(hex.EncodeToString(salt)+"$"+hex.EncodeToString(dst)), 0700)

	return err
}

func DecryptFile(password, inputfile string, sb *strings.Builder) error {

	log.Println("Decrypt file...")
	// decrypt file
	passdata := []byte(password)

	log.Println("Generate key from password...")

	content, err := ioutil.ReadFile(inputfile)

	var salt []byte

	if content[8*2] == '$' {

		salt, err = hex.DecodeString(string(content[:8*2]))
		if err != nil {
			return nil
		}

		content = content[8*2+1:] // skip  'salt' + '$' char

	} else {
		return errors.New("Salt invalid format: should be delimitted by $")
	}

	log.Printf("Generating key with salt=%s...\n", hex.EncodeToString(salt))

	key, err := scrypt.Key(passdata, salt, 1<<15, 8, 1, 32)

	if err != nil {
		return err
	}

	log.Printf("Create new cipher block key=%s...\n", hex.EncodeToString(key))

	block, err := aes.NewCipher(key)

	if err != nil {
		return err
	}

	log.Println("Create new Galois cipher block...")

	aesgcm, err := cipher.NewGCM(block)

	log.Println("Read input file...")

	if err != nil {
		return err
	}

	content, err = hex.DecodeString(string(content))

	if err != nil {
		return err
	}

	nonce, _ := hex.DecodeString("64a9433eae7ccceee2fc0eda")

	log.Println("Decrypting...")

	dst, err := aesgcm.Open(nil, nonce, content, nil)

	if err != nil {
		return err
	}

	//log.Println("Create output file...")
	sb.Write(dst)

	return err
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
