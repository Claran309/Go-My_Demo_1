# ClaranDemo

## 简介

自学过程中开发的一个Go语言本地项目，基于Gin框架构建，实现了完整的用户认证、课程管理和任务清单功能。项目采用渐进式开发，从基础功能开始逐步完善架构和功能模块。

技术力低，大佬勿喷qwq

## **[API测试文档](https://s.apifox.cn/f9fc1e2c-28c6-4f97-97a9-957a088ac638)**

## [更新日志](https://www.claran-blog.work/2025/11/04/Go-Demo%EF%BC%9A%E7%94%A8%E6%88%B7%E6%B3%A8%E5%86%8C%E4%B8%8E%E7%99%BB%E5%BD%95%E7%B3%BB%E7%BB%9F/)

## Future plans：
- 建立合理的错误抛出和处理机制
- 完善信息修改接口系列（PUT）：
    - 课程信息修改
    - 密码修改
    - 用户名修改
    - 待办事项修改
- Docker容器化部署

## Docker
```bash
# 创建 .env 文件
cp .env.example .env
# 编辑 .env 文件

# 启动所有服务
docker-compose up -d

# 查看运行状态
docker-compose ps

# 查看日志
docker-compose logs -f web
docker-compose logs -f mysql
docker-compose logs -f redis

# 停止服务
docker-compose down

# 停止并删除数据卷
docker-compose down -v

# 检查容器状态
docker ps

# 应该看到：
# MySQL
# Redis
# ClaranDemo

# 测试连接
curl http://localhost:8080/health
```

## 技术栈
### 后端框架
- **Gin** - 高性能Go Web框架
- **GORM** - ORM库，数据库操作

### 数据存储
- **MySQL** - 关系型数据库，数据持久化存储
- **Redis** - 缓存数据库，提升系统性能

### 认证授权
- **JWT** - JSON Web Tokens无状态认证机制
- **bcrypt** - 密码加密算法

### 架构设计
- 面向接口编程
- 依赖注入
- 分层架构（Handler → Service → Dao）
- 环境配置管理（.env）

### 开发工具
- Go Modules - 依赖管理
- 内置测试支持