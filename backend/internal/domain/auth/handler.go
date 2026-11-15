package auth

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/KawahPutih/backend/internal/middleware"
	"github.com/KawahPutih/backend/pkg/apperror"
	"github.com/KawahPutih/backend/pkg/response"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	usecase *Usecase
}

func NewHandler(usecase *Usecase) *Handler {
	return &Handler{usecase: usecase}
}

func (h *Handler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body", nil)
		return
	}
	if errs := Validate(req); errs != nil {
		response.ValidationFailed(c, errs)
		return
	}
	if errs := ValidatePasswordStrength("password", req.Password); errs != nil {
		response.ValidationFailed(c, errs)
		return
	}
	result, err := h.usecase.Register(c.Request.Context(), req)
	if err != nil {
		apperror.WriteHTTP(c, err)
		return
	}
	response.OK(c, "account created successfully", result)
}

func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body", nil)
		return
	}
	if errs := Validate(req); errs != nil {
		response.ValidationFailed(c, errs)
		return
	}
	result, err := h.usecase.Login(c.Request.Context(), req, c.Request.UserAgent(), c.ClientIP())
	if err != nil {
		apperror.WriteHTTP(c, err)
		return
	}
	response.OK(c, "login successful", result)
}

func (h *Handler) Me(c *gin.Context) {
	resp, err := h.usecase.Me(c.Request.Context(), middleware.CurrentUserID(c))
	if err != nil {
		apperror.WriteHTTP(c, err)
		return
	}
	response.OK(c, "profile retrieved", resp)
}

func (h *Handler) UpdateProfile(c *gin.Context) {
	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body", nil)
		return
	}
	if errs := Validate(req); errs != nil {
		response.ValidationFailed(c, errs)
		return
	}
	resp, err := h.usecase.UpdateProfile(c.Request.Context(), middleware.CurrentUserID(c), req)
	if err != nil {
		apperror.WriteHTTP(c, err)
		return
	}
	response.OK(c, "profile updated", resp)
}

// ChangePassword memungkinkan pemilik akun (role apa pun) mengganti password-nya sendiri setelah memverifikasi password lama.
func (h *Handler) ChangePassword(c *gin.Context) {
	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body", nil)
		return
	}
	if errs := Validate(req); errs != nil {
		response.ValidationFailed(c, errs)
		return
	}
	if errs := ValidatePasswordStrength("new_password", req.NewPassword); errs != nil {
		response.ValidationFailed(c, errs)
		return
	}
	if err := h.usecase.ChangePassword(c.Request.Context(), middleware.CurrentUserID(c), req); err != nil {
		apperror.WriteHTTP(c, err)
		return
	}
	response.OK(c, "password changed successfully, please log in again", nil)
}

func (h *Handler) Refresh(c *gin.Context) {
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body", nil)
		return
	}
	result, err := h.usecase.Refresh(c.Request.Context(), req.RefreshToken)
	if err != nil {
		apperror.WriteHTTP(c, err)
		return
	}
	response.OK(c, "token refreshed", result)
}

func (h *Handler) Logout(c *gin.Context) {
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body", nil)
		return
	}
	if err := h.usecase.Logout(c.Request.Context(), req.RefreshToken); err != nil {
		apperror.WriteHTTP(c, err)
		return
	}
	response.OK(c, "logged out successfully", nil)
}

func (h *Handler) RequestEmailVerification(c *gin.Context) {
	var req RequestEmailVerificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body", nil)
		return
	}
	if errs := Validate(req); errs != nil {
		response.ValidationFailed(c, errs)
		return
	}
	if err := h.usecase.RequestEmailVerification(c.Request.Context(), req.Email); err != nil {
		apperror.WriteHTTP(c, err)
		return
	}
	response.OK(c, "if the email exists, a verification link has been sent", nil)
}

func (h *Handler) VerifyEmail(c *gin.Context) {
	var req VerifyEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body", nil)
		return
	}
	if err := h.usecase.VerifyEmail(c.Request.Context(), req.Token); err != nil {
		apperror.WriteHTTP(c, err)
		return
	}
	response.OK(c, "email verified successfully", nil)
}

func (h *Handler) ForgotPassword(c *gin.Context) {
	var req ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body", nil)
		return
	}
	if errs := Validate(req); errs != nil {
		response.ValidationFailed(c, errs)
		return
	}
	if err := h.usecase.ForgotPassword(c.Request.Context(), req.Email); err != nil {
		apperror.WriteHTTP(c, err)
		return
	}
	response.OK(c, "if the email exists, a reset link has been sent", nil)
}

func (h *Handler) ResetPassword(c *gin.Context) {
	var req ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body", nil)
		return
	}
	if errs := Validate(req); errs != nil {
		response.ValidationFailed(c, errs)
		return
	}
	if errs := ValidatePasswordStrength("new_password", req.NewPassword); errs != nil {
		response.ValidationFailed(c, errs)
		return
	}
	if err := h.usecase.ResetPassword(c.Request.Context(), req.Token, req.NewPassword); err != nil {
		apperror.WriteHTTP(c, err)
		return
	}
	response.OK(c, "password reset successfully, please log in again", nil)
}

func (h *Handler) GoogleLogin(c *gin.Context) {
	url, err := h.usecase.GoogleAuthURL("KawahPutih-oauth-state")
	if err != nil {
		apperror.WriteHTTP(c, err)
		return
	}
	c.Redirect(http.StatusTemporaryRedirect, url)
}

// GoogleCallback menukar kode OAuth menjadi sesi, lalu mengembalikan pengguna
// ke halaman callback frontend (bukan respons JSON mentah) membawa token via
// query — frontend di sana menyimpan sesi seperti login biasa.
func (h *Handler) GoogleCallback(c *gin.Context) {
	base := strings.TrimRight(h.usecase.FrontendURL(), "/")
	redirect := func(params url.Values) {
		c.Redirect(http.StatusTemporaryRedirect, base+"/auth/google/callback?"+params.Encode())
	}
	failure := func(msg string) {
		redirect(url.Values{"error": {msg}})
	}

	// Google mengirim ?error= bila pengguna membatalkan/menolak persetujuan.
	if c.Query("error") != "" {
		failure("Login Google dibatalkan atau ditolak.")
		return
	}
	code := c.Query("code")
	if code == "" {
		failure("Kode otorisasi Google tidak diterima.")
		return
	}
	result, err := h.usecase.GoogleLogin(c.Request.Context(), code)
	if err != nil {
		msg := "Login Google gagal. Silakan coba lagi."
		if appErr, ok := apperror.As(err); ok {
			msg = appErr.Error()
		}
		failure(msg)
		return
	}
	userJSON, err := json.Marshal(result.User)
	if err != nil {
		failure("Gagal menyiapkan data sesi.")
		return
	}
	redirect(url.Values{
		"access_token":  {result.AccessToken},
		"refresh_token": {result.RefreshToken},
		"expires_at":    {strconv.FormatInt(result.ExpiresAt, 10)},
		"user":          {string(userJSON)},
	})
}
