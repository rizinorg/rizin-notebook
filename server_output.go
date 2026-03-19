package main

import (
	"fmt"
	"html/template"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func formatExecutionDuration(duration time.Duration) string {
	if duration < time.Millisecond {
		return duration.Round(time.Microsecond).String()
	}
	if duration < time.Second {
		return duration.Round(time.Millisecond).String()
	}
	return duration.Round(time.Millisecond).String()
}

func formatActionTime(unixTime int64) string {
	if unixTime <= 0 {
		return ""
	}
	return time.Unix(unixTime, 0).Local().Format("15:04")
}

func serverAddOutput(output *gin.RouterGroup) {

	// View a rendered output cell.
	output.GET("/view/*path", func(c *gin.Context) {
		unique, eunique, ok := parseElementPath(c.Param("path"))
		if !ok {
			c.HTML(http.StatusBadRequest, "error.tmpl", gin.H{
				"root":  webroot,
				"error": "invalid path",
			})
			return
		}

		cell, _ := catalog.GetCell(unique, eunique)
		var data []byte
		var lines, words int
		var actionTime string
		var visibleOutput string
		if cell != nil {
			data = cell.Output
			visibleOutput = outputVisibleText(string(data))
			if len(visibleOutput) > 0 {
				lines = strings.Count(visibleOutput, "\n") + 1
				words = len(strings.Fields(visibleOutput))
			}
			if cell.Executed > 0 {
				actionTime = formatActionTime(cell.Executed)
			} else {
				actionTime = formatActionTime(cell.Created)
			}
		}

		htmlOutput := toHtml(string(data))

		c.HTML(200, "output.tmpl", gin.H{
			"root":       webroot,
			"output":     template.HTML(string(htmlOutput)),
			"raw_output": visibleOutput,
			"lines":      lines,
			"words":      words,
			"actionTime": actionTime,
		})
	})

	// Delete an output cell.
	output.GET("/delete/*path", func(c *gin.Context) {
		unique, eunique, ok := parseElementPath(c.Param("path"))
		if !ok {
			c.HTML(http.StatusBadRequest, "error.tmpl", gin.H{
				"root":  webroot,
				"error": "invalid path",
			})
			return
		}

		catalog.DeleteCell(unique, eunique)

		c.Redirect(http.StatusFound, webroot+"output/deleted")
	})

	// Command line input form (rendered in an iframe).
	output.GET("/input/console/:unique", func(c *gin.Context) {
		unique := c.Param("unique")
		if !IsValidNonce(unique, PageNonceSize) {
			c.HTML(http.StatusBadRequest, "error.tmpl", gin.H{
				"root":  webroot,
				"error": "invalid page identifier",
			})
			return
		}

		c.HTML(200, "console.tmpl", gin.H{
			"root":   webroot,
			"unique": unique,
		})
	})

	// Script editor input form (rendered in an iframe).
	output.GET("/input/script/:unique", func(c *gin.Context) {
		unique := c.Param("unique")
		if !IsValidNonce(unique, PageNonceSize) {
			c.HTML(http.StatusBadRequest, "error.tmpl", gin.H{
				"root":  webroot,
				"error": "invalid page identifier",
			})
			return
		}

		c.HTML(200, "script.tmpl", gin.H{
			"root":   webroot,
			"unique": unique,
		})
	})

	// Execute a Rizin command.
	output.POST("/exec/:unique", func(c *gin.Context) {
		unique := c.Param("unique")
		if !IsValidNonce(unique, PageNonceSize) {
			c.HTML(http.StatusBadRequest, "error.tmpl", gin.H{
				"root":  webroot,
				"error": "invalid page identifier",
			})
			return
		}

		command := c.PostForm("command")
		if command == "" {
			c.Redirect(http.StatusFound, webroot+"output/deleted")
			return
		}

		// Get or open the pipe.
		rz := notebook.open(unique, true)
		if rz == nil {
			c.HTML(http.StatusInternalServerError, "console.tmpl", gin.H{
				"root":    webroot,
				"unique":  unique,
				"command": command,
				"error":   "Failed to open Rizin pipe",
			})
			return
		}

		// Execute the command.
		started := time.Now()
		result, err := rz.exec(command)
		if err != nil {
			c.HTML(200, "console.tmpl", gin.H{
				"root":    webroot,
				"unique":  unique,
				"command": command,
				"error":   err.Error(),
			})
			return
		}

		// Create a command cell and save the output in SQLite.
		eunique := Nonce(ElementNonceSize)
		if storeErr := catalog.AddCell(unique, eunique, "command", command); storeErr != nil {
			c.HTML(http.StatusInternalServerError, "console.tmpl", gin.H{
				"root":    webroot,
				"unique":  unique,
				"command": command,
				"error":   "Failed to create command element",
			})
			return
		}

		catalog.UpdateCellOutput(unique, eunique, []byte(result))
		durationText := formatExecutionDuration(time.Since(started))
		lineCount := 0
		wordCount := 0
		if result != "" {
			lineCount = strings.Count(result, "\n") + 1
			wordCount = len(strings.Fields(result))
		}
		actionTime := time.Now().Local().Format("15:04")

		c.HTML(http.StatusOK, "console.tmpl", gin.H{
			"root":       webroot,
			"unique":     unique,
			"command":    command,
			"message":    fmt.Sprintf("Command executed in %s", durationText),
			"lineCount":  lineCount,
			"wordCount":  wordCount,
			"actionTime": actionTime,
			"cellUnique": eunique,
		})
	})

	// Execute a JavaScript script.
	output.POST("/script/:unique", func(c *gin.Context) {
		unique := c.Param("unique")
		if !IsValidNonce(unique, PageNonceSize) {
			c.HTML(http.StatusBadRequest, "error.tmpl", gin.H{
				"root":  webroot,
				"error": "invalid page identifier",
			})
			return
		}

		script := c.PostForm("script")
		if script == "" {
			c.Redirect(http.StatusFound, webroot+"output/deleted")
			return
		}

		// Get or open the pipe.
		rz := notebook.open(unique, true)

		// Create a script cell in SQLite.
		eunique := Nonce(ElementNonceSize)
		if storeErr := catalog.AddCell(unique, eunique, "script", script); storeErr != nil {
			c.HTML(http.StatusInternalServerError, "script.tmpl", gin.H{
				"root":   webroot,
				"unique": unique,
				"script": script,
				"error":  "Failed to create script element",
			})
			return
		}

		// Execute the script. rz may be nil if no pipe is open,
		// which is handled by the JavaScript engine's rizin.cmd error.
		started := time.Now()
		result, err := notebook.jsvm.exec(script, rz)

		if err != nil {
			catalog.UpdateCellOutput(unique, eunique, []byte(err.Error()))
			c.HTML(200, "script.tmpl", gin.H{
				"root":       webroot,
				"unique":     unique,
				"script":     script,
				"error":      err.Error(),
				"cellUnique": eunique,
			})
			return
		}

		catalog.UpdateCellOutput(unique, eunique, []byte(result))
		durationText := formatExecutionDuration(time.Since(started))
		lineCount := 0
		wordCount := 0
		if result != "" {
			lineCount = strings.Count(result, "\n") + 1
			wordCount = len(strings.Fields(result))
		}
		actionTime := time.Now().Local().Format("15:04")

		c.HTML(http.StatusOK, "script.tmpl", gin.H{
			"root":       webroot,
			"unique":     unique,
			"script":     script,
			"message":    fmt.Sprintf("Script executed in %s", durationText),
			"lineCount":  lineCount,
			"wordCount":  wordCount,
			"actionTime": actionTime,
			"cellUnique": eunique,
		})
	})

	// Pseudo-page for iframe deletion detection.
	// When the iframe navigates here, the parent page's output() handler
	// removes the iframe's container div from the DOM.
	output.GET("/deleted", func(c *gin.Context) {
		c.String(200, "deleted")
	})

	// Pseudo-page for parent page reload trigger.
	// When the iframe navigates here, the parent page's output() handler
	// calls location.reload() to refresh the entire page.
	output.GET("/loaded", func(c *gin.Context) {
		c.HTML(200, "reload.tmpl", gin.H{
			"root": webroot,
		})
	})
}
