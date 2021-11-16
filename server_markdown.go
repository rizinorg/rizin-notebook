package main

import (
	"github.com/gin-gonic/gin"
	"strings"
)

func serverAddMarkdown(markdown *gin.RouterGroup) {
	markdown.GET("/deleted", func(c *gin.Context) {
		c.String(200, "deleted.")
	})
	markdown.GET("/view/*path", func(c *gin.Context) {
		path := c.Param("path")
		tokens := strings.Split(path[1:], "/")
		if len(tokens) != 2 || !IsValidNonce(tokens[0]) || !IsValidNonce(tokens[1]) {
			c.String(400, "invalid request")
		} else {
			bytes, err := notebook.file(tokens[0], tokens[1]+".md")
			if err != nil {
				c.String(404, "file not found")
			} else {
				c.HTML(200, "markdown-view.tmpl", gin.H{
					"root": webroot,
					"html": string(bytes),
					"path": "/" + tokens[0] + "/" + tokens[1],
				})
			}
		}
	})
	markdown.GET("/edit/*path", func(c *gin.Context) {
		path := c.Param("path")
		tokens := strings.Split(path[1:], "/")
		if len(tokens) != 2 || !IsValidNonce(tokens[0]) || !IsValidNonce(tokens[1]) {
			c.String(400, "invalid request")
		} else {
			bytes, err := notebook.file(tokens[0], tokens[1]+".md")
			if err != nil {
				c.String(404, "file not found")
			} else {
				c.HTML(200, "markdown-edit.tmpl", gin.H{
					"root": webroot,
					"raw":  bytes,
					"path": "/" + tokens[0] + "/" + tokens[1],
				})
			}
		}
	})
	markdown.GET("/delete/*path", func(c *gin.Context) {
		path := c.Param("path")
		tokens := strings.Split(path[1:], "/")
		if len(tokens) != 2 || !IsValidNonce(tokens[0]) || !IsValidNonce(tokens[1]) {
			c.String(400, "invalid request")
		} else {
			if !notebook.deleteElem(tokens[0], tokens[1], true) {
				c.String(404, "file not found")
			} else {
				c.Redirect(302, webroot+"markdown/deleted")
			}
		}
	})
	markdown.POST("/save/*path", func(c *gin.Context) {
		path := c.Param("path")
		tokens := strings.Split(path[1:], "/")
		if len(tokens) != 2 || !IsValidNonce(tokens[0]) || !IsValidNonce(tokens[1]) {
			c.String(400, "invalid request")
		} else {
			markdown := c.DefaultPostForm("markdown", "")
			if !notebook.save([]byte(markdown), tokens[0], tokens[1]+".md") {
				c.String(500, "can't save markdown.")
			} else {
				c.Redirect(302, webroot+"markdown/view/"+tokens[0]+"/"+tokens[1])
			}
		}
	})
	markdown.GET("/new/:unique", func(c *gin.Context) {
		unique := c.Param("unique")
		page := notebook.get(unique)
		if page == nil {
			c.HTML(404, "error.tmpl", gin.H{
				"root":  webroot,
				"error": "cannot find a notebook page",
			})
		} else {
			if eunique := notebook.newmd(unique); len(eunique) > 0 {
				c.Redirect(302, webroot+"markdown/edit/"+unique+"/"+eunique)
			} else {
				c.HTML(404, "error.tmpl", gin.H{
					"root":  webroot,
					"error": "cannot add element to the page",
				})
			}
		}
	})
}
