package review

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/KawahPutih/backend/pkg/response"
	"github.com/go-playground/validator/v10"
)

const (
	MinRating        = 1
	MaxRating        = 5
	MinCommentLength = 10
	MaxCommentLength = 1000
)

var validate = validator.New()


func Validate(s interface{}) []response.FieldError {
	err := validate.Struct(s)
	if err == nil {
		return nil
	}

	fields, ok := err.(validator.ValidationErrors)
	if !ok {
		return []response.FieldError{{
			Field:   "body",
			Message: "Data yang dikirim tidak valid.",
		}}
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

func ValidateRating(rating int) error {
	if rating < MinRating || rating > MaxRating {
		return ErrInvalidRating
	}
	return nil
}

func ValidateComment(comment string) error {
	length := utf8.RuneCountInString(NormalizeComment(comment))
	switch {
	case length == 0:
		return ErrCommentRequired
	case length < MinCommentLength:
		return ErrCommentTooShort
	case length > MaxCommentLength:
		return ErrCommentTooLong
	default:
		return nil
	}
}

func ValidateBookingID(bookingID string) error {
	if trimSpace(bookingID) == "" {
		return ErrBookingRequired
	}
	return nil
}

func (r *ModerateReviewRequest) Normalize() {
	r.Status = strings.ToUpper(trimSpace(r.Status))
}

func ValidateStatus(status string) error {
	switch strings.ToUpper(trimSpace(status)) {
	case StatusApproved, StatusRejected:
		return nil
	default:
		return ErrInvalidStatus
	}
}

func NormalizeComment(comment string) string {
	return trimSpace(comment)
}

func trimSpace(s string) string {
	return strings.TrimSpace(s)
}

func jsonFieldName(field string) string {
	switch field {
	case "DestinationID":
		return "destination_id"
	case "BookingID":
		return "booking_id"
	case "Rating":
		return "rating"
	case "Comment":
		return "comment"
	case "Status":
		return "status"
	default:
		return strings.ToLower(field)
	}
}

func friendlyMessage(fe validator.FieldError) string {
	switch fe.Field() {
	case "DestinationID":
		return "Destinasi tidak valid."
	case "BookingID":
		switch fe.Tag() {
		case "required":
			return "Booking wajib dipilih."
		default:
			return "Booking tidak valid."
		}
	case "Rating":
		switch fe.Tag() {
		case "required":
			return "Rating wajib diisi."
		case "min", "max":
			return fmt.Sprintf("Rating harus berada di antara %d sampai %d.", MinRating, MaxRating)
		default:
			return "Rating tidak valid."
		}
	case "Comment":
		switch fe.Tag() {
		case "required":
			return "Ulasan wajib diisi."
		case "min":
			return fmt.Sprintf("Ulasan minimal %d karakter.", MinCommentLength)
		case "max":
			return fmt.Sprintf("Ulasan maksimal %d karakter.", MaxCommentLength)
		default:
			return "Ulasan tidak valid."
		}
	case "Status":
		return "Status ulasan tidak valid."
	default:
		return "Nilai " + jsonFieldName(fe.Field()) + " tidak valid."
	}
}
