package service

import (
	"encoding/json"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/keainya/service_temp/object"
	"github.com/keainya/service_temp/utils"
)

// ---------- 用户管理 ----------

// AdminListUsers 用户列表
func AdminListUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	var total int64
	object.Database.Model(&object.User{}).Count(&total)

	var users []object.User
	object.Database.Order("created_at desc").Offset((page - 1) * pageSize).Limit(pageSize).Find(&users)

	c.JSON(200, Response{
		Code: 0,
		Msg:  "ok",
		Data: gin.H{
			"items":     users,
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		},
	})
}

// AdminPromoteUser 提升为管理员
func AdminPromoteUser(c *gin.Context) {
	userID := c.Param("user_id")

	var user object.User
	if err := object.Database.Where("id = ?", userID).First(&user).Error; err != nil {
		c.JSON(200, Response{Code: 5001, Msg: "用户不存在"})
		return
	}
	if user.Role == "admin" {
		c.JSON(200, Response{Code: 5002, Msg: "用户已是管理员"})
		return
	}

	object.Database.Model(&user).Updates(map[string]interface{}{
		"role": "admin",
	})

	c.JSON(200, Response{
		Code: 0,
		Msg:  "已提升为管理员",
		Data: gin.H{
			"id":         user.ID,
			"username":   user.Username,
			"role":       "admin",
			"updated_at": user.UpdatedAt,
		},
	})
}

// AdminDemoteUser 降级为普通用户
func AdminDemoteUser(c *gin.Context) {
	userID := c.Param("user_id")
	currentUserID := c.GetString(string(utils.ContextKeyUserID))

	if userID == currentUserID {
		c.JSON(200, Response{Code: 5003, Msg: "不能降级自己"})
		return
	}

	var user object.User
	if err := object.Database.Where("id = ?", userID).First(&user).Error; err != nil {
		c.JSON(200, Response{Code: 5001, Msg: "用户不存在"})
		return
	}
	if user.Role != "admin" {
		c.JSON(200, Response{Code: 5002, Msg: "用户已是普通用户"})
		return
	}

	// 确保不是最后一个管理员
	var adminCount int64
	object.Database.Model(&object.User{}).Where("role = ?", "admin").Count(&adminCount)
	if adminCount <= 1 {
		c.JSON(200, Response{Code: 5004, Msg: "系统最后一个管理员不可降级"})
		return
	}

	object.Database.Model(&user).Updates(map[string]interface{}{
		"role": "user",
	})

	c.JSON(200, Response{
		Code: 0,
		Msg:  "已降级为普通用户",
		Data: gin.H{
			"id":         user.ID,
			"username":   user.Username,
			"role":       "user",
			"updated_at": user.UpdatedAt,
		},
	})
}

// ---------- 应用管理 ----------

