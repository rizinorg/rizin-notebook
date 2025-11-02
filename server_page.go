package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"path"
	"path/filepath"
)

func serverAddPage(root *gin.RouterGroup) {
	root.GET("/new", func(c *gin.Context) {
		c.HTML(200, "page-new.tmpl", gin.H{
			"root":   webroot,
			"title":  "",
			"unique": "",
		})
	})
	root.POST("/edit", func(c *gin.Context) {
		unique := c.DefaultPostForm("unique", "")
		title := c.DefaultPostForm("title", "")
		if len(unique) > 0 {
			valid := IsValidNonce(unique)
			if valid && notebook.rename(unique, title) {
				c.Redirect(302, webroot+"view/"+unique)
			} else {
				c.HTML(500, "error.tmpl", gin.H{
					"root":  webroot,
					"error": "cannot rename page",
				})
			}
		} else {
			file, err := c.FormFile("binary")
			if err != nil {
				fmt.Println(err)
				c.HTML(500, "error.tmpl", gin.H{
					"root":  webroot,
					"error": "invalid or missing binary",
				})
				return
			}

			filename := filepath.Base(file.Filename)
			if len(filename) < 1 {
				c.HTML(500, "error.tmpl", gin.H{
					"root":  webroot,
					"error": "invalid binary name",
				})
				return
			}

			ext := filepath.Ext(filename)
			binary := Nonce(ELEMENT_NONCE_SIZE) + ext
			unique := notebook.new(title, filename, binary)
			if len(unique) < 1 {
				c.HTML(500, "error.tmpl", gin.H{
					"root":  webroot,
					"error": "cannot create page",
				})
				return
			}

			if c.SaveUploadedFile(file, path.Join(notebook.storage, unique, binary)) != nil {
				notebook.delete(unique)
				c.HTML(500, "error.tmpl", gin.H{
					"root":  webroot,
					"error": "cannot save binary",
				})
			} else {
				c.Redirect(302, webroot+"view/"+unique)
			}
		}
	})
	root.GET("/delete/:unique", func(c *gin.Context) {
		unique := c.Param("unique")
		if !IsValidNonce(unique) || !notebook.delete(unique) {
			c.HTML(404, "error.tmpl", gin.H{
				"root":  webroot,
				"error": "cannot find page",
			})
			return
		}

		c.Redirect(302, webroot)
	})
	root.GET("/edit/:unique", func(c *gin.Context) {
		unique := c.Param("unique")
		page := notebook.get(unique)
		if page == nil {
			c.HTML(404, "error.tmpl", gin.H{
				"root":  webroot,
				"error": "cannot find page",
			})
			return
		}
		c.HTML(200, "page-new.tmpl", gin.H{
			"root":   webroot,
			"title":  page["title"],
			"unique": unique,
		})
	})
	root.GET("/view/:unique", func(c *gin.Context) {
		unique := c.Param("unique")
		page := notebook.get(unique)
		if page == nil {
			c.HTML(404, "error.tmpl", gin.H{
				"root":  webroot,
				"error": "cannot find a new page",
			})
			return
		}

		c.HTML(200, "page-view.tmpl", gin.H{
			"root": webroot,
			"page": page,
			"pipe": notebook.open(unique, false) != nil,
			"cmds": notebook.cmds,
		})
	})
}
