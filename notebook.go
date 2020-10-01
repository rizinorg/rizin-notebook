package main

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"sync"
)

const PAGE_FILE = "page.json"

func readJson(filepath string) gin.H {
	var data = gin.H{}
	bytes, err := ioutil.ReadFile(filepath)
	if err != nil {
		panic(err)
	}
	json.Unmarshal(bytes, &data)
	return data
}

type Notebook struct {
	mutex   sync.Mutex
	pages   gin.H
	nonces  []string
	storage string
}

func NewNotebook(storage string) *Notebook {
	pages := gin.H{}
	nonces := []string{}
	prefix := path.Join(storage) + string(os.PathSeparator)
	suffix := string(os.PathSeparator) + PAGE_FILE

	files, err := filepath.Glob(path.Join(storage, "*", PAGE_FILE))
	if err != nil {
		panic(err)
	}
	for _, file := range files {
		page := readJson(file)
		nonce := strings.TrimSuffix(strings.TrimPrefix(file, prefix), suffix)
		page["nonce"] = nonce
		pages[nonce] = page
		nonces = append(nonces, nonce)
	}
	sort.Strings(nonces)
	return &Notebook{pages: pages, nonces: nonces, storage: storage}
}

func (n *Notebook) list() []gin.H {
	n.mutex.Lock()
	defer n.mutex.Unlock()
	pages := make([]gin.H, len(n.nonces))
	for i, nonce := range n.nonces {
		page := n.pages[nonce].(gin.H)
		pages[i] = gin.H{
			"title": page["title"],
			"nonce": nonce,
		}
	}

	return pages
}

func (n *Notebook) get(nonce string) gin.H {
	n.mutex.Lock()
	defer n.mutex.Unlock()
	if page, ok := n.pages[nonce]; ok {
		return page.(gin.H)
	}
	return nil
}

func (n *Notebook) save(nonce string) bool {
	n.mutex.Lock()
	defer n.mutex.Unlock()
	if data, ok := n.pages[nonce]; ok {
		filepath := path.Join(n.storage, nonce, PAGE_FILE)
		bytes, _ := json.MarshalIndent(data, "", "\t")
		return ioutil.WriteFile(filepath, bytes, 0644) == nil
	}
	return false
}

func (n *Notebook) new() string {
	n.mutex.Lock()
	defer n.mutex.Unlock()
	var nonce = Nonce(NONCE_SIZE)
	for {
		if _, ok := n.pages[nonce]; !ok {
			break
		}
		nonce = Nonce(NONCE_SIZE)
	}
	data := gin.H{
		"title": "untitled page",
		"nonce": nonce,
		"lines": []string{},
	}
	if err := os.MkdirAll(path.Join(n.storage, nonce), os.ModePerm); err != nil {
		return ""
	}
	filepath := path.Join(n.storage, nonce, PAGE_FILE)
	bytes, _ := json.MarshalIndent(data, "", "\t")
	if err := ioutil.WriteFile(filepath, bytes, 0644); err != nil {
		return ""
	}
	n.pages[nonce] = data
	return nonce
}

func (n *Notebook) file(paths ...string) ([]byte, error) {
	args := append([]string{n.storage}, paths...)
	filepath := path.Join(args...)
	return ioutil.ReadFile(filepath)
}
