package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"html/template"
	"io/fs"
	"net/http"
	"path"
	"strings"
)

//go:embed assets/*
var Assets embed.FS

var functionMap = template.FuncMap{
	"raw": func(b []byte) template.HTML {
		return template.HTML(b)
	},
	"stringify": func(input interface{}) string {
		buffer, err := json.Marshal(input)
		if err != nil {
			return `""`
		}
		return string(buffer)
	},
	"keycombo": func(input string) string {
		if len(input) > 0 {
			return strings.Replace(input, ",", " + ", -1)
		}
		return "Not Assigned"
	},
}

func loadEmbedded() (*template.Template, error) {
	t := template.New("")
	fs.WalkDir(Assets, ".", func(filePath string, file fs.DirEntry, err error) error {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".tmpl") {
			return nil
		}

		h, _ := Assets.ReadFile(filePath)
		t, err = t.New(file.Name()).Funcs(functionMap).Parse(string(h))
		return err
	})
	return t, nil
}

func loadAsset(file string) ([]byte, error) {
	file = path.Base(file)
	return Assets.ReadFile("assets/static/" + file)
}

func setupTemplate(assets string, router *gin.Engine) (string, string) {
	var static, templates string
	if len(assets) > 0 {
		fmt.Println("debugging assets.")
		static = path.Join(assets, "static")
		templates = path.Join(assets, "templates", "*")
		router.SetFuncMap(functionMap)
		router.LoadHTMLGlob(templates)
	} else {
		templates, err := loadEmbedded()
		if err != nil {
			panic(err)
		}
		router.SetHTMLTemplate(templates)
	}
	return static, templates
}

func serverAddAssets(root *gin.RouterGroup, assets, static, templates string) {
	if len(assets) > 0 {
		root.GET("/favicon.ico", func(c *gin.Context) {
			c.Redirect(302, "/static/favicon.ico")
		})
		root.Static("/static", static)
	} else {
		root.GET("/favicon.ico", func(c *gin.Context) {
			content, err := loadAsset("favicon.ico")
			if content == nil && err == nil {
				c.Status(404)
				return
			} else if err != nil {
				c.Status(500)
				fmt.Println("[Assets]", err)
				return
			}
			c.Data(200, "image/x-icon", content)
		})
		root.GET("/static/:file", func(c *gin.Context) {
			file := c.Param("file")
			content, err := loadAsset(file)
			if content == nil && err == nil {
				c.Status(404)
				return
			} else if err != nil {
				c.Status(500)
				fmt.Println("[Assets]", err)
				return
			}
			var contentType = "text/plain"
			if strings.HasSuffix(file, ".css") {
				contentType = "text/css"
			} else if strings.HasSuffix(file, ".js") {
				contentType = "text/javascript"
			} else {
				contentType = http.DetectContentType(content)
			}
			c.Data(200, contentType, content)
		})
	}
}
