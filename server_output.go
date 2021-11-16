package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"strings"
)

func serverAddOutput(output *gin.RouterGroup) {
	output.GET("/deleted", func(c *gin.Context) {
		c.String(200, "deleted.")
	})
	output.GET("/loaded", func(c *gin.Context) {
		c.String(200, "found.")
	})
	output.POST("/exec/:unique", func(c *gin.Context) {
		unique := c.Param("unique")
		command := c.DefaultPostForm("command", "")
		if !IsValidNonce(unique) || len(command) < 1 {
			c.HTML(400, "console-error.tmpl", gin.H{
				"error": "invalid request",
				"root":  webroot,
			})
			return
		}

		rizin := notebook.open(unique, false)
		if rizin == nil {
			c.HTML(400, "console-error.tmpl", gin.H{
				"error": "pipe is closed.",
				"root":  webroot,
			})
			return
		}

		section := notebook.newcmd(unique, command)
		if len(section) < 1 {
			c.HTML(400, "console-error.tmpl", gin.H{
				"error": "cannot create section for output.",
				"root":  webroot,
			})
			return
		}

		go func(unique, name, command string, rizin *Rizin) {
			output, err := rizin.exec(command)
			if err != nil {
				output = fmt.Sprintf("pipe error: %v", err)
			}
			notebook.save([]byte(output), unique, name+".out")
		}(unique, section, command, rizin)

		c.Redirect(302, webroot+"output/check/"+unique+"/"+section)
	})
	output.GET("/check/*path", func(c *gin.Context) {
		tokens := strings.Split(c.Param("path")[1:], "/")
		if len(tokens) != 2 || !IsValidNonce(tokens[0]) || !IsValidNonce(tokens[1]) {
			c.HTML(400, "error.tmpl", gin.H{
				"root":  webroot,
				"error": "invalid request",
			})
		} else {
			_, err := notebook.file(tokens[0], tokens[1]+".out")
			if err != nil {
				c.HTML(200, "reload.tmpl", gin.H{
					"root": webroot,
				})
			} else {
				c.Redirect(302, webroot+"output/loaded")
			}
		}
	})
	output.GET("/input/:unique", func(c *gin.Context) {
		unique := c.Param("unique")
		if !IsValidNonce(unique) {
			c.HTML(400, "error.tmpl", gin.H{
				"root":  webroot,
				"error": "invalid request",
			})
		} else {
			c.HTML(200, "console.tmpl", gin.H{
				"unique": unique,
				"root":   webroot,
			})
		}
	})
	output.GET("/view/*path", func(c *gin.Context) {
		tokens := strings.Split(c.Param("path")[1:], "/")
		if len(tokens) != 2 || !IsValidNonce(tokens[0]) || !IsValidNonce(tokens[1]) {
			c.HTML(400, "output.tmpl", gin.H{
				"root":   webroot,
				"output": []byte("invalid request"),
			})
		} else {
			bytes, err := notebook.file(tokens[0], tokens[1]+".out")
			if err != nil {
				c.HTML(404, "output.tmpl", gin.H{
					"root":   webroot,
					"output": []byte("missing output file"),
				})
			} else {
				output := toHtml(bytes)
				c.HTML(200, "output.tmpl", gin.H{
					"root":   webroot,
					"output": output,
				})
			}
		}
	})
	output.GET("/delete/*path", func(c *gin.Context) {
		path := c.Param("path")
		tokens := strings.Split(path[1:], "/")
		if len(tokens) != 2 || !IsValidNonce(tokens[0]) || !IsValidNonce(tokens[1]) {
			c.HTML(400, "error.tmpl", gin.H{
				"root":  webroot,
				"error": "invalid request",
			})
		} else {
			if notebook.deleteElem(tokens[0], tokens[1], false) {
				c.Redirect(302, webroot+"view/"+tokens[0])
			} else {
				c.HTML(400, "error.tmpl", gin.H{
					"root":  webroot,
					"error": "invalid request",
				})
			}
		}
	})
}
