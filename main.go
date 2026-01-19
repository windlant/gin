package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/windlant/gin/handlers"
	"github.com/windlant/gin/middleware"
)

func main() {

	if err := os.MkdirAll("logs", 0755); err != nil {
		log.Fatal("Failed to create logs directory:", err)
	}

	logFile, err := os.OpenFile("logs/gin-access.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}
	defer logFile.Close()

	r := gin.Default()

	// 注册全局中间件
	// r.Use(middleware.LoggerWithWriter(logFile))
	//标准输出
	r.Use(middleware.LoggerWithWriter(os.Stdout))

	// 路由
	r.GET("/users", handlers.GetUsers)
	r.GET("/users/:id", handlers.GetUser)
	r.POST("/users", handlers.CreateUsers)
	r.PUT("/users", handlers.UpdateUsers)
	r.DELETE("/users", handlers.DeleteUsers)

	r.Run(":8080")
}
