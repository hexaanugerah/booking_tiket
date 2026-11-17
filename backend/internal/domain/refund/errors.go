package refund

import "github.com/KawahPutih/backend/pkg/apperror"

var (
	ErrBookingNotFound    = apperror.NotFound("Booking tidak ditemukan.").WithCode("REFUND_001")
	ErrNotFound           = apperror.NotFound("Pengajuan refund tidak ditemukan.").WithCode("REFUND_002")
	ErrForbidden          = apperror.Forbidden("Anda tidak berhak mengakses pengajuan refund ini.").WithCode("REFUND_003")
	ErrBookingNotEligible = apperror.Conflict("Refund hanya dapat diajukan untuk booking yang sudah dibayar atau dikonfirmasi.").WithCode("REFUND_004")
	ErrVisitDatePassed    = apperror.Conflict("Refund hanya dapat diajukan paling lambat H-1 sebelum tanggal kunjungan.").WithCode("REFUND_005")
	ErrAlreadyRequested   = apperror.Conflict("Booking ini sudah memiliki pengajuan refund yang sedang diproses.").WithCode("REFUND_006")
	ErrInvalidStatus      = apperror.Validation("Status refund tidak valid.").WithCode("REFUND_007")
	ErrInvalidTransition  = apperror.Conflict("Status refund tidak dapat diubah ke tahap tersebut.").WithCode("REFUND_008")
)
