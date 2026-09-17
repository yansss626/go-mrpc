package main

import (
	"flag"
	"log"
	"mrpc/generator"
)

var (
	filePath = flag.String("config", "mrpc.yaml", "")
)

func main() {
	flag.Parse()
	idl, err := generator.Parse(*filePath)
	if err != nil {
		log.Fatal(err)
	}

	err = generator.Generate(idl)
	if err != nil {
		log.Fatal(err)
	}
}
