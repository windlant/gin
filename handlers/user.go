package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/windlant/gin/models"
)

var users = []models.User{
	{ID: 1, Name: "Alice", Email: "alice@example.com"},
}
var nextID = 2

// GetUsers 获取所有用户
func GetUsers(c *gin.Context) {
	c.JSON(http.StatusOK, users)
}

// GetUser 根据 ID 获取单个用户
func GetUser(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}
	for _, u := range users {
		if u.ID == id {
			c.JSON(http.StatusOK, u)
			return
		}
	}
	c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
}

// CreateUsers 批量创建用户
// 请求体: [{ "name": "...", "email": "..." }, ...]
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
// 请求体: [{ "id": 1, "name": "...", "email": "..." }, ...]
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
				// 保留原 ID，更新其他字段
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
// 请求体: { "ids": [1, 2, 3] }
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

	// 重建 users slice，跳过要删除的
	toKeep := make([]models.User, 0, len(users))
	for _, u := range users {
		if delSet[u.ID] {
			deleted = append(deleted, u.ID)
		} else {
			toKeep = append(toKeep, u)
		}
	}
	users = toKeep

	// 计算 not_found
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
