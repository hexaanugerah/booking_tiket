package reschedule

import (
	"strconv"

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
	response.Created(c, "Pengajuan jadwal ulang berhasil dikirim dan menunggu persetujuan admin.", result)
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

func (h *Handler) ListAll(c *gin.Context) {
	page, limit := pageParams(c)
	items, total, err := h.service.ListAll(c.Request.Context(), c.Query("status"), page, limit)
	if err != nil {
		apperror.WriteHTTP(c, err)
		return
	}
	response.OKWithMeta(c, "Jadwal ulang retrieved", items, response.BuildMeta(page, limit, total))
}

func (h *Handler) Summary(c *gin.Context) {
	result, err := h.service.Summary(c.Request.Context())
	if err != nil {
		apperror.WriteHTTP(c, err)
		return
	}
	response.OK(c, "Ringkasan jadwal ulang retrieved", result)
}

func (h *Handler) SetStatus(c *gin.Context) {
	var req UpdateStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Format data tidak valid.", nil)
		return
	}
	req.Normalize()
	if errs := Validate(req); errs != nil {
		response.ValidationFailed(c, errs)
		return
	}

	result, err := h.service.SetStatus(
		c.Request.Context(),
		c.Param("id"),
		middleware.CurrentUserID(c),
		req,
	)
	if err != nil {
		apperror.WriteHTTP(c, err)
		return
	}
	response.OK(c, "Status jadwal ulang berhasil diperbarui.", result)
}

func pageParams(c *gin.Context) (int, int) {
	page, _ := strconv.Atoi(c.Query("page"))
	limit, _ := strconv.Atoi(c.Query("limit"))
	return normalizePaging(page, limit)
}
