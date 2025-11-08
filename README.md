## [My Blog](https://www.claran-blog.work/)

## Demo：用户注册与登录系统
自学过程中搓的一个小Demo

本地端口监听，map代替数据库

知识点：Gin，http，JWT

技术力低，大佬勿喷

****

# V1
- 无本地文档数据永久性存储，仅运行时暂时性存储数据
- 外层并发锁
- 普通的`username/email`注册 / 登录系统

****

# V2

- 从前端捕获的数据从表单变为JSON
- 新增`user_id`存储唯一用户信息
- 新增JWT鉴权

****

# V3

- 重构了一下项目结构，现在更规范了