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

func findLineByKey(lines []interface{}, key, eunique string) int {
	for i, n := range lines {
		p := n.(gin.H)
		if v, ok := p[key]; ok && eunique == v {
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
	uniques []string
	storage string
	pipes   map[string]*Rizin
	jsvm    *JavaScript
	rizin   string
	cmds    map[string]RizinCommand
}

func NewNotebook(storage, rizinbin string) *Notebook {
	pages := gin.H{}
	uniques := []string{}
	prefix := path.Join(storage) + string(os.PathSeparator)
	suffix := string(os.PathSeparator) + PAGE_FILE

	cmds, err := RizinCommands(rizinbin)
	if err != nil {
		panic(err)
	}

	files, err := filepath.Glob(path.Join(storage, "*", PAGE_FILE))
	if err != nil {
		panic(err)
	}
	for _, file := range files {
		page := readJson(file)
		page["lines"] = sanitizeLines(page["lines"].([]interface{}))
		unique := strings.TrimSuffix(strings.TrimPrefix(file, prefix), suffix)
		page["unique"] = unique
		pages[unique] = page
		uniques = append(uniques, unique)
	}
	jsvm := NewJavaScript()
	if jsvm == nil {
		panic("failed to create scripting engine")
	}
	sort.Strings(uniques)
	return &Notebook{
		pages:   pages,
		uniques: uniques,
		storage: storage,
		pipes:   map[string]*Rizin{},
		jsvm:    jsvm,
		rizin:   rizinbin,
		cmds:    cmds,
	}
}

func (n *Notebook) list() []gin.H {
	n.mutex.Lock()
	defer n.mutex.Unlock()
	pages := make([]gin.H, len(n.uniques))
	for i, unique := range n.uniques {
		page := n.pages[unique].(gin.H)
		pipe := n.pipes[unique]

		pages[i] = gin.H{
			"title":  page["title"],
			"unique": unique,
			"pipe":   pipe != nil,
		}
	}

	return pages
}

func (n *Notebook) info() ([]string, error) {
	return RizinInfo(n.rizin)
}

func (n *Notebook) get(unique string) gin.H {
	if len(unique) != PAGE_NONCE_SIZE {
		return nil
	}
	n.mutex.Lock()
	defer n.mutex.Unlock()
	if page, ok := n.pages[unique]; ok {
		return page.(gin.H)
	}
	return nil
}

func (n *Notebook) newmd(unique string) string {
	if len(unique) != PAGE_NONCE_SIZE {
		return ""
	}
	n.mutex.Lock()
	defer n.mutex.Unlock()
	if data, ok := n.pages[unique]; ok {
		page := data.(gin.H)
		var eunique = Nonce(ELEMENT_NONCE_SIZE)
		for {
			if _, err := n.file(unique, eunique+".md"); err != nil {
				break
			}
			eunique = Nonce(ELEMENT_NONCE_SIZE)
		}
		if !n.save([]byte{}, unique, eunique+".md") {
			return ""
		}
		page["lines"] = append(page["lines"].([]interface{}), gin.H{
			"type":   "markdown",
			"unique": eunique,
		})
		filepath := path.Join(n.storage, unique, PAGE_FILE)
		bytes, _ := json.MarshalIndent(page, "", "\t")
		if ioutil.WriteFile(filepath, bytes, 0644) == nil {
			n.pages[unique] = page
			return eunique
		}
	}
	return ""
}

func (n *Notebook) newscript(unique, script string) string {
	if len(unique) != PAGE_NONCE_SIZE {
		return ""
	}
	n.mutex.Lock()
	defer n.mutex.Unlock()
	if data, ok := n.pages[unique]; ok {
		page := data.(gin.H)
		var eunique = Nonce(ELEMENT_NONCE_SIZE)
		for {
			if _, err := n.file(unique, eunique+".out"); err != nil {
				break
			}
			eunique = Nonce(ELEMENT_NONCE_SIZE)
		}
		page["lines"] = append(page["lines"].([]interface{}), gin.H{
			"type":   "script",
			"unique": eunique,
			"script": script,
		})
		filepath := path.Join(n.storage, unique, PAGE_FILE)
		bytes, _ := json.MarshalIndent(page, "", "\t")
		if ioutil.WriteFile(filepath, bytes, 0644) == nil {
			n.pages[unique] = page
			return eunique
		}
	}
	return ""
}

func (n *Notebook) newcmd(unique, command string) string {
	if len(unique) != PAGE_NONCE_SIZE {
		return ""
	}
	n.mutex.Lock()
	defer n.mutex.Unlock()
	if data, ok := n.pages[unique]; ok {
		page := data.(gin.H)
		var eunique = Nonce(ELEMENT_NONCE_SIZE)
		for {
			if _, err := n.file(unique, eunique+".out"); err != nil {
				break
			}
			eunique = Nonce(ELEMENT_NONCE_SIZE)
		}
		page["lines"] = append(page["lines"].([]interface{}), gin.H{
			"type":    "command",
			"unique":  eunique,
			"command": command,
		})
		filepath := path.Join(n.storage, unique, PAGE_FILE)
		bytes, _ := json.MarshalIndent(page, "", "\t")
		if ioutil.WriteFile(filepath, bytes, 0644) == nil {
			n.pages[unique] = page
			return eunique
		}
	}
	return ""
}

func (n *Notebook) deleteElem(unique, eunique string, markdown bool) bool {
	if len(unique) != PAGE_NONCE_SIZE || len(eunique) != ELEMENT_NONCE_SIZE {
		return false
	}
	n.mutex.Lock()
	defer n.mutex.Unlock()
	if data, ok := n.pages[unique]; ok {
		suffix := ".md"
		if !markdown {
			suffix = ".out"
		}
		page := data.(gin.H)
		os.Remove(path.Join(n.storage, unique, eunique+suffix))

		lines := page["lines"].([]interface{})
		if idx := findLineByKey(lines, "unique", eunique); idx > -1 {
			page["lines"] = append(lines[:idx], lines[idx+1:]...)
		}

		filepath := path.Join(n.storage, unique, PAGE_FILE)
		bytes, _ := json.MarshalIndent(page, "", "\t")
		if ioutil.WriteFile(filepath, bytes, 0644) == nil {
			n.pages[unique] = page
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
	var unique = Nonce(PAGE_NONCE_SIZE)
	for {
		if _, ok := n.pages[unique]; !ok {
			break
		}
		unique = Nonce(PAGE_NONCE_SIZE)
	}
	data := gin.H{
		"title":    title,
		"unique":   unique,
		"filename": filename,
		"binary":   binary,
		"lines":    []interface{}{},
	}
	if err := os.MkdirAll(path.Join(n.storage, unique), os.ModePerm); err != nil {
		return ""
	}
	filepath := path.Join(n.storage, unique, PAGE_FILE)
	bytes, _ := json.MarshalIndent(data, "", "\t")
	if err := ioutil.WriteFile(filepath, bytes, 0644); err != nil {
		return ""
	}
	n.pages[unique] = data
	n.uniques = append(n.uniques, unique)
	sort.Strings(n.uniques)
	return unique
}

func (n *Notebook) rename(unique, title string) bool {
	if len(title) < 1 || len(unique) != PAGE_NONCE_SIZE {
		return false
	}
	n.mutex.Lock()
	defer n.mutex.Unlock()

	if data, ok := n.pages[unique]; ok {
		page := data.(gin.H)
		page["title"] = title
		filepath := path.Join(n.storage, unique, PAGE_FILE)
		bytes, _ := json.MarshalIndent(page, "", "\t")
		if ioutil.WriteFile(filepath, bytes, 0644) == nil {
			n.pages[unique] = page
			return true
		}
	}
	return false
}

func (n *Notebook) delete(unique string) bool {
	if len(unique) != PAGE_NONCE_SIZE {
		return false
	}
	if _, ok := n.pages[unique]; !ok {
		return false
	}

	filepath := path.Join(n.storage, unique)
	if os.RemoveAll(filepath) == nil {
		delete(n.pages, unique)
		idx := findNonce(n.uniques, unique)
		n.uniques = append(n.uniques[:idx], n.uniques[idx+1:]...)
		return true
	}
	return false
}

func (n *Notebook) open(unique string, open bool) *Rizin {
	if len(unique) != PAGE_NONCE_SIZE {
		return nil
	}
	n.mutex.Lock()
	defer n.mutex.Unlock()
	if p, ok := n.pipes[unique]; ok {
		return p
	}
	if p, ok := n.pages[unique]; ok && open {
		page := p.(gin.H)
		filepath := path.Join(n.storage, unique, page["binary"].(string))
		prjpath := path.Join(n.storage, unique, "project.rzdb")
		// This should be async.
		rizin := NewRizin(n.rizin, filepath, prjpath)
		if rizin == nil {
			return nil
		}
		n.pipes[unique] = rizin
		return rizin
	}
	return nil
}

func (n *Notebook) close(unique string) {
	if len(unique) != PAGE_NONCE_SIZE {
		return
	}
	n.mutex.Lock()
	defer n.mutex.Unlock()
	if p, ok := n.pipes[unique]; ok {
		p.close()
		delete(n.pipes, unique)
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
