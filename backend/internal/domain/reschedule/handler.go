package reschedule

import (
	"github.com/KawahPutih/backend/internal/middleware"
	"github.com/KawahPutih/backend/pkg/apperror"
	"github.com/KawahPutih/backend/pkg/response"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Create(c *gin.Context) {
	var req CreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Format data tidak valid.", nil)
		return
	}
	req.Normalize()
	if errs := Validate(req); errs != nil {
		response.ValidationFailed(c, errs)
		return
	}

	result, err := h.service.Create(
		c.Request.Context(),
		middleware.CurrentUserID(c),
		c.Param("id"),
		req,
	)
	if err != nil {
		apperror.WriteHTTP(c, err)
		return
	}
	response.Created(c, "Jadwal kunjungan berhasil diubah.", result)
}

func (h *Handler) List(c *gin.Context) {
	items, err := h.service.ListByBooking(
		c.Request.Context(),
		middleware.CurrentUserID(c),
		c.Param("id"),
	)
	if err != nil {
		apperror.WriteHTTP(c, err)
		return
	}
	response.OK(c, "Riwayat jadwal ulang retrieved", items)
}

func (h *Handler) ListMine(c *gin.Context) {
	items, err := h.service.ListMine(
		c.Request.Context(),
		middleware.CurrentUserID(c),
	)
	if err != nil {
		apperror.WriteHTTP(c, err)
		return
	}
	response.OK(c, "Riwayat jadwal ulang retrieved", items)
}
