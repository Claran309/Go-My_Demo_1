package routes

import (
	"GoGin/api/handlers"

	"GoGin/middleware"

	"github.com/gin-gonic/gin"
)

func InitRouter() {
	r := gin.Default()

	//注册和登录路由
	r.POST("/register", handlers.Register)
	r.POST("/login", handlers.Login)

	//受保护路由：JWT中间件判断访问权限
	protected := r.Group("/protected")
	protected.Use(middleware.JWTAuthorizationMiddleware())
	{
		protected.GET("/info", handlers.InfoHandler)
	}

	err := r.Run()
	if err != nil {
		panic("Failed to start Gin server: " + err.Error())
	}
}

//http://localhost:8080/
