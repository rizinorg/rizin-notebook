package main

import (
	"github.com/gin-gonic/gin"
)

func serverAddPipe(pipe *gin.RouterGroup) {
	pipe.GET("/open/:nonce", func(c *gin.Context) {
		nonce := c.Param("nonce")
		if !IsValidNonce(nonce) {
			c.HTML(404, "error.tmpl", gin.H{
				"root":  webroot,
				"error": "unknown page.",
			})
			return
		}
		rizin := notebook.open(nonce, true)
		if rizin == nil {
			c.HTML(404, "error.tmpl", gin.H{
				"root":  webroot,
				"error": "failed to open the rizin pipe.",
			})
		} else {
			c.Redirect(302, webroot+"view/"+nonce)
		}
	})
	pipe.GET("/close/:nonce", func(c *gin.Context) {
		nonce := c.Param("nonce")
		if !IsValidNonce(nonce) {
			c.HTML(404, "error.tmpl", gin.H{
				"root":  webroot,
				"error": "unknown page.",
			})
			return
		} else {
			notebook.close(nonce)
			c.Redirect(302, webroot+"view/"+nonce)
		}
	})
}
