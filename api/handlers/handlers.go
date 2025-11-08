package handlers

import (
	"GoGin/dao"
	"GoGin/services"
	"GoGin/util"

	"github.com/gin-gonic/gin"

	"net/http"
)

// RegisterRequest 注册请求
type RegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Email    string `json:"email" binding:"required"`
}

// LoginRequest 登录请求
type LoginRequest struct {
	LoginKey string `json:"loginKey" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func Register(c *gin.Context) {
	//并发安全
	dao.DataSync.Lock()
	defer dao.DataSync.Unlock()

	//捕获数据
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.Error(c, 400, err.Error())
		return
	}

	//调用服务层
	user, err := services.Register(req.Username, req.Password, req.Email)
	if err != nil {
		util.Error(c, 500, err.Error())
		return
	}

	//返回响应
	util.Success(c, gin.H{
		"username": user.Username,
		"user_id":  user.UserID,
		"email":    user.Email,
	}, "RegisterRequest registered successfully")
}

func Login(c *gin.Context) {
	//并发安全
	dao.DataSync.Lock()
	defer dao.DataSync.Unlock()

	//捕获数据
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": http.StatusBadRequest,
			"msg":    "Information is incomplete",
		})
		return
	}

	//调用服务层
	token, user, err := services.Login(req.LoginKey, req.Password)
	if err != nil {
		util.Error(c, 500, err.Error())
	}

	//返回响应
	util.Success(c, gin.H{
		"username": user.Username,
		"user_id":  user.UserID,
		"email":    user.Email,
		"token":    token,
	}, "login successful")
}

func InfoHandler(c *gin.Context) {
	//捕获数据
	userID, _ := c.Get("user_id")
	username, _ := c.Get("username")

	//返回响应
	util.Success(c, gin.H{
		"user_id":  userID,
		"username": username,
	}, "Protected resource")
}
