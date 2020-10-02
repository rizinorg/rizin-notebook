package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"gopkg.in/russross/blackfriday.v2"
	"html/template"
	"path"
	"path/filepath"
	"strings"
)

func sanitizeWebRoot(path string) string {
	if len(path) < 2 {
		return "/"
	} else if len(path) > 1 && path[len(path)-1] != '/' {
		return path + "/"
	}
	return path
}

func server(webRoot, assets, bind string) {
	var root *gin.RouterGroup

	static := path.Join(assets, "static")
	templates := path.Join(assets, "templates", "*")

	router := gin.Default()
	router.SetFuncMap(template.FuncMap{
		"raw": func(b []byte) template.HTML {
			return template.HTML(b)
		},
	})
	router.LoadHTMLGlob(templates)

	root = router.Group(sanitizeWebRoot(webRoot))
	root.GET("/favicon.ico", func(c *gin.Context) {
		c.Redirect(302, "/static/favicon.ico")
	})
	root.Static("/static", static)

	{
		root.GET("/", func(c *gin.Context) {
			c.HTML(200, "index.tmpl", gin.H{
				"root": webRoot,
				"list": notebook.list(),
			})
		})
		root.GET("/new", func(c *gin.Context) {
			c.HTML(200, "page-new.tmpl", gin.H{
				"root":  webRoot,
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
					c.Redirect(302, webRoot+"view/"+nonce)
				} else {
					c.HTML(500, "error.tmpl", gin.H{
						"root":  webRoot,
						"error": "cannot rename page",
					})
				}
			} else {
				file, err := c.FormFile("binary")
				if err != nil {
					fmt.Println(err)
					c.HTML(500, "error.tmpl", gin.H{
						"root":  webRoot,
						"error": "invalid or missing binary",
					})
					return
				}

				filename := filepath.Base(file.Filename)
				if len(filename) < 1 {
					c.HTML(500, "error.tmpl", gin.H{
						"root":  webRoot,
						"error": "invalid binary name",
					})
					return
				}

				ext := filepath.Ext(filename)
				binary := Nonce(ELEMENT_NONCE_SIZE) + ext
				nonce := notebook.new(title, filename, binary)
				if len(nonce) < 1 {
					c.HTML(500, "error.tmpl", gin.H{
						"root":  webRoot,
						"error": "cannot create page",
					})
					return
				}

				if c.SaveUploadedFile(file, path.Join(notebook.storage, nonce, binary)) != nil {
					notebook.delete(nonce)
					c.HTML(500, "error.tmpl", gin.H{
						"root":  webRoot,
						"error": "cannot save binary",
					})
				} else {
					c.Redirect(302, webRoot+"view/"+nonce)
				}
			}
		})
		root.GET("/delete/:nonce", func(c *gin.Context) {
			nonce := c.Param("nonce")
			if !IsValidNonce(nonce) || !notebook.delete(nonce) {
				c.HTML(404, "error.tmpl", gin.H{
					"root":  webRoot,
					"error": "cannot find page",
				})
			} else {
				c.Redirect(302, webRoot)
			}
		})
		root.GET("/edit/:nonce", func(c *gin.Context) {
			nonce := c.Param("nonce")
			page := notebook.get(nonce)
			if page == nil {
				c.HTML(404, "error.tmpl", gin.H{
					"root":  webRoot,
					"error": "cannot find page",
				})
			} else {
				c.HTML(200, "page-new.tmpl", gin.H{
					"root":  webRoot,
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
					"root":  webRoot,
					"error": "cannot find a new page",
				})
			} else {
				c.HTML(200, "page-view.tmpl", gin.H{
					"root": webRoot,
					"page": page,
					"pipe": notebook.open(nonce, false) != nil,
				})
			}
		})
	}

	pipe := root.Group("/pipe")
	{
		pipe.GET("/open/:nonce", func(c *gin.Context) {
			nonce := c.Param("nonce")
			if !IsValidNonce(nonce) {
				c.HTML(404, "error.tmpl", gin.H{
					"root":  webRoot,
					"error": "unknown page.",
				})
				return
			}
			rizin := notebook.open(nonce, true)
			if rizin == nil {
				c.HTML(404, "error.tmpl", gin.H{
					"root":  webRoot,
					"error": "failed to open the rizin pipe.",
				})
			} else {
				c.Redirect(302, webRoot+"view/"+nonce)
			}
		})
		pipe.GET("/close/:nonce", func(c *gin.Context) {
			nonce := c.Param("nonce")
			if !IsValidNonce(nonce) {
				c.HTML(404, "error.tmpl", gin.H{
					"root":  webRoot,
					"error": "unknown page.",
				})
				return
			} else {
				notebook.close(nonce)
				c.Redirect(302, webRoot+"view/"+nonce)
			}
		})
	}

	markdown := root.Group("/markdown")
	{
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
					html := blackfriday.Run(bytes)
					c.HTML(200, "markdown-view.tmpl", gin.H{
						"root": webRoot,
						"html": html,
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
						"root": webRoot,
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
					c.Redirect(302, webRoot+"markdown/deleted")
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
					c.Redirect(302, webRoot+"markdown/view/"+tokens[0]+"/"+tokens[1])
				}
			}
		})
		markdown.GET("/new/:nonce", func(c *gin.Context) {
			nonce := c.Param("nonce")
			page := notebook.get(nonce)
			if page == nil {
				c.HTML(404, "error.tmpl", gin.H{
					"root":  webRoot,
					"error": "cannot find a notebook page",
				})
			} else {
				if enonce := notebook.newmd(nonce); len(enonce) > 0 {
					c.Redirect(302, webRoot+"markdown/edit/"+nonce+"/"+enonce)
				} else {
					c.HTML(404, "error.tmpl", gin.H{
						"root":  webRoot,
						"error": "cannot add element to the page",
					})
				}
			}
		})
	}

	output := root.Group("/output")
	{
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
					"root":  webRoot,
				})
				return
			}

			rizin := notebook.open(nonce, false)
			if rizin == nil {
				c.HTML(400, "console-error.tmpl", gin.H{
					"error": "pipe is closed.",
					"root":  webRoot,
				})
				return
			}

			enonce := notebook.newcmd(nonce, command)
			if len(enonce) < 1 {
				c.HTML(400, "console-error.tmpl", gin.H{
					"error": "cannot create enonce for output.",
					"root":  webRoot,
				})
				return
			}

			go func(nonce, enonce, command string, rizin *Rizin) {
				output, err := rizin.exec(command)
				if len(output) < 1 && err == nil {
					output = "no output from rizin."
				} else if err != nil {
					output = fmt.Sprintf("pipe error: %v", err)
				}
				notebook.save([]byte(output), nonce, enonce+".out")
			}(nonce, enonce, command, rizin)

			c.Redirect(302, webRoot+"output/check/"+nonce+"/"+enonce)
		})
		output.GET("/check/*path", func(c *gin.Context) {
			tokens := strings.Split(c.Param("path")[1:], "/")
			if len(tokens) != 2 || !IsValidNonce(tokens[0]) || !IsValidNonce(tokens[1]) {
				c.String(400, "invalid request")
			} else {
				_, err := notebook.file(tokens[0], tokens[1]+".out")
				if err != nil {
					c.HTML(200, "reload.tmpl", gin.H{
						"root": webRoot,
					})
				} else {
					c.Redirect(302, webRoot+"output/loaded")
				}
			}
		})
		output.GET("/input/:nonce", func(c *gin.Context) {
			nonce := c.Param("nonce")
			if !IsValidNonce(nonce) {
				c.String(400, "invalid request")
			} else {
				c.HTML(200, "console.tmpl", gin.H{
					"nonce": nonce,
					"root":  webRoot,
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
				c.String(400, "invalid request")
			} else {
				if !notebook.deleteElem(tokens[0], tokens[1], false) {
					c.String(404, "file not found")
				} else {
					c.Redirect(302, webRoot+"view/"+tokens[0])
				}
			}
		})
	}
	fmt.Printf("Server listening at http://%s\n", bind)
	router.Run(bind)
}
