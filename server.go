package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"gopkg.in/russross/blackfriday.v2"
	"html/template"
	"path"
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
	root.GET("/", func(c *gin.Context) {
		c.HTML(200, "index.tmpl", gin.H{
			"root": webRoot,
			"list": notebook.list(),
		})
	})
	root.GET("/new", func(c *gin.Context) {
		nonce := notebook.new()
		if len(nonce) > 0 {
			c.Redirect(302, "/open/"+nonce)
		} else {
			c.HTML(500, "error.tmpl", gin.H{
				"root":  webRoot,
				"error": "cannot create a new page",
			})
		}
	})
	root.GET("/open/:nonce", func(c *gin.Context) {
		nonce := c.Param("nonce")
		page := notebook.get(nonce)
		if page == nil {
			c.HTML(404, "error.tmpl", gin.H{
				"root":  webRoot,
				"error": "cannot find a new page",
			})
		} else {
			c.HTML(200, "page.tmpl", gin.H{
				"root": webRoot,
				"page": page,
			})
		}
	})
	root.GET("/markdown/*path", func(c *gin.Context) {
		tokens := strings.Split(c.Param("path")[1:], "/")
		if len(tokens) != 2 || !IsValidNonce(tokens[0]) || !IsValidNonce(tokens[1]) {
			c.String(400, "invalid request")
		} else {
			bytes, err := notebook.file(tokens[0], tokens[1]+".md")
			if err != nil {
				c.String(404, "file not found")
			} else {
				html := blackfriday.Run(bytes)
				c.HTML(200, "markdown.tmpl", gin.H{
					"root": webRoot,
					"html": html,
				})
			}
		}
	})
	root.GET("/output/*path", func(c *gin.Context) {
		tokens := strings.Split(c.Param("path")[1:], "/")
		if len(tokens) != 2 || !IsValidNonce(tokens[0]) || !IsValidNonce(tokens[1]) {
			c.String(400, "invalid request")
		} else {
			bytes, err := notebook.file(tokens[0], tokens[1]+".output")
			if err != nil {
				c.String(404, "file not found")
			} else {
				c.String(200, string(bytes))
			}
		}
	})
	fmt.Printf("Server listening at http://%s\n", bind)
	router.Run(bind)
}
