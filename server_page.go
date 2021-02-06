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
			"root":  webroot,
			"title": "",
			"nonce": "",
		})
	})
	root.POST("/edit", func(c *gin.Context) {
		nonce := c.DefaultPostForm("nonce", "")
		title := c.DefaultPostForm("title", "")
		if len(nonce) > 0 {
			valid := IsValidNonce(nonce)
			if valid && notebook.rename(nonce, title) {
				c.Redirect(302, webroot+"view/"+nonce)
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
			nonce := notebook.new(title, filename, binary)
			if len(nonce) < 1 {
				c.HTML(500, "error.tmpl", gin.H{
					"root":  webroot,
					"error": "cannot create page",
				})
				return
			}

			if c.SaveUploadedFile(file, path.Join(notebook.storage, nonce, binary)) != nil {
				notebook.delete(nonce)
				c.HTML(500, "error.tmpl", gin.H{
					"root":  webroot,
					"error": "cannot save binary",
				})
			} else {
				c.Redirect(302, webroot+"view/"+nonce)
			}
		}
	})
	root.GET("/delete/:nonce", func(c *gin.Context) {
		nonce := c.Param("nonce")
		if !IsValidNonce(nonce) || !notebook.delete(nonce) {
			c.HTML(404, "error.tmpl", gin.H{
				"root":  webroot,
				"error": "cannot find page",
			})
		} else {
			c.Redirect(302, webroot)
		}
	})
	root.GET("/edit/:nonce", func(c *gin.Context) {
		nonce := c.Param("nonce")
		page := notebook.get(nonce)
		if page == nil {
			c.HTML(404, "error.tmpl", gin.H{
				"root":  webroot,
				"error": "cannot find page",
			})
		} else {
			c.HTML(200, "page-new.tmpl", gin.H{
				"root":  webroot,
				"title": page["title"],
				"nonce": nonce,
			})
		}
	})
	root.GET("/view/:nonce", func(c *gin.Context) {
		nonce := c.Param("nonce")
		page := notebook.get(nonce)
		if page == nil {
			c.HTML(404, "error.tmpl", gin.H{
				"root":  webroot,
				"error": "cannot find a new page",
			})
		} else {
			c.HTML(200, "page-view.tmpl", gin.H{
				"root": webroot,
				"page": page,
				"pipe": notebook.open(nonce, false) != nil,
			})
		}
	})
}
