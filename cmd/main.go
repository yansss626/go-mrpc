package main

import (
	"flag"
	"log"

	"github.com/yansss/mrpc/generator"
)

var (
	filePath      = flag.String("config", "mrpc.yaml", "")
	generatedPath = flag.String("out", "./generated", "")
)

func main() {
	flag.Parse()
	idl, err := generator.Parse(*filePath)
	if err != nil {
		log.Fatal(err)
	}

	err = generator.Generate(idl, *generatedPath)
	if err != nil {
		log.Fatal(err)
	}
}
