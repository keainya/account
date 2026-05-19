package router

import (
	"embed"
	"io/fs"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/keainya/service_temp/service"
)

func InitRouter(webEmbed embed.FS) *gin.Engine {
	webRoot, err := fs.Sub(webEmbed, "web")
	if err != nil {
		panic("embedded web directory not found: " + err.Error())
	}

	// 为每个子目录创建子文件系统，匹配 StaticFS 的路径剥离逻辑
	cssFS, _ := fs.Sub(webRoot, "css")
	jsFS, _ := fs.Sub(webRoot, "js")

	// 预加载 index.html，用于 SPA fallback
	indexHTML, err := fs.ReadFile(webRoot, "index.html")
	if err != nil {
		panic("embedded index.html not found: " + err.Error())
	}

	r := gin.Default()

	// ---- CORS 中间件 ----
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// ---- Session 中间件 ----
	store := cookie.NewStore([]byte("change-me-to-a-secure-random-key"))
	r.Use(sessions.Sessions("account_session", store))

	// ========== 公共路由 ==========
	r.GET("/status", service.Status)

	// OAuth 授权页面（浏览器访问）
	r.GET("/oauth/login", service.OAuthLoginPage)

	// ========== 认证 API ==========
	auth := r.Group("/api/auth")
	{
		auth.POST("/register", service.Register)
		auth.POST("/login", service.Login)
		auth.POST("/logout", service.Logout)
		auth.GET("/me", service.AuthRequired(), service.Me)
	}

	// ========== OAuth 2.0 API ==========
	// authorize 内部会判断登录状态，不强制要求已登录
	r.GET("/oauth/authorize", service.OAuthAuthorize)
	r.POST("/oauth/token", service.OAuthToken)
	r.GET("/oauth/userinfo", service.BearerAuth(), service.OAuthUserinfo)

	// ========== 元数据 API (Bearer Token) ==========
	metadata := r.Group("/api/apps")
	metadata.Use(service.BearerAuth())
	{
		metadata.GET("/:client_id/users/:user_id/metadata", service.MetadataGetUser)
		metadata.PUT("/:client_id/users/:user_id/metadata", service.MetadataPutUser)
		metadata.GET("/:client_id/metadata", service.MetadataBatchGet)
	}

	// ========== 元数据 API (Session) ==========
	myMeta := r.Group("/api/apps")
	myMeta.Use(service.AuthRequired())
	{
		myMeta.GET("/:client_id/my-metadata", service.MetadataGetMy)
		myMeta.PUT("/:client_id/my-metadata", service.MetadataPutMy)
	}

	// ========== 管理员 API ==========
	admin := r.Group("/api/admin")
	admin.Use(service.AuthRequired(), service.AdminRequired())
	{
		// 用户管理
		admin.GET("/users", service.AdminListUsers)
		admin.PUT("/users/:user_id/promote", service.AdminPromoteUser)
		admin.PUT("/users/:user_id/demote", service.AdminDemoteUser)

		// 应用管理
		admin.POST("/apps", service.AdminCreateApp)
		admin.GET("/apps", service.AdminListApps)
		admin.GET("/apps/:app_id", service.AdminGetApp)
		admin.PUT("/apps/:app_id", service.AdminUpdateApp)
		admin.POST("/apps/:app_id/reset-secret", service.AdminResetAppSecret)
		admin.DELETE("/apps/:app_id", service.AdminDeleteApp)
	}

	// ========== 前端静态文件 & SPA fallback ==========
	// 注意：必须在所有 API 路由之后，避免路由冲突
	r.StaticFS("/css", http.FS(cssFS))
	r.StaticFS("/js", http.FS(jsFS))
	r.GET("/favicon.ico", func(c *gin.Context) {
		c.FileFromFS("favicon.ico", http.FS(webRoot))
	})

	// SPA fallback：未匹配的任何路径都返回 index.html
	r.NoRoute(func(c *gin.Context) {
		c.Data(http.StatusOK, "text/html; charset=utf-8", indexHTML)
	})

	return r
}
