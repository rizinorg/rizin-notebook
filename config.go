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

type KBindings map[string]map[string]string

type NotebookConfig struct {
	Environment map[string]string `json:"environment"`
	KeyBindings KBindings         `json:"keybindings"`
	filename    string
	mutex       sync.Mutex
}

var availableKeyBindings = KBindings{
	KB_INDEX: {
		"New Page": "Control,N",
		"About":    "Control,H",
		"Settings": "Control,S",
	},
	KB_PAGE: {
		"Open/Close Pipe": "Control,O",
		"New Markdown":    "Control,M",
		"Execute Command": "Control,E",
	},
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

func sanitizeKeyBindings(keyBindings KBindings) KBindings {
	sanitized := KBindings{}
	for section, kb := range availableKeyBindings {
		msection := map[string]string{}
		fsection := keyBindings[section]
		for action, key := range kb {
			msection[action] = getValue(fsection, action, key)
		}
		sanitized[section] = msection
	}
	return sanitized
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
	if config.KeyBindings == nil {
		config.KeyBindings = availableKeyBindings
		config.Save()
	} else {
		config.KeyBindings = sanitizeKeyBindings(config.KeyBindings)
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

func (nc *NotebookConfig) SetKeyBindings(section, key, value string) bool {
	nc.mutex.Lock()
	defer nc.mutex.Unlock()
	value = strings.TrimSpace(value)
	key = strings.TrimSpace(key)
	if _, ok := nc.KeyBindings[section]; !ok {
		return false
	}
	if _, ok := nc.KeyBindings[section][key]; !ok {
		return false
	}
	nc.KeyBindings[section][key] = value
	return true
}

func (nc *NotebookConfig) Save() {
	nc.mutex.Lock()
	defer nc.mutex.Unlock()
	bytes, _ := json.MarshalIndent(nc, "", "\t")
	if err := ioutil.WriteFile(nc.filename, bytes, 0644); err != nil {
		fmt.Println(err)
	}
}
