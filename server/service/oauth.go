package service

import (
	"encoding/json"
	"net/url"
	"strings"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/keainya/service_temp/object"
	"github.com/keainya/service_temp/utils"
)

// OAuthAuthorize 授权端点
func OAuthAuthorize(c *gin.Context) {
	responseType := c.Query("response_type")
	clientID := c.Query("client_id")
	redirectURI := c.Query("redirect_uri")
	scope := c.Query("scope")
	state := c.Query("state")

	// 校验参数
	if responseType != "code" {
		c.JSON(200, Response{Code: 3006, Msg: "grant_type 不支持"})
		return
	}
	if state == "" {
		c.JSON(200, Response{Code: 3002, Msg: "state 不能为空"})
		return
	}

	// 验证应用
	var app object.App
	if err := object.Database.Where("client_id = ?", clientID).First(&app).Error; err != nil {
		c.JSON(200, Response{Code: 3001, Msg: "无效的 client_id"})
		return
	}

	// 验证 redirect_uri
	var uris []string
	if err := json.Unmarshal([]byte(app.RedirectURIs), &uris); err != nil || !containsURI(uris, redirectURI) {
		c.JSON(200, Response{Code: 3002, Msg: "redirect_uri 不匹配"})
		return
	}

	// 检查用户是否登录
	session := sessions.Default(c)
	userID := session.Get("user_id")
	if userID == nil {
		// 未登录，返回简单 JSON（API 模式下），也可以重定向到登录页
		c.JSON(200, Response{Code: 2001, Msg: "请先登录", Data: gin.H{
			"authorize_url": c.Request.URL.String(),
		}})
		return
	}

	// 生成授权码
	code, err := utils.GenerateRandomString(32)
	if err != nil {
		c.JSON(200, Response{Code: 9000, Msg: "服务器内部错误"})
		return
	}

	oauthCode := object.OAuthCode{
		Code:        code,
		UserID:      userID.(string),
		ClientID:    clientID,
		RedirectURI: redirectURI,
		Scope:       scope,
		ExpiresAt:   time.Now().Add(10 * time.Minute),
	}
	if err := object.Database.Create(&oauthCode).Error; err != nil {
		c.JSON(200, Response{Code: 9001, Msg: "数据库错误"})
		return
	}

	// 重定向回应用
	redirectURL, _ := url.Parse(redirectURI)
	q := redirectURL.Query()
	q.Set("code", code)
	q.Set("state", state)
	redirectURL.RawQuery = q.Encode()
	c.Redirect(302, redirectURL.String())
}

// OAuthToken 令牌端点
func OAuthToken(c *gin.Context) {
	grantType := c.PostForm("grant_type")

	if grantType == "authorization_code" {
		handleAuthorizationCode(c)
	} else if grantType == "refresh_token" {
		handleRefreshToken(c)
	} else {
		c.JSON(200, Response{Code: 3006, Msg: "grant_type 不支持"})
	}
}

