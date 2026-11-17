package refund

import (
	"github.com/KawahPutih/backend/internal/middleware"
	"github.com/KawahPutih/backend/pkg/jwtutil"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(rg *gin.RouterGroup, h *Handler, jwtManager *jwtutil.Manager) {
	requireAuth := middleware.RequireAuth(jwtManager)

	user := rg.Group("/refunds", requireAuth)
	user.POST("", h.Create)
	user.GET("/mine", h.ListMine)

	admin := rg.Group("/admin/refunds", requireAuth, middleware.RequireRole("ADMIN", "super_admin"))
	admin.GET("", h.ListAll)
	admin.GET("/summary", h.Summary)
	admin.PATCH("/:id/status", h.SetStatus)
}
