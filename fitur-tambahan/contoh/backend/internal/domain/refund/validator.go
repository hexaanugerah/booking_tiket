package refund

import (
	"strings"

	"github.com/KawahPutih/backend/pkg/response"
	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

func Validate(s interface{}) []response.FieldError {
	err := validate.Struct(s)
	if err == nil {
		return nil
	}
	fields, ok := err.(validator.ValidationErrors)
	if !ok {
		return []response.FieldError{{Field: "body", Message: "Data yang dikirim tidak valid."}}
	}
	out := make([]response.FieldError, 0, len(fields))
	for _, fe := range fields {
		out = append(out, response.FieldError{
			Field:   jsonFieldName(fe.Field()),
			Message: friendlyMessage(fe),
		})
	}
	return out
}

func jsonFieldName(field string) string {
	switch field {
	case "BookingID":
		return "booking_id"
	case "Reason":
		return "reason"
	case "BankName":
		return "bank_name"
	case "AccountName":
		return "account_name"
	case "AccountNumber":
		return "account_number"
	case "Status":
		return "status"
	case "AdminNote":
		return "admin_note"
	default:
		return strings.ToLower(field)
	}
}

func friendlyMessage(fe validator.FieldError) string {
	switch fe.Field() {
	case "BookingID":
		return "Booking tidak valid."
	case "Reason":
		switch fe.Tag() {
		case "required":
			return "Alasan refund wajib diisi."
		case "min":
			return "Alasan refund minimal 10 karakter."
		case "max":
			return "Alasan refund maksimal 500 karakter."
		default:
			return "Alasan refund tidak valid."
		}
	case "BankName":
		return "Nama bank wajib diisi (2-100 karakter)."
	case "AccountName":
		return "Nama pemilik rekening wajib diisi (2-150 karakter)."
	case "AccountNumber":
		return "Nomor rekening wajib diisi (4-50 karakter)."
	case "Status":
		return "Status refund harus APPROVED, REJECTED, atau REFUNDED."
	case "AdminNote":
		return "Catatan admin maksimal 500 karakter."
	default:
		return "Nilai " + jsonFieldName(fe.Field()) + " tidak valid."
	}
}
