package main

import (
	"fmt"
	"log"
	"os"

	document "service-go-document/internal"

	"ariga.io/atlas-provider-gorm/gormschema"
)

func main() {
	loader := gormschema.New("postgres")
	schema, err := loader.Load(document.Models()...)
	if err != nil {
		log.Fatalf("schema load error: %v", err)
	}
	if _, err := fmt.Fprint(os.Stdout, schema); err != nil {
		log.Fatalf("schema write error: %v", err)
	}
}
