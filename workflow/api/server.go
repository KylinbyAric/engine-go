package api

import (
	"io/fs"
	"net/http"

	"github.com/engine-go/workflow/api/handler"
	"github.com/gin-gonic/gin"
)

// NewEngine 构造 gin 引擎：静态资源 + /api 路由。
// staticFS 应当能直接读到 index.html / vis-network.min.js 等文件（即已剥离 public/static 前缀）。
func NewEngine(staticFS fs.FS) *gin.Engine {
	r := gin.Default()
	r.Use(corsMiddleware())

	if staticFS != nil {
		r.StaticFS("/static", http.FS(staticFS))
		r.GET("/", func(c *gin.Context) {
			data, err := fs.ReadFile(staticFS, "index.html")
			if err != nil {
				c.String(http.StatusNotFound, "index.html not found: %v", err)
				return
			}
			c.Data(http.StatusOK, "text/html; charset=utf-8", data)
		})
	}

	r.GET("/healthz", func(c *gin.Context) { c.String(200, "ok") })

	gh := handler.NewGraphHandler()
	apiGroup := r.Group("/api")
	{
		g := apiGroup.Group("/graph")
		g.GET("", gh.List)
		g.GET("/:id", gh.Get)
		g.POST("", gh.Create)
		g.PUT("/:id", gh.Update)
		g.DELETE("/:id", gh.Delete)
	}
	// 兼容 index.html 写死的接口路径
	r.POST("/onboard/api/v1/workflow/graph/detail", gh.Detail)
	return r
}

func Run(addr string, staticFS fs.FS) error {
	return NewEngine(staticFS).Run(addr)
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
		c.Header("Access-Control-Allow-Headers", "*")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}
