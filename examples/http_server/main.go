package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"fmt"
)

func index(c *gin.Context) {
	data := gin.H{"Title": "My Gin Web Page", "Heading": "Ohhh yes!", "Message": "This page is rendered using Gin with template variables!",}
	c.HTML(http.StatusOK, "index.html", data)
}


func main() {
	app := gin.Default()
	app.LoadHTMLGlob("../templates/*")
	app.GET("/", index)
	fmt.Println("http://localhost:8080")
	app.Run(":8080")
}
