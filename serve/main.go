package main

import (
	"GoGin/api/routes"
)

func main() {
	routes.InitRouter()
}

/*
注册时：
前端JSON:
username
password
email

登录时：
前端JSON：
loginkey
password
*/
