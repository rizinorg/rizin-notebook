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

func sanitizeLines(lines []interface{}) []interface{} {
	for i, v := range lines {
		n := gin.H{}
		p := v.(map[string]interface{})
		for k, _ := range p {
			n[k] = p[k]
		}
		lines[i] = n
	}
	return lines
}

func findLineByKey(lines []interface{}, key, enonce string) int {
	for i, n := range lines {
		p := n.(gin.H)
		if v, ok := p[key]; ok && enonce == v {
			return i
		}
	}
	return -1
}

func findNonce(a []string, b string) int {
	for i, n := range a {
		if n == b {
			return i
		}
	}
	return -1
}

type Notebook struct {
	mutex   sync.Mutex
	pages   gin.H
	nonces  []string
	storage string
	pipes   map[string]*Rizin
	rizin   string
}

func NewNotebook(storage, rizinbin string) *Notebook {
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
		page["lines"] = sanitizeLines(page["lines"].([]interface{}))
		nonce := strings.TrimSuffix(strings.TrimPrefix(file, prefix), suffix)
		page["nonce"] = nonce
		pages[nonce] = page
		nonces = append(nonces, nonce)
	}
	sort.Strings(nonces)
	return &Notebook{
		pages:   pages,
		nonces:  nonces,
		storage: storage,
		pipes:   map[string]*Rizin{},
		rizin:   rizinbin,
	}
}

func (n *Notebook) list() []gin.H {
	n.mutex.Lock()
	defer n.mutex.Unlock()
	pages := make([]gin.H, len(n.nonces))
	for i, nonce := range n.nonces {
		page := n.pages[nonce].(gin.H)
		pipe := n.pipes[nonce]

		pages[i] = gin.H{
			"title": page["title"],
			"nonce": nonce,
			"pipe":  pipe != nil,
		}
	}

	return pages
}

func (n *Notebook) get(nonce string) gin.H {
	if len(nonce) != PAGE_NONCE_SIZE {
		return nil
	}
	n.mutex.Lock()
	defer n.mutex.Unlock()
	if page, ok := n.pages[nonce]; ok {
		return page.(gin.H)
	}
	return nil
}

func (n *Notebook) newmd(nonce string) string {
	if len(nonce) != PAGE_NONCE_SIZE {
		return ""
	}
	n.mutex.Lock()
	defer n.mutex.Unlock()
	if data, ok := n.pages[nonce]; ok {
		page := data.(gin.H)
		var enonce = Nonce(ELEMENT_NONCE_SIZE)
		for {
			if _, err := n.file(nonce, enonce+".md"); err != nil {
				break
			}
			enonce = Nonce(ELEMENT_NONCE_SIZE)
		}
		if !n.save([]byte{}, nonce, enonce+".md") {
			return ""
		}
		page["lines"] = append(page["lines"].([]interface{}), gin.H{
			"type":  "markdown",
			"nonce": enonce,
		})
		filepath := path.Join(n.storage, nonce, PAGE_FILE)
		bytes, _ := json.MarshalIndent(page, "", "\t")
		if ioutil.WriteFile(filepath, bytes, 0644) == nil {
			n.pages[nonce] = page
			return enonce
		}
	}
	return ""
}

func (n *Notebook) newcmd(nonce, command string) string {
	if len(nonce) != PAGE_NONCE_SIZE {
		return ""
	}
	n.mutex.Lock()
	defer n.mutex.Unlock()
	if data, ok := n.pages[nonce]; ok {
		page := data.(gin.H)
		var enonce = Nonce(ELEMENT_NONCE_SIZE)
		for {
			if _, err := n.file(nonce, enonce+".out"); err != nil {
				break
			}
			enonce = Nonce(ELEMENT_NONCE_SIZE)
		}
		page["lines"] = append(page["lines"].([]interface{}), gin.H{
			"type":    "command",
			"nonce":   enonce,
			"command": command,
		})
		filepath := path.Join(n.storage, nonce, PAGE_FILE)
		bytes, _ := json.MarshalIndent(page, "", "\t")
		if ioutil.WriteFile(filepath, bytes, 0644) == nil {
			n.pages[nonce] = page
			return enonce
		}
	}
	return ""
}

