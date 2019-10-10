package openapi

import (
	"encoding/json"
	"github.com/mitchellh/go-homedir"
	"io/ioutil"
	"log"
	"net/url"
)

func prettyPrint(v interface{}) {
	b, _ := json.MarshalIndent(v, "", "  ")
	log.Printf(string(b))
	log.Println()
}

func sPrettyPrint(v interface{}) string {
	if v == nil {
		return "nil"
	}
	b, _ := json.MarshalIndent(v, "", "  ")
	return string(b)
}

func expandPath(filePath string) (string, error) {
	fullPath, err := homedir.Expand(filePath)
	if err != nil {
		return "", err
	}
	return fullPath, err
}

func getFileContent(filePath string) (string, error) {
	fullPath, err := expandPath(filePath)
	if err != nil {
		return "", err
	}
	data, err := ioutil.ReadFile(fullPath)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func isUrl(str string) bool {
	u, err := url.Parse(str)
	return err == nil && u.Scheme != "" && u.Host != ""
}
