package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"sync"
)

const (
	CONFIG_FILE = "config.json"
	KB_INDEX    = "Index"
	KB_PAGE     = "Page"
)

type NotebookConfig struct {
	Environment map[string]string `json:"environment"`
	filename    string
	mutex       sync.Mutex
}

func getValue(kmap map[string]string, action, defkey string) string {
	if kmap == nil {
		return defkey
	}
	if value, ok := kmap[action]; ok && len(value) > 0 {
		return value
	}
	return defkey
}

func NewNotebookConfig(folder string) *NotebookConfig {
	var config = &NotebookConfig{}
	config.filename = path.Join(folder, CONFIG_FILE)
	bytes, err := ioutil.ReadFile(config.filename)
	if err == nil {
		json.Unmarshal(bytes, config)
	}
	if config.Environment == nil {
		config.Environment = map[string]string{}
	}

	if value, ok := config.Environment["RIZIN_PATH"]; !ok || len(value) < 1 {
		config.Environment["RIZIN_PATH"] = os.Getenv("RIZIN_PATH")
	}
	return config
}

func (nc *NotebookConfig) UpdateEnvironment() {
	nc.mutex.Lock()
	defer nc.mutex.Unlock()
	for key, value := range nc.Environment {
		os.Setenv(key, value)
	}
}

func (nc *NotebookConfig) DelEnvironment(key string) {
	nc.mutex.Lock()
	defer nc.mutex.Unlock()
	key = strings.TrimSpace(key)
	delete(nc.Environment, key)
	os.Unsetenv(key)
}

func (nc *NotebookConfig) SetEnvironment(key, value string) {
	nc.mutex.Lock()
	defer nc.mutex.Unlock()
	value = strings.TrimSpace(value)
	key = strings.TrimSpace(key)
	os.Setenv(key, value)
	nc.Environment[key] = value
}

func (nc *NotebookConfig) Save() {
	nc.mutex.Lock()
	defer nc.mutex.Unlock()
	bytes, _ := json.MarshalIndent(nc, "", "\t")
	if err := ioutil.WriteFile(nc.filename, bytes, 0644); err != nil {
		fmt.Println(err)
	}
}
