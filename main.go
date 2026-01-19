package main

import (
	"github.com/gin-gonic/gin"
	"github.com/windlant/gin/handlers"
	"github.com/windlant/gin/middleware"
)

func main() {
	r := gin.Default()

	// 注册全局中间件
	r.Use(middleware.Logger())

	// 路由
	r.GET("/users", handlers.GetUsers)
	r.GET("/users/:id", handlers.GetUser)
	r.POST("/users", handlers.CreateUsers)
	r.PUT("/users", handlers.UpdateUsers)
	r.DELETE("/users", handlers.DeleteUsers)

	r.Run(":8080")
}
