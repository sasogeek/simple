package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
)

func index(c *gin.Context) {
	data := gin.H{
		"Title":   "My Gin Web Page",
		"Heading": "Ohhh yes!",
		"Message": "This page is rendered using Gin with template variables!",
	}
	c.HTML(http.StatusOK, "index.html", data)
}

func main() {
	app := gin.Default()
	app.LoadHTMLGlob("templates/*")
	app.GET("/", index)
	port := os.Getenv("PORT")
	if fmt.Sprintf("%v", port) == fmt.Sprintf("%v", "") {
		port = "8081"
	}
	app.Run(fmt.Sprintf("%v+%v", ":", fmt.Sprintf("%v", port)))
}
