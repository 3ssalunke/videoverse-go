package main

import (
	"log"
	"net/http"

	"github.com/3ssalunke/videoverse/controllers"
	"github.com/3ssalunke/videoverse/db"
	"github.com/3ssalunke/videoverse/repository"
	"github.com/gin-gonic/gin"
)

func main() {
	db.Init()

	port := ":8080"

	r := gin.Default()

	api := r.Group("/api")
	{
		videoV1 := api.Group("/v1/videos")
		videoRepo := repository.NewVideoRepository(db.DB)
		videoController := controllers.NewVideoController(videoRepo)

		{
			videoV1.GET("/health-check", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{
					"message": "Welcome to videoverse",
				})
			})

			videoV1.Use(controllers.AuthMiddleware())

			videoV1.POST("/upload", videoController.UploadVideo)
		}
	}

	log.Printf("server started on port %s", port)
	r.Run(port)
}
