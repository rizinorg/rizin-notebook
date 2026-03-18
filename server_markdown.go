package main

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func serverAddMarkdown(markdown *gin.RouterGroup) {

	// Create a new markdown cell.
	markdown.GET("/new/:unique", func(c *gin.Context) {
		unique := c.Param("unique")
		if !IsValidNonce(unique, PageNonceSize) {
			c.HTML(http.StatusBadRequest, "error.tmpl", gin.H{
				"root":  webroot,
				"error": "invalid page identifier",
			})
			return
		}

		eunique := Nonce(ElementNonceSize)
		if err := catalog.AddCell(unique, eunique, "markdown", ""); err != nil {
			c.HTML(http.StatusNotFound, "error.tmpl", gin.H{
				"root":  webroot,
				"error": "failed to create markdown cell",
			})
			return
		}

		c.Redirect(http.StatusFound, webroot+"markdown/edit/"+unique+"/"+eunique)
	})

	// View a rendered markdown cell.
	markdown.GET("/view/*path", func(c *gin.Context) {
		unique, eunique, ok := parseElementPath(c.Param("path"))
		if !ok {
			c.HTML(http.StatusBadRequest, "error.tmpl", gin.H{
				"root":  webroot,
				"error": "invalid path",
			})
			return
		}

		cell, _ := catalog.GetCell(unique, eunique)
		var content string
		if cell != nil {
			content = cell.Content
		}

		c.HTML(200, "markdown-view.tmpl", gin.H{
			"root": webroot,
			"path": "/" + unique + "/" + eunique,
			"raw":  content,
		})
	})

	// Edit a markdown cell.
	markdown.GET("/edit/*path", func(c *gin.Context) {
		unique, eunique, ok := parseElementPath(c.Param("path"))
		if !ok {
			c.HTML(http.StatusBadRequest, "error.tmpl", gin.H{
				"root":  webroot,
				"error": "invalid path",
			})
			return
		}

		cell, _ := catalog.GetCell(unique, eunique)
		var content []byte
		if cell != nil {
			content = []byte(cell.Content)
		}

		c.HTML(200, "markdown-edit.tmpl", gin.H{
			"root": webroot,
			"path": "/" + unique + "/" + eunique,
			"raw":  content,
		})
	})

	// Save markdown cell content.
	markdown.POST("/save/*path", func(c *gin.Context) {
		unique, eunique, ok := parseElementPath(c.Param("path"))
		if !ok {
			c.HTML(http.StatusBadRequest, "error.tmpl", gin.H{
				"root":  webroot,
				"error": "invalid path",
			})
			return
		}

		content := c.PostForm("markdown")
		catalog.UpdateCellContent(unique, eunique, content)

		c.Redirect(http.StatusFound, webroot+"markdown/view/"+unique+"/"+eunique)
	})

	// Delete a markdown cell.
	markdown.GET("/delete/*path", func(c *gin.Context) {
		unique, eunique, ok := parseElementPath(c.Param("path"))
		if !ok {
			c.HTML(http.StatusBadRequest, "error.tmpl", gin.H{
				"root":  webroot,
				"error": "invalid path",
			})
			return
		}

		catalog.DeleteCell(unique, eunique)

		// Redirect to "deleted" pseudo-page so iframe can detect removal.
		c.Redirect(http.StatusFound, webroot+"markdown/deleted")
	})

	// Pseudo-page for iframe deletion detection.
	markdown.GET("/deleted", func(c *gin.Context) {
		c.String(200, "deleted")
	})
}

/* ─── Helpers ─────────────────────────────────────────────── */

func parseElementPath(path string) (string, string, bool) {
	parts := splitPath(path)
	if len(parts) != 2 {
		return "", "", false
	}
	unique := parts[0]
	eunique := parts[1]
	if !IsValidNonce(unique, PageNonceSize) || !IsValidNonce(eunique, ElementNonceSize) {
		return "", "", false
	}
	return unique, eunique, true
}

func splitPath(path string) []string {
	raw := strings.Split(path, "/")
	var parts []string
	for _, p := range raw {
		if p != "" {
			parts = append(parts, p)
		}
	}
	return parts
}
