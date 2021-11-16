package main

import (
	"flag"
	"fmt"
	"os"
	"path"
)

var (
	NBVERSION string // Notebook version string
	webroot   string
	notebook  *Notebook       // Notebook object
	config    *NotebookConfig // NotebookConfig object
)

func usage() {
	flag.PrintDefaults()
	fmt.Fprintf(os.Stderr, "Environment vars:\n  RIZIN_PATH\n    	overrides where rizin executable is installed\n")
}

func main() {

	var debug bool
	var assets string
	var bind string
	var rizinbin = "rizin"
	var dataDir string = ".rizin-notebook"

	if len(NBVERSION) < 1 {
		NBVERSION = "unknown"
	}

	if homedir, err := os.UserHomeDir(); err == nil {
		dataDir = path.Join(homedir, ".rizin-notebook")
	}

	config = NewNotebookConfig(dataDir)
	config.UpdateEnvironment()

	if loc := os.Getenv("RIZIN_PATH"); len(loc) > 1 {
		rizinbin = loc
	}

	flag.StringVar(&bind, "bind", "127.0.0.1:8000", "[address]:[port] address to bind to.")
	flag.StringVar(&webroot, "root", "/", "defines where the web root of the application is.")
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

	runServer(assets, bind, debug)
}
