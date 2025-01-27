package main

import (
	"github.com/3ssalunke/videoverse/db"
	"github.com/gin-gonic/gin"
)

func main() {
	db.Init()

	r := gin.Default()

	r.Run(":8080")
}