func handleAuthorizationCode(c *gin.Context) {
	codeStr := c.PostForm("code")
	redirectURI := c.PostForm("redirect_uri")
	clientID := c.PostForm("client_id")
	clientSecret := c.PostForm("client_secret")

	// 验证应用凭据
	var app object.App
	if err := object.Database.Where("client_id = ? AND client_secret = ?", clientID, clientSecret).First(&app).Error; err != nil {
		c.JSON(200, Response{Code: 3005, Msg: "client_secret 错误"})
		return
	}

	// 验证授权码
	var oauthCode object.OAuthCode
	if err := object.Database.Where("code = ?", codeStr).First(&oauthCode).Error; err != nil {
		c.JSON(200, Response{Code: 3004, Msg: "授权码无效"})
		return
	}

	if oauthCode.Used {
		c.JSON(200, Response{Code: 3004, Msg: "授权码已使用"})
		return
	}
	if time.Now().After(oauthCode.ExpiresAt) {
		c.JSON(200, Response{Code: 3004, Msg: "授权码已过期"})
		return
	}
	if oauthCode.ClientID != clientID || oauthCode.RedirectURI != redirectURI {
		c.JSON(200, Response{Code: 3004, Msg: "redirect_uri 不匹配"})
		return
	}

	// 标记授权码已使用
	object.Database.Model(&oauthCode).Update("used", true)

	// 获取用户
	var user object.User
	if err := object.Database.Where("id = ?", oauthCode.UserID).First(&user).Error; err != nil {
		c.JSON(200, Response{Code: 3004, Msg: "用户不存在"})
		return
	}

	// 生成 tokens
	accessToken, err := utils.GenerateAccessToken(user.ID, user.Username, user.Role, clientID)
	if err != nil {
		c.JSON(200, Response{Code: 9000, Msg: "服务器内部错误"})
		return
	}
	refreshTokenStr, err := utils.GenerateRandomString(32)
	if err != nil {
		c.JSON(200, Response{Code: 9000, Msg: "服务器内部错误"})
		return
	}

	refreshToken := object.RefreshToken{
		Token:     refreshTokenStr,
		UserID:    user.ID,
		ClientID:  clientID,
		ExpiresAt: time.Now().Add(utils.RefreshTokenTTL),
	}
	object.Database.Create(&refreshToken)

	c.JSON(200, gin.H{
		"access_token":  accessToken,
		"token_type":    "Bearer",
		"expires_in":    3600,
		"refresh_token": refreshTokenStr,
	})
}

func handleRefreshToken(c *gin.Context) {
	refreshTokenStr := c.PostForm("refresh_token")
	clientID := c.PostForm("client_id")
	clientSecret := c.PostForm("client_secret")

	// 验证应用
	var app object.App
	if err := object.Database.Where("client_id = ? AND client_secret = ?", clientID, clientSecret).First(&app).Error; err != nil {
		c.JSON(200, Response{Code: 3005, Msg: "client_secret 错误"})
		return
	}

	// 验证 refresh_token
	var rt object.RefreshToken
	if err := object.Database.Where("token = ?", refreshTokenStr).First(&rt).Error; err != nil {
		c.JSON(200, Response{Code: 3007, Msg: "refresh_token 无效或已过期"})
		return
	}
	if rt.Revoked || time.Now().After(rt.ExpiresAt) {
		c.JSON(200, Response{Code: 3007, Msg: "refresh_token 无效或已过期"})
		return
	}

	// 废除旧 token
	object.Database.Model(&rt).Update("revoked", true)

	// 获取用户
	var user object.User
	if err := object.Database.Where("id = ?", rt.UserID).First(&user).Error; err != nil {
		c.JSON(200, Response{Code: 3007, Msg: "用户不存在"})
		return
	}

	// 生成新 tokens
	accessToken, err := utils.GenerateAccessToken(user.ID, user.Username, user.Role, clientID)
	if err != nil {
		c.JSON(200, Response{Code: 9000, Msg: "服务器内部错误"})
		return
	}
	newRefreshTokenStr, err := utils.GenerateRandomString(32)
	if err != nil {
		c.JSON(200, Response{Code: 9000, Msg: "服务器内部错误"})
		return
	}

	newRefreshToken := object.RefreshToken{
		Token:     newRefreshTokenStr,
		UserID:    user.ID,
		ClientID:  clientID,
		ExpiresAt: time.Now().Add(utils.RefreshTokenTTL),
	}
	object.Database.Create(&newRefreshToken)

	c.JSON(200, gin.H{
		"access_token":  accessToken,
		"token_type":    "Bearer",
		"expires_in":    3600,
		"refresh_token": newRefreshTokenStr,
	})
}