// AdminCreateApp 创建应用
func AdminCreateApp(c *gin.Context) {
	var req struct {
		Name         string   `json:"name"`
		Description  string   `json:"description"`
		RedirectURIs []string `json:"redirect_uris"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Name == "" || len(req.RedirectURIs) == 0 {
		c.JSON(200, Response{Code: 1002, Msg: "参数校验失败"})
		return
	}

	clientID, err := utils.GenerateRandomString(8)
	if err != nil {
		c.JSON(200, Response{Code: 9000, Msg: "服务器内部错误"})
		return
	}
	clientID = "app_" + clientID[:16]

	clientSecret, err := utils.GenerateRandomString(16)
	if err != nil {
		c.JSON(200, Response{Code: 9000, Msg: "服务器内部错误"})
		return
	}
	clientSecret = "sec_" + clientSecret

	urisJSON, _ := json.Marshal(req.RedirectURIs)

	app := object.App{
		ID:           uuid.NewString(),
		Name:         req.Name,
		Description:  req.Description,
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURIs: string(urisJSON),
	}
	if err := object.Database.Create(&app).Error; err != nil {
		c.JSON(200, Response{Code: 9001, Msg: "数据库错误"})
		return
	}

	c.JSON(201, Response{
		Code: 0,
		Msg:  "应用创建成功",
		Data: gin.H{
			"id":            app.ID,
			"name":          app.Name,
			"description":   app.Description,
			"client_id":     app.ClientID,
			"client_secret": app.ClientSecret,
			"redirect_uris": req.RedirectURIs,
			"created_at":    app.CreatedAt,
		},
	})
}

// AdminListApps 应用列表
func AdminListApps(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	var total int64
	object.Database.Model(&object.App{}).Count(&total)

	var apps []object.App
	object.Database.Order("created_at desc").Offset((page - 1) * pageSize).Limit(pageSize).Find(&apps)

	// 列表中不返回 client_secret
	type appItem struct {
		ID           string   `json:"id"`
		Name         string   `json:"name"`
		Description  string   `json:"description"`
		ClientID     string   `json:"client_id"`
		RedirectURIs []string `json:"redirect_uris"`
		CreatedAt    string   `json:"created_at"`
	}
	items := make([]appItem, 0, len(apps))
	for _, a := range apps {
		var uris []string
		json.Unmarshal([]byte(a.RedirectURIs), &uris)
		items = append(items, appItem{
			ID:           a.ID,
			Name:         a.Name,
			Description:  a.Description,
			ClientID:     a.ClientID,
			RedirectURIs: uris,
			CreatedAt:    a.CreatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}

	c.JSON(200, Response{
		Code: 0,
		Msg:  "ok",
		Data: gin.H{
			"items":     items,
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		},
	})
}

// AdminGetApp 获取单个应用
func AdminGetApp(c *gin.Context) {
	appID := c.Param("app_id")

	var app object.App
	if err := object.Database.Where("id = ?", appID).First(&app).Error; err != nil {
		c.JSON(200, Response{Code: 1004, Msg: "应用不存在"})
		return
	}

	var uris []string
	json.Unmarshal([]byte(app.RedirectURIs), &uris)

	c.JSON(200, Response{
		Code: 0,
		Msg:  "ok",
		Data: gin.H{
			"id":            app.ID,
			"name":          app.Name,
			"description":   app.Description,
			"client_id":     app.ClientID,
			"client_secret": app.ClientSecret,
			"redirect_uris": uris,
			"created_at":    app.CreatedAt,
		},
	})
}

// AdminUpdateApp 更新应用
func AdminUpdateApp(c *gin.Context) {
	appID := c.Param("app_id")

	var app object.App
	if err := object.Database.Where("id = ?", appID).First(&app).Error; err != nil {
		c.JSON(200, Response{Code: 1004, Msg: "应用不存在"})
		return
	}

	var req struct {
		Name         *string   `json:"name"`
		Description  *string   `json:"description"`
		RedirectURIs *[]string `json:"redirect_uris"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(200, Response{Code: 1002, Msg: "参数校验失败"})
		return
	}

	updates := map[string]interface{}{}
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.RedirectURIs != nil {
		urisJSON, _ := json.Marshal(*req.RedirectURIs)
		updates["redirect_uris"] = string(urisJSON)
	}

	if len(updates) > 0 {
		object.Database.Model(&app).Updates(updates)
	}

	c.JSON(200, Response{
		Code: 0,
		Msg:  "应用已更新",
		Data: gin.H{
			"id":          app.ID,
			"name":        app.Name,
			"description": app.Description,
		},
	})
}

// AdminResetAppSecret 重置密钥
func AdminResetAppSecret(c *gin.Context) {
	appID := c.Param("app_id")

	var app object.App
	if err := object.Database.Where("id = ?", appID).First(&app).Error; err != nil {
		c.JSON(200, Response{Code: 1004, Msg: "应用不存在"})
		return
	}

	newSecret, err := utils.GenerateRandomString(16)
	if err != nil {
		c.JSON(200, Response{Code: 9000, Msg: "服务器内部错误"})
		return
	}
	newSecret = "sec_" + newSecret

	object.Database.Model(&app).Update("client_secret", newSecret)

	c.JSON(200, Response{
		Code: 0,
		Msg:  "密钥已重置",
		Data: gin.H{
			"client_secret": newSecret,
		},
	})
}

// AdminDeleteApp 删除应用
func AdminDeleteApp(c *gin.Context) {
	appID := c.Param("app_id")

	var app object.App
	if err := object.Database.Where("id = ?", appID).First(&app).Error; err != nil {
		c.JSON(200, Response{Code: 1004, Msg: "应用不存在"})
		return
	}

	// 删除相关元数据
	object.Database.Where("app_id = ?", appID).Delete(&object.UserMetadata{})
	// 删除相关 token
	object.Database.Where("client_id = ?", app.ClientID).Delete(&object.RefreshToken{})
	object.Database.Where("client_id = ?", app.ClientID).Delete(&object.OAuthCode{})
	// 删除应用
	object.Database.Delete(&app)

	c.JSON(200, Response{Code: 0, Msg: "应用已删除"})
}
