package main

import (
	"embed"
	"encoding/json"
	"html/template"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/render"
)

//go:embed assets/templates/*
var embedTemplates embed.FS

//go:embed assets/static/*
var embedStatic embed.FS

// Template helper functions: stringify, keybindings.
var functionMap = template.FuncMap{
	"stringify": func(v interface{}) string {
		bytes, err := json.Marshal(v)
		if err != nil {
			return "{}"
		}
		return string(bytes)
	},
	"keybindings": func() template.JS {
		return getKeybindingsJSON()
	},
	"autocomplete_config": func() template.JS {
		return getAutocompleteConfigJSON()
	},
	"search_config": func() template.JS {
		return getNotebookSearchConfigJSON()
	},
}

/* ─── Custom template render ─────────────────────────────── */

type customHTMLRender struct {
	Template *template.Template
}

func (r *customHTMLRender) Instance(name string, data interface{}) render.Render {
	return &customHTMLRenderInstance{
		Template: r.Template,
		Name:     name,
		Data:     data,
	}
}

type customHTMLRenderInstance struct {
	Template *template.Template
	Name     string
	Data     interface{}
}

func (r *customHTMLRenderInstance) Render(w http.ResponseWriter) error {
	r.WriteContentType(w)
	return r.Template.ExecuteTemplate(w, r.Name, r.Data)
}

func (r *customHTMLRenderInstance) WriteContentType(w http.ResponseWriter) {
	header := w.Header()
	if val := header["Content-Type"]; len(val) == 0 {
		header["Content-Type"] = []string{"text/html; charset=utf-8"}
	}
}

/* ─── Template loading ───────────────────────────────────── */

func setupTemplate(assets string, router *gin.Engine) (http.FileSystem, *template.Template) {
	if len(assets) > 0 {
		return setupTemplateDebug(assets, router)
	}
	return setupTemplateEmbed(router)
}

func setupTemplateEmbed(router *gin.Engine) (http.FileSystem, *template.Template) {
	tmpl := template.New("").Funcs(functionMap)

	err := fs.WalkDir(embedTemplates, "assets/templates", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".tmpl") {
			return nil
		}
		data, readErr := embedTemplates.ReadFile(path)
		if readErr != nil {
			return readErr
		}
		name := filepath.Base(path)
		_, parseErr := tmpl.New(name).Parse(string(data))
		return parseErr
	})
	if err != nil {
		panic("failed to load embedded templates: " + err.Error())
	}

	router.HTMLRender = &customHTMLRender{Template: tmpl}

	staticFS, fsErr := fs.Sub(embedStatic, "assets/static")
	if fsErr != nil {
		panic("failed to load embedded static assets: " + fsErr.Error())
	}
	return http.FS(staticFS), tmpl
}

func setupTemplateDebug(assets string, router *gin.Engine) (http.FileSystem, *template.Template) {
	tmpl := template.New("").Funcs(functionMap)
	templateDir := filepath.Join(assets, "templates")

	err := filepath.Walk(templateDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || !strings.HasSuffix(path, ".tmpl") {
			return nil
		}
		data, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}
		name := filepath.Base(path)
		_, parseErr := tmpl.New(name).Parse(string(data))
		return parseErr
	})
	if err != nil {
		panic("failed to load debug templates: " + err.Error())
	}

	router.HTMLRender = &customHTMLRender{Template: tmpl}

	staticDir := filepath.Join(assets, "static")
	return http.Dir(staticDir), tmpl
}

/* ─── Static file routes ─────────────────────────────────── */

func serverAddAssets(root *gin.RouterGroup, assets string, static http.FileSystem, templates *template.Template) {
	root.StaticFS("/static", static)
	root.GET("/favicon.ico", func(c *gin.Context) {
		c.Writer.WriteHeader(http.StatusNoContent)
		_, _ = io.WriteString(c.Writer, "")
	})

	root.GET("/reload", func(c *gin.Context) {
		if len(assets) > 0 {
			templateDir := filepath.Join(assets, "templates")
			newTmpl := template.New("").Funcs(functionMap)
			_ = filepath.Walk(templateDir, func(path string, info os.FileInfo, err error) error {
				if err != nil || info.IsDir() || !strings.HasSuffix(path, ".tmpl") {
					return err
				}
				data, readErr := os.ReadFile(path)
				if readErr != nil {
					return readErr
				}
				name := filepath.Base(path)
				_, parseErr := newTmpl.New(name).Parse(string(data))
				return parseErr
			})
			templates = newTmpl
		}
		c.HTML(200, "reload.tmpl", gin.H{
			"root": webroot,
		})
	})
}
