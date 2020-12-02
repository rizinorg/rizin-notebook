package main

import (
	"flag"
	"fmt"
	"os"
	"path"
)

var (
	notebook  *Notebook // Notebook object
	NBVERSION string    // Notebook version string
)

func usage() {
	flag.PrintDefaults()
	fmt.Fprintf(os.Stderr, "Environment vars:\n  RIZIN_PATH\n    	overrides where rizin executable is installed\n")
}

func main() {

	var debug bool
	var root string
	var assets string
	var bind string
	var rizinbin = "rizin"
	var dataDir string = ".rizin-notebook"

	if len(NBVERSION) < 1 {
		NBVERSION = "unknown"
	}

	if loc := os.Getenv("RIZIN_PATH"); len(loc) > 1 {
		rizinbin = loc
	}

	if homedir, err := os.UserHomeDir(); err == nil {
		dataDir = path.Join(homedir, ".rizin-notebook")
	}

	flag.StringVar(&bind, "bind", "127.0.0.1:8000", "[address]:[port] address to bind to.")
	flag.StringVar(&root, "root", "/", "defines where the web root of the application is.")
	flag.StringVar(&dataDir, "notebook", ""+dataDir, "defines where the notebook folder is located.")
	flag.StringVar(&assets, "debug-assets", "", "allows you to debug the assets (-debug-assets /path/to/assets).")
	flag.BoolVar(&debug, "debug", false, "enable http debug logs.")
	flag.Usage = usage
	flag.Parse()

	if err := os.MkdirAll(dataDir, os.ModePerm); err != nil && !os.IsExist(err) {
		panic(err)
	}

	fmt.Printf("Server data dir '%s'\n", dataDir)
	notebook = NewNotebook(dataDir, rizinbin)

	runServer(root, assets, bind, debug)
}
