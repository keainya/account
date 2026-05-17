package service

import (
	"strings"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/keainya/service_temp/object"
	"github.com/keainya/service_temp/utils"
)

// AuthRequired Session 登录校验中间件
func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		userID := session.Get("user_id")
		if userID == nil {
			c.AbortWithStatusJSON(200, Response{Code: 2001, Msg: "未登录"})
			return
		}

		var user object.User
		if err := object.Database.Where("id = ?", userID).First(&user).Error; err != nil {
			c.AbortWithStatusJSON(200, Response{Code: 2001, Msg: "用户不存在"})
			return
		}

		c.Set(string(utils.ContextKeyUserID), user.ID)
		c.Set(string(utils.ContextKeyUsername), user.Username)
		c.Set(string(utils.ContextKeyRole), user.Role)
		c.Set(string(utils.ContextKeyUser), user)
		c.Next()
	}
}

// AdminRequired 管理员权限校验中间件
func AdminRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get(string(utils.ContextKeyRole))
		if !exists || role.(string) != "admin" {
			c.AbortWithStatusJSON(200, Response{Code: 2002, Msg: "无权限（非管理员）"})
			return
		}
		c.Next()
	}
}

// BearerAuth Bearer Token 校验中间件
func BearerAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			c.AbortWithStatusJSON(200, Response{Code: 3008, Msg: "access_token 无效或已过期"})
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := utils.ParseToken(tokenString)
		if err != nil {
			c.AbortWithStatusJSON(200, Response{Code: 3008, Msg: "access_token 无效或已过期"})
			return
		}

		c.Set(string(utils.ContextKeyUserID), claims.Subject)
		c.Set(string(utils.ContextKeyUsername), claims.Username)
		c.Set(string(utils.ContextKeyRole), claims.Role)
		c.Set(string(utils.ContextKeyClientID), claims.ClientID)
		c.Next()
	}
}
