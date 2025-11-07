package dao

import (
	"GoGin/model"
	"sync"
)

var (
	database    = make(map[string]model.User) // 以username存储用户信息
	emailToData = make(map[string]string)     // 存储email对应username
	ID          = 1
	DataSync    sync.Mutex // 并发安全
)

func AddUser(username, password, email string, userID string) { // 添加数据
	database[username] = model.User{
		Username: username,
		Password: password,
		Email:    email,
		UserID:   userID,
	}
	emailToData[email] = username
}

func CheckUsername(username string) bool { // 注册时：检查用户名是否已存在
	_, ok := database[username]
	return ok
}

func CheckEmail(email string) bool { // 注册时：检查邮箱是否已被注册
	_, ok := emailToData[email]
	return ok
}

// SelectPassword 登录时：返回正确密码供检查
func SelectPassword(username string) string {

	return database[username].Password
}

func EmailToUsername(email string) string { // 由邮箱查询用户名
	return emailToData[email]
}

// CheckID 查询userID
func CheckID(username string) string {
	return database[username].UserID
}
