package service

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func Status(c *gin.Context) {
	c.JSON(200, DefaultResponse())
}

// SessionSet 设置 session 值。Query: key, value
func SessionSet(c *gin.Context) {
	session := sessions.Default(c)
	key := c.Query("key")
	value := c.Query("value")
	if key == "" {
		c.JSON(400, Response{Code: -1, Msg: "key is required"})
		return
	}
	session.Set(key, value)
	if err := session.Save(); err != nil {
		c.JSON(500, Response{Code: -1, Msg: "failed to save session"})
		return
	}
	c.JSON(200, Response{Code: 0, Msg: "ok", Data: gin.H{key: value}})
}

// SessionGet 获取 session 值。Query: key
func SessionGet(c *gin.Context) {
	session := sessions.Default(c)
	key := c.Query("key")
	if key == "" {
		c.JSON(400, Response{Code: -1, Msg: "key is required"})
		return
	}
	value := session.Get(key)
	c.JSON(200, Response{Code: 0, Msg: "ok", Data: gin.H{key: value}})
}

// SessionDel 删除 session 值。Query: key
func SessionDel(c *gin.Context) {
	session := sessions.Default(c)
	key := c.Query("key")
	if key == "" {
		c.JSON(400, Response{Code: -1, Msg: "key is required"})
		return
	}
	session.Delete(key)
	if err := session.Save(); err != nil {
		c.JSON(500, Response{Code: -1, Msg: "failed to save session"})
		return
	}
	c.JSON(200, Response{Code: 0, Msg: "ok"})
}
