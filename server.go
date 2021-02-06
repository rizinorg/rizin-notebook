package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
)

func sanitizeWebRoot(path string) string {
	if len(path) < 2 {
		return "/"
	} else if len(path) > 1 && path[len(path)-1] != '/' {
		return path + "/"
	}
	return path
}

func runServer(assets, bind string, debug bool) {
	var root *gin.RouterGroup

	gin.DisableConsoleColor()
	if debug {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	static, templates := setupTemplate(assets, router)

	root = router.Group(sanitizeWebRoot(webroot))

	serverAddAssets(root, assets, static, templates)

	root.GET("/", func(c *gin.Context) {
		c.HTML(200, "index.tmpl", gin.H{
			"root": webroot,
			"list": notebook.list(),
		})
	})

	serverAddAbout(root)

	serverAddSettings(root)

	serverAddPage(root)

	pipe := root.Group("/pipe")
	serverAddPipe(pipe)

	markdown := root.Group("/markdown")
	serverAddMarkdown(markdown)

	output := root.Group("/output")
	serverAddOutput(output)

	fmt.Printf("Server listening at http://%s\n", bind)
	router.Run(bind)
}
