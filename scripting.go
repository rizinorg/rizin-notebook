package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dop251/goja"
	"golang.org/x/sync/semaphore"
	"time"
)

type JavaScript struct {
	semaphore *semaphore.Weighted
	runtime   *goja.Runtime
	rizin     *Rizin
	output    string
}

func rizinCmd(command string, rizin *Rizin) (string, error) {
	rizin.exec("scr.color=0")
	result, err := rizin.exec(command)
	rizin.exec("scr.color=3")
	return result, err
}

func convertValue(ivalue interface{}) string {
	switch value := ivalue.(type) {
	case []interface{}:
		bytes, _ := json.MarshalIndent(value, "", "\t")
		return string(bytes)
	case map[string]interface{}:
		bytes, _ := json.MarshalIndent(value, "", "\t")
		return string(bytes)
	default:
		return fmt.Sprintf("%v", value)
	}
}

func NewJavaScript() *JavaScript {
	runtime := goja.New()
	if runtime == nil {
		fmt.Println("cannot create a JavaScript runtime.")
		return nil
	}
	sem := semaphore.NewWeighted(1)
	js := &JavaScript{semaphore: sem, runtime: runtime, rizin: nil}
	rizin := map[string]interface{}{}
	rizin["cmd"] = func(args ...interface{}) goja.Value {
		if js.rizin == nil {
			panic(js.runtime.ToValue("Rizin pipe is closed."))
		} else if len(args) < 1 {
			panic(js.runtime.ToValue("No string was passed."))
		}
		switch value := args[0].(type) {
		case string:
			result, err := rizinCmd(value, js.rizin)
			if err != nil {
				panic(js.runtime.ToValue(err))
			}
			return js.runtime.ToValue(result)
		default:
			panic(js.runtime.ToValue("input is not a string."))
		}
	}
	rizin["cmdj"] = func(args ...interface{}) goja.Value {
		if js.rizin == nil {
			panic(js.runtime.ToValue("Rizin pipe is closed."))
		} else if len(args) < 1 {
			panic(js.runtime.ToValue("No string was passed."))
		}
		switch value := args[0].(type) {
		case string:
			result, err := rizinCmd(value, js.rizin)
			if err != nil {
				panic(js.runtime.ToValue(err))
			}
			var data interface{}
			err = json.Unmarshal([]byte(result), &data)
			if err != nil {
				panic(js.runtime.ToValue(err))
			}
			return js.runtime.ToValue(data)
		default:
			panic(js.runtime.ToValue("input is not a string."))
		}
	}

	console := map[string]interface{}{}
	console["log"] = func(args ...interface{}) {
		n_args := len(args)
		for i, value := range args {
			js.output += convertValue(value)
			if i+1 < n_args {
				js.output += " "
			}
		}
		js.output += "\n"
	}

	runtime.Set("rizin", rizin)
	runtime.Set("console", console)
	return js
}

func (js *JavaScript) exec(script string, rizin *Rizin) (string, error) {
	if !js.semaphore.TryAcquire(1) {
		return "", errors.New("A script is already running.")
	}
	defer js.semaphore.Release(1)
	js.rizin = rizin
	js.output = ""
	timer := time.AfterFunc(5*time.Minute, func() {
		js.runtime.Interrupt("The script execution has timed out.")
	})
	_, err := js.runtime.RunScript("script.js", script)
	result := js.output
	js.rizin = nil
	js.output = ""
	timer.Stop()
	return result, err
}
