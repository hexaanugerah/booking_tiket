package review

import (
	"github.com/KawahPutih/backend/internal/middleware"
	"github.com/KawahPutih/backend/pkg/jwtutil"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(rg *gin.RouterGroup, h *Handler, jwtManager *jwtutil.Manager) {
	rg.GET("/reviews", h.ListApproved)
	rg.GET("/reviews/summary", h.Summary)

	requireAuth := middleware.RequireAuth(jwtManager)

	mine := rg.Group("/reviews", requireAuth)
	mine.POST("", h.Create)
	mine.GET("/mine", h.ListMine)
	mine.PUT("/:id", h.Update)
	mine.DELETE("/:id", h.Delete)

	admin := rg.Group("/admin/reviews", requireAuth, middleware.RequireRole("ADMIN", "super_admin"))
	admin.GET("", h.ListAll)
	admin.PATCH("/:id/status", h.Moderate)
}
