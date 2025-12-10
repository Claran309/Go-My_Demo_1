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

type mysqlCourseRepo struct {
	db    *gorm.DB
	cache *cache.RedisClient
}

func NewMysqlCourseRepo(db *gorm.DB, cache *cache.RedisClient) dao.CourseRepository {
	err := db.AutoMigrate(&model.Student{}, &model.Course{})
	if err != nil {
		log.Fatal("Failed to migrate student & course table:", err)
	}
	err = db.AutoMigrate(&model.Enrollment{})
	if err != nil {
		log.Fatal("Failed to migrate enrollment table:", err)
	}

	return &mysqlCourseRepo{
		db:    db,
		cache: cache,
	}
}

func (repo *mysqlCourseRepo) PickCourse(StudentID, CourseID int) error {
	// 分布式锁
	if repo.cache != nil {
		lockKey := fmt.Sprintf("lock:pick:%d:%d", StudentID, CourseID)
		if success, _ := repo.cache.Lock(lockKey, 5*time.Second); !success {
			return errors.New("system busy, please try again")
		}
		defer repo.cache.Unlock(lockKey)
	}

	err := repo.db.Transaction(func(tx *gorm.DB) error {
		// 检查学生是否存在
		var student model.Student
		if err := tx.First(&student, StudentID).Error; err != nil {
			return errors.New("student Not Found")
		}

		// 检查课程是否存在
		var course model.Course
		if err := tx.First(&course, CourseID).Error; err != nil {
			return errors.New("course Not Found")
		}

		// 检查课程是否已满
		if course.Enroll >= course.Capital {
			return errors.New("course is full")
		}

		// 是否重复选择
		var exists int64
		if err := tx.Model(&model.Enrollment{}).
			Where("student_id = ? AND course_id = ?", StudentID, CourseID).
			Count(&exists).Error; err != nil {
			return err
		}
		if exists >= 1 {
			return errors.New("enrollment exists")
		}

		// 创建选课关系
		enrollment := model.Enrollment{
			StudentID: StudentID,
			CourseID:  CourseID,
		}
		if err := tx.Create(&enrollment).Error; err != nil {
			return errors.New("enrollment create failed")
		}

		// 更新选课人数
		if err := tx.Model(&model.Course{}).
			Where("course_id = ?", CourseID).
			Update("enroll", gorm.Expr("enroll + ?", 1)).Error; err != nil {
			return err
		}

		return nil
	})

	if repo.cache != nil && err == nil {
		// 清除课程列表缓存
		err := repo.cache.Clean("course:all")
		if err != nil {
			return errors.New("cache clean failed")
		}
		// 清除该课程缓存
		courseKey := fmt.Sprintf("course:%d", CourseID)
		err = repo.cache.Clean(courseKey)
		if err != nil {
			return errors.New("cache clean failed")
		}
		// 清除学生的选课记录缓存
		enrollKey := fmt.Sprintf("enroll:student:%d", StudentID)
		err = repo.cache.Clean(enrollKey)
		if err != nil {
			return errors.New("cache clean failed")
		}
	}
	return nil
}

func (repo *mysqlCourseRepo) DropCourse(StudentID, CourseID int) error {
	// 分布式锁
	if repo.cache != nil {
		lockKey := fmt.Sprintf("lock:drop:%d:%d", StudentID, CourseID)
		if success, _ := repo.cache.Lock(lockKey, 5*time.Second); !success {
			return errors.New("system busy, please try again")
		}
		defer repo.cache.Unlock(lockKey)
	}

	err := repo.db.Transaction(func(tx *gorm.DB) error {
		//是否存在记录
		var enrollment model.Enrollment
		if err := tx.Where("student_id = ? AND course_id = ?", StudentID, CourseID).
			First(&enrollment).Error; err != nil {
			return errors.New("enrollment Not Found")
		}

		//删除
		if err := tx.Delete(&enrollment).Error; err != nil {
			return errors.New("delete failed")
		}

		//更新人数
		if err := tx.Model(&model.Course{}).
			Where("course_id = ?", CourseID).
			Update("enroll", gorm.Expr("enroll - ?", 1)).Error; err != nil {
			return errors.New("update failed")
		}

		return nil
	})

	if repo.cache != nil && err == nil {
		// 清除课程列表缓存
		err := repo.cache.Clean("course:all")
		if err != nil {
			return errors.New("cache clean failed")
		}
		// 清除该课程缓存
		courseKey := fmt.Sprintf("course:%d", CourseID)
		err = repo.cache.Clean(courseKey)
		if err != nil {
			return errors.New("cache clean failed")
		}
		// 清除学生的选课记录缓存
		enrollKey := fmt.Sprintf("enroll:student:%d", StudentID)
		err = repo.cache.Clean(enrollKey)
		if err != nil {
			return errors.New("cache clean failed")
		}
	}

	return nil
}

