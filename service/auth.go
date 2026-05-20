package service

import (
	"regexp"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/keainya/account/object"
	"github.com/keainya/account/utils"
)

var usernameRegex = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_]{2,31}$`)

// Register 用户注册
func Register(c *gin.Context) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Email    string `json:"email"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(200, Response{Code: 1002, Msg: "参数校验失败"})
		return
	}

	// 校验用户名
	if !usernameRegex.MatchString(req.Username) {
		c.JSON(200, Response{Code: 1002, Msg: "用户名需 3-32 字符，字母开头，仅允许字母数字下划线"})
		return
	}

	// 校验密码
	if len(req.Password) < 6 || len(req.Password) > 128 {
		c.JSON(200, Response{Code: 1002, Msg: "密码需 6-128 字符"})
		return
	}

	// 检查是否允许新用户注册
	var totalCount int64
	object.Database.Model(&object.User{}).Count(&totalCount)
	if totalCount > 0 {
		// 非首个用户时，检查注册开关
		var cfg object.SystemConfig
		if err := object.Database.Where("key = ?", "registration_enabled").First(&cfg).Error; err == nil {
			if cfg.Value != "true" {
				c.JSON(200, Response{Code: 1006, Msg: "管理员已关闭新用户注册"})
				return
			}
		}
	}

	// 检查用户名唯一性
	var count int64
	object.Database.Model(&object.User{}).Where("username = ?", req.Username).Count(&count)
	if count > 0 {
		c.JSON(200, Response{Code: 1001, Msg: "用户名已存在"})
		return
	}

	// 第一个注册的用户为管理员
	object.Database.Model(&object.User{}).Count(&count)
	role := "user"
	if count == 0 {
		role = "admin"
	}

	// 哈希密码
	hash, err := utils.HashPassword(req.Password)
	if err != nil {
		c.JSON(200, Response{Code: 9000, Msg: "服务器内部错误"})
		return
	}

	user := object.User{
		ID:           uuid.NewString(),
		Username:     req.Username,
		PasswordHash: hash,
		Email:        req.Email,
		Role:         role,
	}
	if err := object.Database.Create(&user).Error; err != nil {
		c.JSON(200, Response{Code: 9001, Msg: "数据库错误"})
		return
	}

	c.JSON(201, Response{
		Code: 0,
		Msg:  "注册成功",
		Data: gin.H{
			"id":         user.ID,
			"username":   user.Username,
			"email":      user.Email,
			"role":       user.Role,
			"created_at": user.CreatedAt,
		},
	})
}

// Login 用户登录
func Login(c *gin.Context) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(200, Response{Code: 1002, Msg: "参数校验失败"})
		return
	}

	var user object.User
	if err := object.Database.Where("username = ?", req.Username).First(&user).Error; err != nil {
		c.JSON(200, Response{Code: 1004, Msg: "用户不存在"})
		return
	}

	if !utils.CheckPassword(req.Password, user.PasswordHash) {
		c.JSON(200, Response{Code: 1003, Msg: "用户名或密码错误"})
		return
	}

	// 设置 session
	session := sessions.Default(c)
	session.Set("user_id", user.ID)
	session.Set("username", user.Username)
	session.Set("role", user.Role)
	if err := session.Save(); err != nil {
		c.JSON(200, Response{Code: 9000, Msg: "服务器内部错误"})
		return
	}

	c.JSON(200, Response{
		Code: 0,
		Msg:  "登录成功",
		Data: gin.H{
			"id":       user.ID,
			"username": user.Username,
			"email":    user.Email,
			"role":     user.Role,
		},
	})
}

// Logout 用户登出
func Logout(c *gin.Context) {
	session := sessions.Default(c)
	session.Clear()
	session.Save()
	c.JSON(200, Response{Code: 0, Msg: "已登出"})
}

// Me 获取当前用户信息
func Me(c *gin.Context) {
	userVal, exists := c.Get(string(utils.ContextKeyUser))
	if !exists {
		c.JSON(200, Response{Code: 2001, Msg: "未登录"})
		return
	}
	user := userVal.(object.User)
	c.JSON(200, Response{
		Code: 0,
		Msg:  "ok",
		Data: gin.H{
			"id":         user.ID,
			"username":   user.Username,
			"email":      user.Email,
			"role":       user.Role,
			"created_at": user.CreatedAt,
			"updated_at": user.UpdatedAt,
		},
	})
}
