package refund

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
	var req CreateRefundRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Format data tidak valid.", nil)
		return
	}
	req.Normalize()
	if errs := Validate(req); errs != nil {
		response.ValidationFailed(c, errs)
		return
	}

	result, err := h.service.Create(c.Request.Context(), middleware.CurrentUserID(c), req)
	if err != nil {
		apperror.WriteHTTP(c, err)
		return
	}
	response.Created(c, "Pengajuan refund berhasil dikirim.", result)
}

func (h *Handler) CreateAdmin(c *gin.Context) {
	var req CreateAdminRefundRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Format data tidak valid.", nil)
		return
	}
	req.Normalize()
	if errs := Validate(req); errs != nil {
		response.ValidationFailed(c, errs)
		return
	}

	result, err := h.service.CreateByAdmin(c.Request.Context(), middleware.CurrentUserID(c), req)
	if err != nil {
		apperror.WriteHTTP(c, err)
		return
	}
	response.Created(c, "Refund langsung berhasil diproses.", result)
}

func (h *Handler) ListMine(c *gin.Context) {
	page, limit := pageParams(c)
	items, total, err := h.service.ListMine(c.Request.Context(), middleware.CurrentUserID(c), page, limit)
	if err != nil {
		apperror.WriteHTTP(c, err)
		return
	}
	response.OKWithMeta(c, "Refund saya retrieved", items, response.BuildMeta(page, limit, total))
}

func (h *Handler) ListAll(c *gin.Context) {
	page, limit := pageParams(c)
	items, total, err := h.service.ListAll(c.Request.Context(), c.Query("status"), page, limit)
	if err != nil {
		apperror.WriteHTTP(c, err)
		return
	}
	response.OKWithMeta(c, "Refund retrieved", items, response.BuildMeta(page, limit, total))
}

func (h *Handler) Summary(c *gin.Context) {
	result, err := h.service.Summary(c.Request.Context())
	if err != nil {
		apperror.WriteHTTP(c, err)
		return
	}
	response.OK(c, "Ringkasan refund retrieved", result)
}

func (h *Handler) SetStatus(c *gin.Context) {
	var req UpdateRefundStatusRequest
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
	response.OK(c, "Status refund berhasil diperbarui.", result)
}

func pageParams(c *gin.Context) (int, int) {
	page, _ := strconv.Atoi(c.Query("page"))
	limit, _ := strconv.Atoi(c.Query("limit"))
	return normalizePaging(page, limit)
}
