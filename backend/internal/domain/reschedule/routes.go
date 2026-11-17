package reschedule

import (
	"github.com/KawahPutih/backend/internal/middleware"
	"github.com/KawahPutih/backend/pkg/jwtutil"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(rg *gin.RouterGroup, h *Handler, jwtManager *jwtutil.Manager) {
	requireAuth := middleware.RequireAuth(jwtManager)

	bookings := rg.Group("/bookings", requireAuth)
	bookings.POST("/:id/reschedule", h.Create)
	bookings.GET("/:id/reschedules", h.List)

	rg.GET("/reschedules/mine", requireAuth, h.ListMine)
}
