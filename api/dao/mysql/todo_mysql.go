package mysql

import (
	"GoGin/api/dao"
	"GoGin/api/dao/cache"
	"GoGin/internal/model"
	"errors"
	"fmt"
	"log"
	"time"

	"gorm.io/gorm"
)

type mysqlTodoRepo struct {
	db    *gorm.DB
	cache *cache.RedisClient
}

func NewMysqlTodoRepo(db *gorm.DB, cache *cache.RedisClient) dao.TodoRepository {
	err := db.AutoMigrate(&model.TodoTask{})
	if err != nil {
		log.Fatal("Failed to migrate student & course table:", err)
	}

	return &mysqlTodoRepo{
		db:    db,
		cache: cache,
	}
}

func (repo *mysqlTodoRepo) CreateTodoTask(task *model.TodoTask) error {
	if err := repo.db.Create(task).Error; err != nil {
		return errors.New("failed to create task")
	}

	//写后删除
	if repo.cache != nil {
		todosKey := fmt.Sprintf("todo:user:%d:todos", task.UserID)
		donesKey := fmt.Sprintf("todo:user:%d:dones", task.UserID)
		err := repo.cache.Clean(todosKey, donesKey)
		if err != nil {
			return errors.New("failed to clean redis key: dones,todos")
		}
	}

	return nil
}

func (repo *mysqlTodoRepo) DeleteTodoTask(taskID int) error {
	var task model.TodoTask
	if err := repo.db.First(&task, taskID).Error; err != nil {
		return errors.New("task not found")
	}

	if err := repo.db.Delete(&model.TodoTask{}, taskID).Error; err != nil {
		return errors.New("failed to delete task")
	}

	// 写后删除
	if repo.cache != nil {
		todosKey := fmt.Sprintf("todo:user:%d:todos", task.UserID)
		donesKey := fmt.Sprintf("todo:user:%d:dones", task.UserID)
		err := repo.cache.Clean(todosKey, donesKey)
		if err != nil {
			return errors.New("failed to clean redis key: dones,todos")
		}
	}

	return nil
}

func (repo *mysqlTodoRepo) FinishTodoTask(taskID int) error {
	var task model.TodoTask
	if err := repo.db.First(&task, taskID).Error; err != nil {
		return errors.New("task not found")
	}

	if err := repo.db.Where("task_id = ?", taskID).Update("completed", true).Error; err != nil {
		return errors.New("failed to finish task")
	}

	// 写后删除
	if repo.cache != nil {
		todosKey := fmt.Sprintf("todo:user:%d:todos", task.UserID)
		donesKey := fmt.Sprintf("todo:user:%d:dones", task.UserID)
		err := repo.cache.Clean(todosKey, donesKey)
		if err != nil {
			return errors.New("failed to clean redis key: dones,todos")
		}
	}
	return nil
}

func (repo *mysqlTodoRepo) CheckTodoTask(userID int) ([]model.TodoTask, []model.TodoTask, error) {
	// 尝试从缓存获取
	if repo.cache != nil {
		todosKey := fmt.Sprintf("todo:user:%d:todos", userID)
		donesKey := fmt.Sprintf("todo:user:%d:dones", userID)

		var todos []model.TodoTask
		var dones []model.TodoTask

		todosErr := repo.cache.Get(todosKey, &todos)
		donesErr := repo.cache.Get(donesKey, &dones)

		if todosErr == nil && donesErr == nil {
			return todos, dones, nil
		}
	}

	//缓存未命中，查询数据库
	var todos []model.TodoTask
	var dones []model.TodoTask
	if err := repo.db.Where("user_id = ? AND complete = ?", userID, false).Find(&todos).Error; err != nil {
		return nil, nil, errors.New("failed to check task")
	}
	if err := repo.db.Where("user_id = ? AND complete = ?", userID, true).Where("complete = ?", true).Find(&dones).Error; err != nil {
		return nil, nil, errors.New("failed to check task")
	}

	// 写入缓存
	if repo.cache != nil {
		todosKey := fmt.Sprintf("todo:user:%d:todos", userID)
		donesKey := fmt.Sprintf("todo:user:%d:dones", userID)

		// 使用分布式锁防止缓存击穿
		lockKey := fmt.Sprintf("lock:todo:user:%d", userID)
		if success, _ := repo.cache.Lock(lockKey, 10*time.Second); success {
			defer repo.cache.Unlock(lockKey)

			err := repo.cache.Set(todosKey, todos, repo.cache.RandExp(2*time.Minute))
			if err != nil {
				return nil, nil, errors.New("failed to write cache")
			}
			err = repo.cache.Set(donesKey, dones, repo.cache.RandExp(2*time.Minute))
			if err != nil {
				return nil, nil, errors.New("failed to write cache")
			}
		}
	}
	return todos, dones, nil
}
