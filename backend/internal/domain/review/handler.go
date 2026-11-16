package review

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
	var req CreateReviewRequest
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
	response.Created(c, "Ulasan berhasil dikirim.", ToResponse(result))
}

func (h *Handler) Update(c *gin.Context) {
	var req UpdateReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Format data tidak valid.", nil)
		return
	}
	req.Normalize()
	if errs := Validate(req); errs != nil {
		response.ValidationFailed(c, errs)
		return
	}

	result, err := h.service.Update(c.Request.Context(), middleware.CurrentUserID(c), c.Param("id"), req)
	if err != nil {
		apperror.WriteHTTP(c, err)
		return
	}
	response.OK(c, "Ulasan berhasil diperbarui.", ToResponse(result))
}

func (h *Handler) Delete(c *gin.Context) {
	if err := h.service.Delete(c.Request.Context(), middleware.CurrentUserID(c), c.Param("id")); err != nil {
		apperror.WriteHTTP(c, err)
		return
	}
	response.OK(c, "Ulasan berhasil dihapus.", nil)
}

func (h *Handler) ListMine(c *gin.Context) {
	page, limit := pageParams(c)
	items, total, err := h.service.ListMine(c.Request.Context(), middleware.CurrentUserID(c), page, limit)
	if err != nil {
		apperror.WriteHTTP(c, err)
		return
	}
	response.OKWithMeta(c, "Ulasan saya retrieved", ToResponseList(items), response.BuildMeta(page, limit, total))
}

func (h *Handler) ListAll(c *gin.Context) {
	page, limit := pageParams(c)
	items, total, err := h.service.ListAll(c.Request.Context(), c.Query("status"), page, limit)
	if err != nil {
		apperror.WriteHTTP(c, err)
		return
	}
	response.OKWithMeta(c, "Ulasan retrieved", ToResponseList(items), response.BuildMeta(page, limit, total))
}

func (h *Handler) ListApproved(c *gin.Context) {
	page, limit := pageParams(c)
	items, total, err := h.service.ListApproved(c.Request.Context(), page, limit)
	if err != nil {
		apperror.WriteHTTP(c, err)
		return
	}
	response.OKWithMeta(c, "Ulasan retrieved", ToResponseList(items), response.BuildMeta(page, limit, total))
}

func (h *Handler) Summary(c *gin.Context) {
	average, total, err := h.service.Summary(c.Request.Context())
	if err != nil {
		apperror.WriteHTTP(c, err)
		return
	}
	response.OK(c, "Ringkasan rating retrieved", SummaryResponse{
		AverageRating: average,
		TotalReviews:  total,
	})
}

func (h *Handler) Moderate(c *gin.Context) {
	var req ModerateReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Format data tidak valid.", nil)
		return
	}
	req.Normalize()
	if errs := Validate(req); errs != nil {
		response.ValidationFailed(c, errs)
		return
	}

	result, err := h.service.Moderate(c.Request.Context(), c.Param("id"), req)
	if err != nil {
		apperror.WriteHTTP(c, err)
		return
	}
	response.OK(c, "Status ulasan berhasil diperbarui.", ToResponse(result))
}

func pageParams(c *gin.Context) (int, int) {
	page, _ := strconv.Atoi(c.Query("page"))
	limit, _ := strconv.Atoi(c.Query("limit"))
	return normalizePaging(page, limit)
}
