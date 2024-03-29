package main

import (
	"github.com/gin-gonic/gin"
	"strings"
)

func serverAddSettings(root *gin.RouterGroup) {
	root.GET("/settings", func(c *gin.Context) {
		c.HTML(200, "settings.tmpl", gin.H{
			"root":        webroot,
			"environment": config.Environment,
		})
	})
	root.GET("/settings/:action/:section/:editkey", func(c *gin.Context) {
		action := c.Param("action")
		section := c.Param("section")
		editkey := c.Param("editkey")
		if action != "environment" {
			c.HTML(404, "error.tmpl", gin.H{
				"root":  webroot,
				"error": "cannot find page",
			})
			return
		}
		if editkey == "new" {
			editkey = ""
		}
		var data map[string]string = nil
		if action == "environment" {
			data = config.Environment
		}
		c.HTML(200, "settings-edit.tmpl", gin.H{
			"root":    webroot,
			"action":  action,
			"section": section,
			"editkey": editkey,
			"data":    data,
		})
	})
	root.POST("/settings", func(c *gin.Context) {
		key := c.DefaultPostForm("key", "")
		value := c.DefaultPostForm("value", "")
		action := c.DefaultPostForm("action", "")

		if action != "environment" {
			c.HTML(404, "error.tmpl", gin.H{
				"root":     webroot,
				"location": webroot + "settings",
				"error":    "invalid settings (Action)",
			})
			return
		} else if action == "environment" {
			editkey := c.DefaultPostForm("editkey", "")
			subaction := c.DefaultPostForm("subaction", "")
			editkey = strings.TrimSpace(editkey)
			if editkey == "" && subaction != "new" {
				c.HTML(404, "error.tmpl", gin.H{
					"root":     webroot,
					"location": webroot + "settings",
					"error":    "invalid settings (Environment Variable)",
				})
			} else if subaction == "new" {
				config.SetEnvironment(key, value)
			} else if subaction == "delete" {
				config.DelEnvironment(editkey)
			} else {
				if editkey != key {
					config.DelEnvironment(editkey)
				}
				config.SetEnvironment(key, value)
			}
		}
		config.Save()
		c.Redirect(302, webroot+"settings")
	})
}