// OAuthUserinfo 用户信息端点
func OAuthUserinfo(c *gin.Context) {
	userID := c.GetString(string(utils.ContextKeyUserID))

	var user object.User
	if err := object.Database.Where("id = ?", userID).First(&user).Error; err != nil {
		c.JSON(200, Response{Code: 3008, Msg: "access_token 无效或已过期"})
		return
	}

	c.JSON(200, gin.H{
		"sub":      user.ID,
		"username": user.Username,
		"email":    user.Email,
		"role":     user.Role,
	})
}

// containsURI 检查 redirect_uri 是否在允许列表中
func containsURI(uris []string, target string) bool {
	for _, u := range uris {
		if u == target {
			return true
		}
	}
	return false
}

// OAuthLoginPage OAuth 登录页面（简易 HTML）
func OAuthLoginPage(c *gin.Context) {
	clientID := c.Query("client_id")
	redirectURI := c.Query("redirect_uri")
	state := c.Query("state")
	responseType := c.Query("response_type")

	if clientID == "" || redirectURI == "" || state == "" || responseType != "code" {
		c.String(400, "参数不完整")
		return
	}

	// 验证应用
	var app object.App
	if err := object.Database.Where("client_id = ?", clientID).First(&app).Error; err != nil {
		c.String(400, "无效的应用")
		return
	}

	html := `<!DOCTYPE html>
<html lang="zh">
<head>
	<meta charset="UTF-8">
	<title>授权登录 - Account</title>
	<meta name="viewport" content="width=device-width, initial-scale=1">
	<style>
		* { margin: 0; padding: 0; box-sizing: border-box; }
		body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif; background: #f5f5f5; display: flex; justify-content: center; align-items: center; min-height: 100vh; }
		.card { background: #fff; border-radius: 12px; box-shadow: 0 2px 12px rgba(0,0,0,0.08); padding: 32px; width: 100%; max-width: 400px; }
		h2 { margin-bottom: 8px; font-size: 20px; }
		.sub { color: #666; font-size: 14px; margin-bottom: 24px; }
		.form-group { margin-bottom: 16px; }
		label { display: block; margin-bottom: 4px; font-size: 14px; font-weight: 500; }
		input { width: 100%; padding: 10px 12px; border: 1px solid #ddd; border-radius: 8px; font-size: 14px; outline: none; transition: border-color .2s; }
		input:focus { border-color: #4f46e5; }
		.btn { width: 100%; padding: 10px; border: none; border-radius: 8px; font-size: 15px; font-weight: 500; cursor: pointer; margin-top: 8px; }
		.btn-primary { background: #4f46e5; color: #fff; }
		.btn-primary:hover { background: #4338ca; }
		.msg { color: #e53e3e; font-size: 13px; margin-top: 8px; display: none; }
</style>
</head>
<body>
	<div class="card">
		<h2>授权登录</h2>
		<p class="sub">应用 <strong>` + escapeHTML(app.Name) + `</strong> 请求访问您的账户</p>
		<form id="loginForm">
			<div class="form-group">
				<label>用户名</label>
				<input type="text" name="username" required>
			</div>
			<div class="form-group">
				<label>密码</label>
				<input type="password" name="password" required>
			</div>
			<button type="submit" class="btn btn-primary">登录并授权</button>
			<p id="msg" class="msg"></p>
		</form>
	</div>
	<script>
		const params = new URLSearchParams(window.location.search);
		document.getElementById('loginForm').onsubmit = async (e) => {
			e.preventDefault();
			const fd = new FormData(e.target);
			const res = await fetch('/api/auth/login', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				credentials: 'same-origin',
				body: JSON.stringify({ username: fd.get('username'), password: fd.get('password') })
			});
			const data = await res.json();
			if (data.code === 0) {
				window.location.href = '/oauth/authorize?' + params.toString();
			} else {
				document.getElementById('msg').style.display = 'block';
				document.getElementById('msg').textContent = data.msg;
			}
		};
</script>
</body>
</html>`
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(200, html)
}

func escapeHTML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	return s
}



