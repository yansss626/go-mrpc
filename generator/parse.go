package generator

import (
	"os"

	"github.com/goccy/go-yaml"
)

func Parse(filename string) (*IDL, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var idl IDL
	err = yaml.Unmarshal(data, &idl)
	if err != nil {
		return nil, err
	}
	return &idl, nil
}
