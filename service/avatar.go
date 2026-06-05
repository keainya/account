package service

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/keainya/account/object"
	"github.com/keainya/account/utils"
)

const maxAvatarSize = 5 << 20 // 5MB
const avatarDir = "data/avatar"

// UploadAvatar 上传头像
func UploadAvatar(c *gin.Context) {
	userVal, exists := c.Get(string(utils.ContextKeyUser))
	if !exists {
		c.JSON(200, Response{Code: 2001, Msg: "未登录"})
		return
	}
	user := userVal.(object.User)

	file, header, err := c.Request.FormFile("avatar")
	if err != nil {
		c.JSON(200, Response{Code: 1002, Msg: "请选择图片文件"})
		return
	}
	defer file.Close()

	if header.Size > maxAvatarSize {
		c.JSON(200, Response{Code: 1002, Msg: "文件过大，请压缩至 5MB 以内"})
		return
	}

	// 验证文件类型
	ext := strings.ToLower(filepath.Ext(header.Filename))
	switch ext {
	case ".jpg", ".jpeg", ".png", ".webp":
	default:
		c.JSON(200, Response{Code: 1002, Msg: "仅支持 JPG / PNG / WebP 格式"})
		return
	}
	if ext == ".jpeg" {
		ext = ".jpg"
	}

	// 删除旧头像文件
	if user.Avatar != "" {
		oldPath := filepath.Join(avatarDir, user.Avatar)
		os.Remove(oldPath)
	}

	// 保存新头像
	newFilename := fmt.Sprintf("%s%s", user.ID, ext)
	savePath := filepath.Join(avatarDir, newFilename)
	if err := c.SaveUploadedFile(header, savePath); err != nil {
		c.JSON(200, Response{Code: 9000, Msg: "保存头像失败"})
		return
	}

	// 更新数据库
	if err := object.Database.Model(&user).Update("avatar", newFilename).Error; err != nil {
		os.Remove(savePath)
		c.JSON(200, Response{Code: 9001, Msg: "数据库错误"})
		return
	}

	c.JSON(200, Response{Code: 0, Msg: "头像上传成功", Data: gin.H{"avatar": newFilename}})
}

// GetAvatar 获取头像文件
func GetAvatar(c *gin.Context) {
	filename := c.Param("filename")
	// 防止路径遍历
	filename = filepath.Base(filename)

	filePath := filepath.Join(avatarDir, filename)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		c.JSON(404, Response{Code: -1, Msg: "头像不存在"})
		return
	}
	c.File(filePath)
}
