package common

import (
	"api/models"
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"path"
	"runtime"

	yaml "gopkg.in/yaml.v2"
)

func InitConfig() *models.Config {
	var cfg models.Config

	// get this file's path
	_, filename, _, _ := runtime.Caller(0)
	folderpath := path.Dir(filename)

	// retrieve common's test.config
	f, err := os.Open(fmt.Sprintf("%s/test.config.yml", folderpath))
	if err != nil {
		processError(err)
	}
	defer f.Close()

	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(&cfg)
	if err != nil {
		processError(err)
	}

	if cfg.Debug {
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	return &cfg
}

func processError(err error) {
	fmt.Println(err)
	os.Exit(2)
}
