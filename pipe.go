package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"sync"
)

type RizinCommandArg struct {
	Type       string   `json:"type"`
	Name       string   `json:"name"`
	DefaultArg string   `json:"default,omitempty"`
	Required   bool     `json:"required,omitempty"`
	IsOption   bool     `json:"is_option,omitempty"`
	IsArray    bool     `json:"is_array,omitempty"`
	Choices    []string `json:"choices,omitempty"`
}

type RizinCommandDetailEntry struct {
	Text    string `json:"text,omitempty"`
	Comment string `json:"comment,omitempty"`
	Arg     string `json:"arg_str,omitempty"`
}

type RizinCommandDetail struct {
	Name    string                    `json:"name"`
	Entries []RizinCommandDetailEntry `json:"entries,omitempty"`
}

type RizinCommand struct {
	Command     string               `json:"cmd"`
	ArgsStr     string               `json:"args_str"`
	Args        []RizinCommandArg    `json:"args,omitempty"`
	Description string               `json:"description,omitempty"`
	Summary     string               `json:"summary,omitempty"`
	Details     []RizinCommandDetail `json:"details,omitempty"`
}

type Rizin struct {
	pipe    *exec.Cmd
	stdin   io.WriteCloser
	stdout  io.ReadCloser
	mutex   sync.Mutex
	project string
}

func RizinInfo(rizinbin string) ([]string, error) {
	out, err := exec.Command(rizinbin, "-version").Output()
	if err != nil {
		return nil, err
	}
	return strings.Split(string(out), "\n"), nil
}

func RizinCommands(rizinbin string) (map[string]RizinCommand, error) {
	out, err := exec.Command(rizinbin, "-qc", "?*j").Output()
	if err != nil {
		return nil, err
	}

	var commands map[string]RizinCommand
	err = json.Unmarshal(out, &commands)
	return commands, err
}

func NewRizin(rizinbin, file, project string) *Rizin {
	args := []string{
		"-2",
		"-0",
		"-e",
		"scr.color=3",
		"-p",
		project,
		file,
	}
	pipe := exec.Command(rizinbin, args...)

	stdin, err := pipe.StdinPipe()
	if err != nil {
		fmt.Println("pipe error:", err)
		return nil
	}

	stdout, err := pipe.StdoutPipe()
	if err != nil {
		fmt.Println("pipe error:", err)
		return nil
	}

	if err := pipe.Start(); err != nil {
		fmt.Println("pipe error:", err)
		return nil
	}

	rizin := &Rizin{pipe: pipe, stdin: stdin, stdout: stdout, project: project}

	go func(r *Rizin) {
		r.mutex.Lock()
		if _, err := bufio.NewReader(r.stdout).ReadString('\x00'); err != nil {
			fmt.Println("pipe error:", err)
		}
		r.mutex.Unlock()
	}(rizin)
	return rizin
}

func (r *Rizin) getCommands(cmd string) (string, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	if _, err := fmt.Fprintln(r.stdin, cmd); err != nil {
		fmt.Println("pipe error:", err)
		return "", err
	}
	buf, err := bufio.NewReader(r.stdout).ReadString('\x00')
	if err != nil && err != io.EOF {
		fmt.Println("pipe error:", err)
		return "", err
	}
	buf = string(bytes.Trim([]byte(buf), "\x00"))
	return buf, nil
}

func (r *Rizin) exec(cmd string) (string, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	if _, err := fmt.Fprintln(r.stdin, cmd); err != nil {
		fmt.Println("pipe error:", err)
		return "", err
	}
	buf, err := bufio.NewReader(r.stdout).ReadString('\x00')
	if err != nil && err != io.EOF {
		fmt.Println("pipe error:", err)
		return "", err
	}
	buf = string(bytes.Trim([]byte(buf), "\x00"))
	return buf, nil
}

func (r *Rizin) close() {
	r.exec("Ps " + r.project)
	r.exec("q!")
}
