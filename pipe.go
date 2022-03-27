package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"sync"
)

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
