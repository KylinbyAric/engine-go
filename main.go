package main

import (
	"embed"
	"flag"
	"io/fs"
	"log"

	"github.com/engine-go/workflow/api"
)

//go:embed public/static
var embeddedStatic embed.FS

func main() {
	addr := flag.String("addr", ":8080", "http listen addr")
	flag.Parse()

	sub, err := fs.Sub(embeddedStatic, "public/static")
	if err != nil {
		log.Fatalf("sub static fs: %v", err)
	}
	log.Printf("engine-go listening on %s", *addr)
	if err := api.Run(*addr, sub); err != nil {
		log.Fatalf("api server: %v", err)
	}
}