func (repo *mysqlCourseRepo) CheckEnrollment(studentID int) ([]model.Enrollment, error) {
	// 尝试从缓存获取
	if repo.cache != nil {
		cacheKey := fmt.Sprintf("enroll:student:%d", studentID)
		var enrollments []model.Enrollment
		if err := repo.cache.Get(cacheKey, &enrollments); err == nil {
			return enrollments, nil
		}
	}

	// 数据库
	var enrollment []model.Enrollment
	if err := repo.db.Where("student_id = ?", studentID).First(&enrollment).Error; err != nil {
		return nil, errors.New("enrollment select failed")
	}

	// 写入缓存
	if repo.cache != nil {
		cacheKey := fmt.Sprintf("enroll:student:%d", studentID)
		// 使用分布式锁
		lockKey := fmt.Sprintf("lock:enroll:student:%d", studentID)
		if success, _ := repo.cache.Lock(lockKey, 10*time.Second); success {
			defer repo.cache.Unlock(lockKey)
			err := repo.cache.Set(cacheKey, enrollment, repo.cache.RandExp(2*time.Minute))
			if err != nil {
				return nil, errors.New("cache set failed")
			}
		}
	}

	return enrollment, nil
}

func (repo *mysqlCourseRepo) CheckInfo() ([]model.Course, error) {
	// 尝试从缓存获取
	if repo.cache != nil {
		var courses []model.Course
		if err := repo.cache.Get("course:all", &courses); err == nil {
			return courses, nil
		}
	}

	//数据库
	var course []model.Course
	if err := repo.db.Find(&course).Error; err != nil {
		return nil, errors.New("course select failed")
	}

	// 写入缓存
	if repo.cache != nil {
		// 使用分布式锁
		if success, _ := repo.cache.Lock("lock:course:all", 10*time.Second); success {
			defer repo.cache.Unlock("lock:course:all")
			err := repo.cache.Set("course:all", course, repo.cache.RandExp(2*time.Minute))
			if err != nil {
				return nil, errors.New("cache set failed")
			}
		}
	}

	return course, nil
}

func (repo *mysqlCourseRepo) AddCourse(Course model.Course) error {
	if err := repo.db.Create(&Course).Error; err != nil {
		return errors.New("course create failed")
	}

	// 写后删除
	if repo.cache != nil {
		err := repo.cache.Clean("course:all")
		if err != nil {
			return errors.New("cache clean failed")
		}
		// 缓存新创建的课程
		courseKey := fmt.Sprintf("course:%d", Course.ID)
		err = repo.cache.Set(courseKey, Course, repo.cache.RandExp(5*time.Minute))
		if err != nil {
			return errors.New("cache set failed")
		}
	}

	return nil
}

func (repo *mysqlCourseRepo) CheckCourse(courseID int) (model.Course, error) {
	// 尝试从缓存获取
	if repo.cache != nil {
		cacheKey := fmt.Sprintf("course:%d", courseID)
		var course model.Course
		if err := repo.cache.Get(cacheKey, &course); err == nil {
			return course, nil
		}
	}

	//数据库
	var course model.Course
	if err := repo.db.First(&course, courseID).Error; err != nil {
		return model.Course{}, errors.New("course not found")
	}

	// 写入缓存
	if repo.cache != nil {
		cacheKey := fmt.Sprintf("course:%d", courseID)
		// 使用分布式锁
		lockKey := fmt.Sprintf("lock:course:%d", courseID)
		if success, _ := repo.cache.Lock(lockKey, 10*time.Second); success {
			defer repo.cache.Unlock(lockKey)
			err := repo.cache.Set(cacheKey, course, repo.cache.RandExp(5*time.Minute))
			if err != nil {
				return model.Course{}, errors.New("cache set failed")
			}
		}
	}
	return course, nil
}
