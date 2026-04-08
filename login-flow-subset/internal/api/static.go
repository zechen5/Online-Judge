package api

import (
	"mime"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

func registerStaticRoutes(router *gin.Engine, webRoot string) {
	router.GET("/", func(c *gin.Context) {
		serveStaticFile(c, filepath.Join(webRoot, "index.html"))
	})

	router.GET("/assets/:name", func(c *gin.Context) {
		name := path.Clean(c.Param("name"))
		if name == "." || strings.Contains(name, "/") {
			c.Status(http.StatusNotFound)
			return
		}
		serveStaticFile(c, filepath.Join(webRoot, name))
	})
}

func serveStaticFile(c *gin.Context, absolutePath string) {
	data, err := os.ReadFile(absolutePath)
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}

	contentType := mime.TypeByExtension(filepath.Ext(absolutePath))
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	c.Data(http.StatusOK, contentType, data)
}
