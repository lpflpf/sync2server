package internal

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v2"

	"path/filepath"
)

func LoadConfig(file string, s *[]ProjectSyncConfig) error {
	if !fileExist(file) {
		return errors.New("config " + file + "not exists")
	}

	data, err := ioutil.ReadFile(file)
	if err != nil {
		return errors.New(file + " load err." + err.Error())
	}

	switch filepath.Ext(file) {
	case ".json":
		return json.Unmarshal(data, s)
	case ".yaml":
		return yaml.Unmarshal(data, s)
	}

	return errors.New("Do not support this type file:" + file)
}

func fileExist(path string) bool {
	_, err := os.Lstat(path)
	return !os.IsNotExist(err)
}
