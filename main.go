package main

import (
	"flag"
	"fmt"
	"os"
	"path"
)

var (
	DEFAULT_RIZIN_PATH = os.Getenv("RIZIN_PATH")
	notebook           *Notebook
)

func main() {

	var webRoot string
	var assets string
	var bind string
	var dataDir string = ".rizin-notebook"

	if homedir, err := os.UserHomeDir(); err == nil {
		dataDir = path.Join(homedir, ".rizin-notebook")
	}

	flag.StringVar(&webRoot, "web-root", "/", "web root of the application")
	flag.StringVar(&assets, "assets", "assets", "assets directory path")
	flag.StringVar(&dataDir, "data-dir", ""+dataDir, "web root of the application")
	flag.StringVar(&bind, "bind", "127.0.0.1:8000", "[address]:[port] address to bind to")
	flag.Parse()

	if err := os.MkdirAll(dataDir, os.ModePerm); err != nil && !os.IsExist(err) {
		panic(err)
	}

	fmt.Printf("Server data dir '%s'\n", dataDir)
	notebook = NewNotebook(dataDir)

	server(webRoot, assets, bind)
}
