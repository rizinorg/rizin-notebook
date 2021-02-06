package main

import (
	"github.com/gin-gonic/gin"
)

func serverAddAbout(root *gin.RouterGroup) {
	root.GET("/about", func(c *gin.Context) {
		var rzversion, rzbuild string
		info, err := notebook.info()
		rzpath := notebook.rizin

		if err != nil {
			rzversion = err.Error()
			rzbuild = rzversion
			if rzpath == "rizin" {
				rzpath = "Not found in $PATH"
			} else {
				rzpath = "Not found in " + rzpath
			}
		} else {
			rzversion = info[0]
			rzbuild = info[1]
			if rzpath == "rizin" {
				rzpath = "Available from $PATH"
			}
		}

		c.HTML(200, "about.tmpl", gin.H{
			"root":      webroot,
			"rzversion": rzversion,
			"rzbuild":   rzbuild,
			"rzpath":    rzpath,
			"nbversion": NBVERSION,
			"storage":   notebook.storage,
		})
	})
}
