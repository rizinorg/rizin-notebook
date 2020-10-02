package main

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"sync"
)

type Rizin struct {
	pipe   *exec.Cmd
	stdin  io.WriteCloser
	stdout io.ReadCloser
	mutex  sync.Mutex
}

func NewRizin(file string) *Rizin {
	pipe := exec.Command("radare2", "-2", "-e", "scr.color=0", "-A", file)

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

	rizin := &Rizin{pipe: pipe, stdin: stdin, stdout: stdout}

	go func(r *Rizin) {
		r.mutex.Lock()
		defer r.mutex.Unlock()
		if _, err := bufio.NewReader(r.stdout).ReadString('\x00'); err != nil {
			fmt.Println("pipe error:", err)
		}
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
	idx := strings.Index(buf, cmd+"\x1b")
	if idx > -1 {
		tmp := buf[idx+len(cmd)+4:]
		idx = strings.Index(tmp, "\x1b")
		if idx > -1 {
			return strings.Trim(tmp[:idx], "\n"), nil
		}
		return strings.Trim(tmp, "\n"), nil
	}
	return strings.Trim(buf, "\n"), nil
}

func (r *Rizin) close() {
	r.exec("q!")
}
