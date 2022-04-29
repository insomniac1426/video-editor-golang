package main

import (
	"github.com/gin-gonic/gin"
	"github.com/insomniac1426/video-editor/controllers"
)

func main() {
	r := gin.Default()

	r.PUT("/upload-video", controllers.PutUser)
	r.Run(":9090")
}