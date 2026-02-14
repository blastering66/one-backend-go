package http

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"github.com/one-backend-go/internal/config"
	"github.com/one-backend-go/internal/domain/auth"
	"github.com/one-backend-go/internal/domain/product"
	"github.com/one-backend-go/internal/domain/user"
)

// NewRouter creates and configures the Gin engine with all routes.
func NewRouter(
	cfg *config.Config,
	jwtMgr *auth.JWTManager,
	userRepo *user.Repository,
	userHandler *user.Handler,
	authHandler *auth.Handler,
	productHandler *product.Handler,
) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)

	r := gin.New()

	// ── Global middleware ───────────────────────────────────────────────
	r.Use(RequestID())
	r.Use(StructuredLogger())
	r.Use(Recovery())

	// ── CORS ───────────────────────────────────────────────────────────
	corsConfig := cors.Config{
		AllowOrigins:     cfg.CORSAllowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "X-Request-ID"},
		ExposeHeaders:    []string{"X-Request-ID"},
		AllowCredentials: true,
	}
	r.Use(cors.New(corsConfig))

	// ── Health check ───────────────────────────────────────────────────
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// ── API v1 ─────────────────────────────────────────────────────────
	v1 := r.Group("/api/v1")
	{
		// Auth routes (public)
		authGroup := v1.Group("/auth")
		{
			authGroup.POST("/register", userHandler.Register)
			authGroup.POST("/login", authHandler.Login)
			authGroup.POST("/refresh", authHandler.Refresh)
		}

		// Product routes
		productsGroup := v1.Group("/products")
		{
			// Public
			productsGroup.GET("", productHandler.List)

			// Admin-only
			admin := productsGroup.Group("")
			admin.Use(AuthRequired(jwtMgr), AdminRequired(userRepo))
			{
				admin.POST("", productHandler.Create)
				admin.PUT("/:id", productHandler.Update)
				admin.DELETE("/:id", productHandler.Delete)
			}
		}
	}

	return r
}
