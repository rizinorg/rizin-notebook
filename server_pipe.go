package main

import (
	"github.com/gin-gonic/gin"
)

func serverAddPipe(pipe *gin.RouterGroup) {
	pipe.GET("/open/:unique", func(c *gin.Context) {
		unique := c.Param("unique")
		if !IsValidNonce(unique) {
			c.HTML(404, "error.tmpl", gin.H{
				"root":  webroot,
				"error": "unknown page.",
			})
			return
		}
		rizin := notebook.open(unique, true)
		if rizin == nil {
			c.HTML(404, "error.tmpl", gin.H{
				"root":  webroot,
				"error": "failed to open the rizin pipe.",
			})
		} else {
			c.Redirect(302, webroot+"view/"+unique)
		}
	})
	pipe.GET("/close/:unique", func(c *gin.Context) {
		unique := c.Param("unique")
		if !IsValidNonce(unique) {
			c.HTML(404, "error.tmpl", gin.H{
				"root":  webroot,
				"error": "unknown page.",
			})
			return
		} else {
			notebook.close(unique)
			c.Redirect(302, webroot+"view/"+unique)
		}
	})
}
