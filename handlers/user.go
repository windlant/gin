package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/windlant/gin/models"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// 全局变量（实际项目中建议通过依赖注入）
var (
	db    *gorm.DB
	rdb   *redis.Client
	users = []models.User{
		{ID: 1, Name: "Alice", Email: "alice@example.com"},
	}
	nextID = 2
)

// InitDBAndRedis 初始化数据库和Redis连接
func InitDBAndRedis() error {
	// === MySQL 连接配置 ===
	dsn := "root:G1y2#X.c9Om!@tcp(127.0.0.1:3306)/goframe_demo?charset=utf8mb4&parseTime=True&loc=Local"

	var err error

	// 打开数据库连接
	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		// 禁用外键约束检查
		DisableForeignKeyConstraintWhenMigrating: true,
		// 跳过默认事务
		SkipDefaultTransaction: true,
	})
	if err != nil {
		return fmt.Errorf("failed to connect to MySQL: %w", err)
	}

	// 验证数据库连接是否有效
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get database instance: %w", err)
	}

	// 设置连接池参数
	sqlDB.SetMaxOpenConns(25)                 // 最大打开连接数
	sqlDB.SetMaxIdleConns(5)                  // 最大空闲连接数
	sqlDB.SetConnMaxLifetime(time.Minute * 3) // 连接最大生命周期

	// 测试数据库连接
	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("failed to ping MySQL database: %w", err)
	}

	// === Redis 连接配置 ===
	rdb = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
		// 连接池配置
		PoolSize:     10,               // 连接池大小
		MinIdleConns: 5,                // 最小空闲连接
		MaxConnAge:   0,                // 连接最大年龄（0表示无限制）
		PoolTimeout:  time.Second * 30, // 从连接池获取连接的超时时间
		ReadTimeout:  time.Second * 3,  // 读取超时
		WriteTimeout: time.Second * 3,  // 写入超时
	})

	// 测试 Redis 连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if _, err := rdb.Ping(ctx).Result(); err != nil {
		return fmt.Errorf("failed to connect to Redis: %w", err)
	}

	log.Println("Database and Redis initialized successfully")
	return nil
}

// UserDB 是数据库模型（与 API 模型分离）
type UserDB struct {
	ID        uint      `gorm:"primaryKey;autoIncrement"`
	Name      string    `gorm:"size:64;not null;default:''"`
	Email     string    `gorm:"size:128;not null;uniqueIndex"`
	CreatedAt time.Time `gorm:"not null;default:CURRENT_TIMESTAMP"`
}

func (u *UserDB) ToAPIModel() models.User {
	return models.User{
		ID:    int(u.ID),
		Name:  u.Name,
		Email: u.Email,
	}
}
func (UserDB) TableName() string {
	return "users"
}

func GetUser(c *gin.Context) {
	type input struct {
		ID int `json:"id" binding:"required"`
	}
	var in input
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	id := in.ID

	// 先从 Redis 缓存中查
	cacheKey := "user:" + strconv.Itoa(id)
	var cachedUser models.User

	// 尝试从 Redis 获取
	ctx := c.Request.Context()
	cachedData, err := rdb.Get(ctx, cacheKey).Result()
	if err == nil {
		// 缓存命中，解析 JSON
		err = json.Unmarshal([]byte(cachedData), &cachedUser)
		if err == nil {
			c.JSON(http.StatusOK, cachedUser)
			return
		}
	}

	var userDB UserDB
	result := db.Where("id = ?", id).First(&userDB)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		}
		return
	}

	user := userDB.ToAPIModel()

	userJSON, _ := json.Marshal(user)
	rdb.Set(ctx, cacheKey, userJSON, 5*time.Minute)

	c.JSON(http.StatusOK, user)
}

// GetUsers 获取所有用户
func GetUsers(c *gin.Context) {
	c.JSON(http.StatusOK, users)
}

// CreateUsers 批量创建用户
func CreateUsers(c *gin.Context) {
	var inputs []models.User
	if err := c.ShouldBindJSON(&inputs); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var created []models.User
	var errors []string

	for _, input := range inputs {
		if input.Name == "" || input.Email == "" {
			errors = append(errors, "Name and Email are required")
			continue
		}
		input.ID = nextID
		nextID++
		users = append(users, input)
		created = append(created, input)
	}

	c.JSON(http.StatusCreated, gin.H{
		"created": created,
		"errors":  errors,
	})
}

// UpdateUsers 批量更新用户
func UpdateUsers(c *gin.Context) {
	var inputs []models.User
	if err := c.ShouldBindJSON(&inputs); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var updated []models.User
	var notFound []int

	for _, input := range inputs {
		if input.ID <= 0 {
			notFound = append(notFound, input.ID)
			continue
		}

		found := false
		for i, u := range users {
			if u.ID == input.ID {
				users[i].Name = input.Name
				users[i].Email = input.Email
				updated = append(updated, users[i])
				found = true
				break
			}
		}
		if !found {
			notFound = append(notFound, input.ID)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"updated":   updated,
		"not_found": notFound,
	})
}

// DeleteUsers 批量删除用户
func DeleteUsers(c *gin.Context) {
	var req struct {
		IDs []int `json:"ids" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var deleted []int
	delSet := make(map[int]bool)
	for _, id := range req.IDs {
		delSet[id] = true
	}

	toKeep := make([]models.User, 0, len(users))
	for _, u := range users {
		if delSet[u.ID] {
			deleted = append(deleted, u.ID)
		} else {
			toKeep = append(toKeep, u)
		}
	}
	users = toKeep

	foundSet := make(map[int]bool)
	for _, id := range deleted {
		foundSet[id] = true
	}
	var notFound []int
	for _, id := range req.IDs {
		if !foundSet[id] {
			notFound = append(notFound, id)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"deleted":   deleted,
		"not_found": notFound,
	})
}
