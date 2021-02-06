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
	output.POST("/exec/:nonce", func(c *gin.Context) {
		nonce := c.Param("nonce")
		command := c.DefaultPostForm("command", "")
		if !IsValidNonce(nonce) || len(command) < 1 {
			c.HTML(400, "console-error.tmpl", gin.H{
				"error": "invalid request",
				"root":  webroot,
			})
			return
		}

		rizin := notebook.open(nonce, false)
		if rizin == nil {
			c.HTML(400, "console-error.tmpl", gin.H{
				"error": "pipe is closed.",
				"root":  webroot,
			})
			return
		}

		enonce := notebook.newcmd(nonce, command)
		if len(enonce) < 1 {
			c.HTML(400, "console-error.tmpl", gin.H{
				"error": "cannot create enonce for output.",
				"root":  webroot,
			})
			return
		}

		go func(nonce, enonce, command string, rizin *Rizin) {
			output, err := rizin.exec(command)
			if len(strings.TrimSpace(output)) < 1 && err == nil {
				output = "no output from rizin."
			} else if err != nil {
				output = fmt.Sprintf("pipe error: %v", err)
			}
			notebook.save([]byte(output), nonce, enonce+".out")
		}(nonce, enonce, command, rizin)

		c.Redirect(302, webroot+"output/check/"+nonce+"/"+enonce)
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
	output.GET("/input/:nonce", func(c *gin.Context) {
		nonce := c.Param("nonce")
		if !IsValidNonce(nonce) {
			c.HTML(400, "error.tmpl", gin.H{
				"root":  webroot,
				"error": "invalid request",
			})
		} else {
			c.HTML(200, "console.tmpl", gin.H{
				"nonce": nonce,
				"root":  webroot,
			})
		}
	})
	output.GET("/view/*path", func(c *gin.Context) {
		tokens := strings.Split(c.Param("path")[1:], "/")
		if len(tokens) != 2 || !IsValidNonce(tokens[0]) || !IsValidNonce(tokens[1]) {
			c.String(400, "invalid request")
		} else {
			bytes, err := notebook.file(tokens[0], tokens[1]+".out")
			if err != nil {
				c.String(404, "file not found")
			} else {
				c.String(200, string(bytes))
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
