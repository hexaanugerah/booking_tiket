package review

import "github.com/KawahPutih/backend/pkg/apperror"

var (
	ErrInvalidRating   = apperror.Validation("Rating harus berada di antara 1 sampai 5.")
	ErrCommentRequired = apperror.Validation("Ulasan wajib diisi.")
	ErrCommentTooShort = apperror.Validation("Ulasan minimal 10 karakter.")
	ErrCommentTooLong  = apperror.Validation("Ulasan maksimal 1000 karakter.")
	ErrInvalidStatus   = apperror.Validation("Status ulasan tidak valid.")

	ErrBookingRequired    = apperror.Validation("Booking wajib dipilih.")
	ErrBookingNotFound    = apperror.NotFound("Booking tidak ditemukan.")
	ErrBookingNotOwned    = apperror.Forbidden("Booking ini bukan milik Anda.")
	ErrBookingNotEligible = apperror.Validation("Ulasan hanya dapat diberikan untuk booking yang sudah dibayar atau digunakan.")
	ErrAlreadyReviewed    = apperror.Conflict("Booking ini sudah pernah diulas.")

	ErrDestinationNotFound = apperror.NotFound("Destinasi tidak ditemukan.")
	ErrNotFound            = apperror.NotFound("Ulasan tidak ditemukan.")
	ErrNotEditable         = apperror.Conflict("Ulasan yang sudah dimoderasi tidak dapat diubah.")
	ErrForbidden           = apperror.Forbidden("Anda tidak berhak mengubah ulasan ini.")
)
