package handlers

import (
	"GoGin/dao"
	"GoGin/util"
	"strconv"

	"github.com/gin-gonic/gin"

	"net/http"
)

func Register(c *gin.Context) {
	//并发安全
	dao.DataSync.Lock()
	defer dao.DataSync.Unlock()

	//传入用户信息
	type User struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
		Email    string `json:"email" binding:"required"`
	}
	var user User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{ //信息不完整
			"status": http.StatusBadRequest,
			"msg":    "Information is incomplete",
		})
		return
	}

	//判断用户信息是否合法：
	//1. 用户名是否被使用过
	if flag := dao.CheckUsername(user.Username); flag {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": http.StatusInternalServerError,
			"msg":    "user already exists",
		})
		return
	}

	//2. 密码时候否符合格式（仅包含英文字母和数字）
	var flagPassword bool
	for i := 0; i < len(user.Password); i++ {
		if !((user.Password[i] >= 'a' && user.Password[i] <= 'z') || (user.Password[i] >= '0' && user.Password[i] <= '9') || (user.Password[i] >= 'A' && user.Password[i] <= 'Z')) {
			flagPassword = true
		}
	}
	if flagPassword {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": http.StatusInternalServerError,
			"msg":    "Password format is incorrect",
		})
		return
	}

	//3. 邮箱是否被注册过
	if flag := dao.CheckEmail(user.Email); flag {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": http.StatusInternalServerError,
			"msg":    "Email already been registered",
		})
		return
	}

	// 验证成功，填入UserID
	userID := strconv.Itoa(dao.ID)
	dao.ID++
	//传出用户信息到数据库
	dao.AddUser(user.Username, user.Password, user.Email, userID)

	//JSON返回成功信息
	c.JSON(http.StatusOK, gin.H{
		"status": http.StatusOK,
		"msg":    "User registered successfully",
	})
}

func Login(c *gin.Context) {
	//并发安全
	dao.DataSync.Lock()
	defer dao.DataSync.Unlock()

	//导入数据
	type LoginInfo struct {
		LoginKey string `json:"loginKey" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	var info LoginInfo
	if err := c.ShouldBindJSON(&info); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": http.StatusBadRequest,
			"msg":    "Information is incomplete",
		})
		return
	}

	//判断是邮箱登录还是用户名登录
	var username string
	var at, point bool
	for i := 0; i < len(info.LoginKey); i++ {
		if info.LoginKey[i] == '@' {
			at = true
		}
		if info.LoginKey[i] == '.' {
			point = true
		}
	}
	if at && point { // 邮箱登录
		username = dao.EmailToUsername(info.LoginKey)
	} else { // 用户名登录
		username = info.LoginKey
	}

	//检查用户是否存在
	if flag := dao.CheckUsername(username); !flag {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": http.StatusInternalServerError,
			"msg":    "User does not exist",
		})
		return
	}

	//检验密码正确性
	if info.Password != dao.SelectPassword(username) {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": http.StatusInternalServerError,
			"msg":    "Incorrect password",
		})
		return
	}

	//返回token
	token, err := util.GenerateToken(dao.CheckID(username), username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": http.StatusInternalServerError,
			"msg":    "Token generation failed",
		})
		return
	}

	//JSON返回成功信息
	c.JSON(http.StatusOK, gin.H{
		"status": http.StatusOK,
		"token":  token,
		"msg":    "login successful",
	})
}

func InfoHandler(c *gin.Context) {
	userID, _ := c.Get("user_id")
	username, _ := c.Get("username")

	c.JSON(http.StatusOK, gin.H{
		"status":   http.StatusOK,
		"user_id":  userID,
		"username": username,
		"msg":      "Protected resource",
	})
}
