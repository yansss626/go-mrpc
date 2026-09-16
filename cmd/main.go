package main

import (
	"log"
	"mrpc/generator"
)

const (
	Filepath = "./mrpc.yaml"
)

func main() {
	idl, err := generator.Parse(Filepath)
	if err != nil {
		log.Fatal(err)
	}

	err = generator.Generate(idl)
	if err != nil {
		log.Fatal(err)
	}
}
