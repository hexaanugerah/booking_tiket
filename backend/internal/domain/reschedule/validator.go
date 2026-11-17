package reschedule

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
	case "NewVisitDate":
		return "new_visit_date"
	case "Reason":
		return "reason"
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
	case "NewVisitDate":
		return "Tanggal kunjungan baru wajib diisi dengan format YYYY-MM-DD."
	case "Reason":
		switch fe.Tag() {
		case "required":
			return "Alasan jadwal ulang wajib diisi."
		case "min":
			return "Alasan jadwal ulang minimal 10 karakter."
		case "max":
			return "Alasan jadwal ulang maksimal 500 karakter."
		default:
			return "Alasan jadwal ulang tidak valid."
		}
	case "Status":
		return "Status jadwal ulang harus APPROVED atau REJECTED."
	case "AdminNote":
		return "Catatan admin maksimal 500 karakter."
	default:
		return "Nilai " + jsonFieldName(fe.Field()) + " tidak valid."
	}
}