func (n *Notebook) deleteElem(nonce, enonce string, markdown bool) bool {
	if len(nonce) != PAGE_NONCE_SIZE || len(enonce) != ELEMENT_NONCE_SIZE {
		return false
	}
	n.mutex.Lock()
	defer n.mutex.Unlock()
	if data, ok := n.pages[nonce]; ok {
		suffix := ".md"
		if !markdown {
			suffix = ".out"
		}
		page := data.(gin.H)
		if os.Remove(path.Join(n.storage, nonce, enonce+suffix)) != nil {
			return false
		}

		lines := page["lines"].([]interface{})
		if idx := findLineByKey(lines, "nonce", enonce); idx > -1 {
			page["lines"] = append(lines[:idx], lines[idx+1:]...)
		}

		filepath := path.Join(n.storage, nonce, PAGE_FILE)
		bytes, _ := json.MarshalIndent(page, "", "\t")
		if ioutil.WriteFile(filepath, bytes, 0644) == nil {
			n.pages[nonce] = page
			return true
		}
	}
	return false
}

func (n *Notebook) new(title, filename, binary string) string {
	if len(title) < 1 {
		return ""
	}

	n.mutex.Lock()
	defer n.mutex.Unlock()
	var nonce = Nonce(PAGE_NONCE_SIZE)
	for {
		if _, ok := n.pages[nonce]; !ok {
			break
		}
		nonce = Nonce(PAGE_NONCE_SIZE)
	}
	data := gin.H{
		"title":    title,
		"nonce":    nonce,
		"filename": filename,
		"binary":   binary,
		"lines":    []interface{}{},
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
	n.nonces = append(n.nonces, nonce)
	sort.Strings(n.nonces)
	return nonce
}

func (n *Notebook) rename(nonce, title string) bool {
	if len(title) < 1 || len(nonce) != PAGE_NONCE_SIZE {
		return false
	}
	n.mutex.Lock()
	defer n.mutex.Unlock()

	if data, ok := n.pages[nonce]; ok {
		page := data.(gin.H)
		page["title"] = title
		filepath := path.Join(n.storage, nonce, PAGE_FILE)
		bytes, _ := json.MarshalIndent(page, "", "\t")
		if ioutil.WriteFile(filepath, bytes, 0644) == nil {
			n.pages[nonce] = page
			return true
		}
	}
	return false
}

func (n *Notebook) delete(nonce string) bool {
	if len(nonce) != PAGE_NONCE_SIZE {
		return false
	}
	if _, ok := n.pages[nonce]; !ok {
		return false
	}

	filepath := path.Join(n.storage, nonce)
	if os.RemoveAll(filepath) == nil {
		delete(n.pages, nonce)
		idx := findNonce(n.nonces, nonce)
		n.nonces = append(n.nonces[:idx], n.nonces[idx+1:]...)
		return true
	}
	return false
}

func (n *Notebook) open(nonce string, open bool) *Rizin {
	if len(nonce) != PAGE_NONCE_SIZE {
		return nil
	}
	n.mutex.Lock()
	defer n.mutex.Unlock()
	if p, ok := n.pipes[nonce]; ok {
		return p
	}
	if p, ok := n.pages[nonce]; ok && open {
		page := p.(gin.H)
		filepath := path.Join(n.storage, nonce, page["binary"].(string))
		prjpath := path.Join(n.storage, nonce, "project.rzdb")
		// This should be async.
		rizin := NewRizin(n.rizin, filepath, prjpath)
		if rizin == nil {
			return nil
		}
		n.pipes[nonce] = rizin
		return rizin
	}
	return nil
}

func (n *Notebook) close(nonce string) {
	if len(nonce) != PAGE_NONCE_SIZE {
		return
	}
	n.mutex.Lock()
	defer n.mutex.Unlock()
	if p, ok := n.pipes[nonce]; ok {
		p.close()
		delete(n.pipes, nonce)
	}
}

func (n *Notebook) exists(paths ...string) bool {
	args := append([]string{n.storage}, paths...)
	_, err := os.Stat(path.Join(args...))
	if os.IsNotExist(err) {
		return false
	}
	return true
}

func (n *Notebook) file(paths ...string) ([]byte, error) {
	args := append([]string{n.storage}, paths...)
	filepath := path.Join(args...)
	return ioutil.ReadFile(filepath)
}

func (n *Notebook) save(bytes []byte, paths ...string) bool {
	args := append([]string{n.storage}, paths...)
	filepath := path.Join(args...)
	return ioutil.WriteFile(filepath, bytes, 0644) == nil
}
