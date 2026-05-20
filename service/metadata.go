package service

import (
	"encoding/json"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/keainya/account/object"
	"github.com/keainya/account/utils"
)

// ---------- Bearer Token 端点 ----------

// MetadataGetUser 获取指定用户元数据 (Bearer Token)
func MetadataGetUser(c *gin.Context) {
	clientID := c.Param("client_id")
	userID := c.Param("user_id")
	tokenClientID := c.GetString(string(utils.ContextKeyClientID))
	if tokenClientID != clientID {
		c.JSON(200, Response{Code: 4001, Msg: "无权访问该应用的数据"})
		return
	}
	var user object.User
	if err := object.Database.Where("id = ?", userID).First(&user).Error; err != nil {
		c.JSON(200, Response{Code: 4002, Msg: "用户不存在"})
		return
	}
	var app object.App
	if err := object.Database.Where("client_id = ?", clientID).First(&app).Error; err != nil {
		c.JSON(200, Response{Code: 4001, Msg: "无权访问该应用的数据"})
		return
	}
	returnMetadata(c, app.ID, userID)
}

// MetadataPutUser 写入指定用户元数据 (Bearer Token)
func MetadataPutUser(c *gin.Context) {
	clientID := c.Param("client_id")
	userID := c.Param("user_id")
	tokenClientID := c.GetString(string(utils.ContextKeyClientID))
	if tokenClientID != clientID {
		c.JSON(200, Response{Code: 4001, Msg: "无权访问该应用的数据"})
		return
	}
	var req struct {
		Metadata json.RawMessage `json:"metadata"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || len(req.Metadata) == 0 {
		c.JSON(200, Response{Code: 4003, Msg: "metadata 格式无效"})
		return
	}
	if !json.Valid(req.Metadata) || req.Metadata[0] != '{' {
		c.JSON(200, Response{Code: 4003, Msg: "metadata 必须是 JSON 对象"})
		return
	}
	var app object.App
	if err := object.Database.Where("client_id = ?", clientID).First(&app).Error; err != nil {
		c.JSON(200, Response{Code: 4001, Msg: "无权访问该应用的数据"})
		return
	}
	upsertMetadata(c, app.ID, userID, string(req.Metadata))
}

// MetadataBatchGet 批量读取用户元数据 (Bearer Token)
func MetadataBatchGet(c *gin.Context) {
	clientID := c.Param("client_id")
	tokenClientID := c.GetString(string(utils.ContextKeyClientID))
	if tokenClientID != clientID {
		c.JSON(200, Response{Code: 4001, Msg: "无权访问该应用的数据"})
		return
	}
	userIDsStr := c.Query("user_ids")
	if userIDsStr == "" {
		c.JSON(200, Response{Code: 1002, Msg: "user_ids 参数必填"})
		return
	}
	userIDs := strings.Split(userIDsStr, ",")
	if len(userIDs) > 100 {
		userIDs = userIDs[:100]
	}

	var app object.App
	if err := object.Database.Where("client_id = ?", clientID).First(&app).Error; err != nil {
		c.JSON(200, Response{Code: 4001, Msg: "无权访问该应用的数据"})
		return
	}

	type result struct {
		UserID    string          `json:"user_id"`
		Metadata  json.RawMessage `json:"metadata"`
		UpdatedAt interface{}     `json:"updated_at"`
	}
	results := make([]result, 0, len(userIDs))
	for _, uid := range userIDs {
		uid = strings.TrimSpace(uid)
		if uid == "" {
			continue
		}
		var m object.UserMetadata
		err := object.Database.Where("user_id = ? AND app_id = ?", uid, app.ID).First(&m).Error
		if err != nil {
			results = append(results, result{
				UserID:    uid,
				Metadata:  json.RawMessage("{}"),
				UpdatedAt: nil,
			})
		} else {
			meta := json.RawMessage(m.Metadata)
			if meta == nil || string(meta) == "" {
				meta = json.RawMessage("{}")
			}
			results = append(results, result{
				UserID:    uid,
				Metadata:  meta,
				UpdatedAt: m.UpdatedAt,
			})
		}
	}
	c.JSON(200, Response{Code: 0, Msg: "ok", Data: results})
}

// ---------- Session 端点 ----------

// MetadataGetMy 读取当前用户在指定应用下的元数据 (Session)
func MetadataGetMy(c *gin.Context) {
	clientID := c.Param("client_id")
	userID := c.GetString(string(utils.ContextKeyUserID))
	var app object.App
	if err := object.Database.Where("client_id = ?", clientID).First(&app).Error; err != nil {
		c.JSON(200, Response{Code: 4001, Msg: "应用不存在"})
		return
	}
	returnMetadata(c, app.ID, userID)
}

// MetadataPutMy 写入当前用户在指定应用下的元数据 (Session)
func MetadataPutMy(c *gin.Context) {
	clientID := c.Param("client_id")
	userID := c.GetString(string(utils.ContextKeyUserID))
	var req struct {
		Metadata json.RawMessage `json:"metadata"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || len(req.Metadata) == 0 {
		c.JSON(200, Response{Code: 4003, Msg: "metadata 格式无效"})
		return
	}
	if !json.Valid(req.Metadata) || req.Metadata[0] != '{' {
		c.JSON(200, Response{Code: 4003, Msg: "metadata 必须是 JSON 对象"})
		return
	}
	var app object.App
	if err := object.Database.Where("client_id = ?", clientID).First(&app).Error; err != nil {
		c.JSON(200, Response{Code: 4001, Msg: "应用不存在"})
		return
	}
	upsertMetadata(c, app.ID, userID, string(req.Metadata))
}

// ---------- 内部辅助函数 ----------

func returnMetadata(c *gin.Context, appID, userID string) {
	var m object.UserMetadata
	err := object.Database.Where("user_id = ? AND app_id = ?", userID, appID).First(&m).Error
	if err != nil {
		c.JSON(200, Response{
			Code: 0, Msg: "ok",
			Data: gin.H{
				"user_id":    userID,
				"metadata":   json.RawMessage("{}"),
				"created_at": nil,
				"updated_at": nil,
			},
		})
		return
	}
	meta := json.RawMessage(m.Metadata)
	if meta == nil || string(meta) == "" {
		meta = json.RawMessage("{}")
	}
	c.JSON(200, Response{
		Code: 0, Msg: "ok",
		Data: gin.H{
			"user_id":    userID,
			"metadata":   meta,
			"created_at": m.CreatedAt,
			"updated_at": m.UpdatedAt,
		},
	})
}

func upsertMetadata(c *gin.Context, appID, userID, newMeta string) {
	var m object.UserMetadata
	err := object.Database.Where("user_id = ? AND app_id = ?", userID, appID).First(&m).Error
	if err != nil {
		// 创建新记录
		m = object.UserMetadata{
			ID:       uuid.NewString(),
			UserID:   userID,
			AppID:    appID,
			Metadata: newMeta,
		}
		object.Database.Create(&m)
	} else {
		// 深度合并
		merged := deepMergeJSON(m.Metadata, newMeta)
		object.Database.Model(&m).Updates(map[string]interface{}{
			"metadata": merged,
		})
		// 重新获取以获取 updated_at
		object.Database.Where("id = ?", m.ID).First(&m)
	}

	c.JSON(200, Response{
		Code: 0, Msg: "元数据已更新",
		Data: gin.H{
			"user_id":    userID,
			"metadata":   json.RawMessage(m.Metadata),
			"updated_at": m.UpdatedAt,
		},
	})
}

// deepMergeJSON 深度合并两个 JSON 对象
func deepMergeJSON(base, patch string) string {
	var baseMap map[string]interface{}
	var patchMap map[string]interface{}
	json.Unmarshal([]byte(base), &baseMap)
	json.Unmarshal([]byte(patch), &patchMap)
	if baseMap == nil {
		baseMap = make(map[string]interface{})
	}
	deepMergeMap(baseMap, patchMap)
	result, _ := json.Marshal(baseMap)
	return string(result)
}

func deepMergeMap(base, patch map[string]interface{}) {
	for k, v := range patch {
		if baseMap, ok := base[k].(map[string]interface{}); ok {
			if patchMap, ok2 := v.(map[string]interface{}); ok2 {
				deepMergeMap(baseMap, patchMap)
				continue
			}
		}
		base[k] = v
	}
}
