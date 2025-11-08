package services

import (
	"GoGin/dao"
	"GoGin/model"
	"GoGin/util"
	"errors"
	"strconv"
)

func Register(username, password, email string) (*model.User, error) {
	//用户名是否被使用过
	if flag := dao.CheckUsername(username); flag {
		return nil, errors.New("username Already Exist")
	}

	//密码时候否符合格式（仅包含英文字母和数字）
	var flagPassword bool
	for i := 0; i < len(password); i++ {
		if !((password[i] >= 'a' && password[i] <= 'z') || (password[i] >= '0' && password[i] <= '9') || (password[i] >= 'A' && password[i] <= 'Z')) {
			flagPassword = true
		}
	}
	if flagPassword {
		return nil, errors.New("password format Error")
	}

	//邮箱是否被注册过
	if flag := dao.CheckEmail(email); flag {
		return nil, errors.New("email Already Exist")
	}

	//验证成功，填入UserID
	userID := strconv.Itoa(dao.ID)
	dao.ID++

	//创建用户
	user := &model.User{
		Username: username,
		Password: password,
		Email:    email,
		UserID:   userID,
	}

	//传入数据库
	dao.AddUser(user.Username, user.Password, user.Email, userID)

	return user, nil
}

func Login(loginKey, password string) (string, *model.User, error) {
	//判断是邮箱登录还是用户名登录
	var username string
	var at, point bool
	for i := 0; i < len(loginKey); i++ {
		if loginKey[i] == '@' {
			at = true
		}
		if loginKey[i] == '.' {
			point = true
		}
	}
	if at && point { // 邮箱登录
		username = dao.EmailToUsername(loginKey)
	} else { // 用户名登录
		username = loginKey
		loginKey = dao.UsernameToEmail(username)
	}

	//检查用户是否存在
	if flag := dao.CheckUsername(username); !flag {
		return "", nil, errors.New("username Not Exist")
	}

	//检验密码正确性
	if password != dao.SelectPassword(username) {
		return "", nil, errors.New("password error")
	}

	//返回token
	token, err := util.GenerateToken(dao.CheckID(username), username)
	if err != nil {
		return "", nil, errors.New("token Error")
	}

	user := &model.User{
		Username: username,
		UserID:   dao.CheckUserID(username),
		Email:    loginKey,
	}

	return token, user, nil
}
