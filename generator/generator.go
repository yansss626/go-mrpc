package generator

import (
	"bytes"
	"fmt"
	"go/format"
	"html/template"
	"os"
	"path/filepath"
)

const (
	GeneratedPath = "./generated/"
)

func Generate(idl *IDL) error {

	path := GeneratedPath + idl.Package

	err := os.MkdirAll(path, 0755)
	if err != nil {
		return err
	}

	err = GenerateMessage(idl, path)
	if err != nil {
		return err
	}

	err = GenerateClient(idl, path)
	if err != nil {
		return err
	}

	err = GenerateServer(idl, path)
	if err != nil {
		return err
	}

	return nil
}

// message

/*
package hello

import "context"

type HelloRequest struct {
	Name string `json:"name"`
}

type HelloResponse struct {
	Message string `json:"message"`
}

*/

func GenerateMessage(idl *IDL, out string) error {
	file, err := os.Create(out + "/" + idl.Package + ".go")
	if err != nil {
		return err
	}
	defer file.Close()

	fmt.Fprintf(file, "package %s\n\n", idl.Package)

	for name, message := range idl.Messages {
		fmt.Fprintf(file, "type %s struct {\n", name)
		for _, field := range message.Fields {
			fieldName := FieldGoName(field.Name)
			fieldType := FieldGoType(field)
			jsonName := FieldJSONName(field)
			fmt.Fprintf(file, "\t%s %s `json:\"%s\"`\n", fieldName, fieldType, jsonName)
		}
		fmt.Fprintf(file, "}\n\n")
	}
	return nil
}

/*

package reverse

import (
	"context"
	mrpc "mrpc/runtime"
)

type ReverseServiceClient struct {
	client mrpc.client
}

func NewReverseServiceClient(client mrpc.client) *ReverseServiceClient{
	return &ReverseServiceClient{
		client: client
	}
}

func (c *ReverseServiceClient) Reverse(ctx context.Context, req *ReverseRequest) (*ReverseResponse,error) {
	var resp ReverseResponse
	err := c.client.call("ReverseService", "Reverse", &req, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}


*/

func GenerateClient(idl *IDL, out string) error {

	filename := filepath.Join(out, idl.Package+"_client.go")
	tpl := template.Must(template.ParseFiles("./generator/templates/client.tpl"))
	var buf bytes.Buffer
	err := tpl.Execute(&buf, idl)
	if err != nil {
		return err
	}

	formattedCode, err := format.Source(buf.Bytes())
	if err != nil {
		_ = os.WriteFile(filename, buf.Bytes(), 0644)
		return fmt.Errorf("生成的代码存在语法错误，无法格式化: %w", err)
	}

	return os.WriteFile(filename, formattedCode, 0644)
}

func GenerateServer(idl *IDL, out string) error {

	filename := filepath.Join(out, idl.Package+"_server.go")
	tpl := template.Must(template.ParseFiles("./generator/templates/server.tpl"))
	var buf bytes.Buffer
	err := tpl.Execute(&buf, idl)
	if err != nil {
		return err
	}

	formattedCode, err := format.Source(buf.Bytes())
	if err != nil {
		_ = os.WriteFile(filename, buf.Bytes(), 0644)
		return fmt.Errorf("生成的代码存在语法错误，无法格式化: %w", err)
	}

	return os.WriteFile(filename, formattedCode, 0644)
}
