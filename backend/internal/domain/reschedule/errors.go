package reschedule

import "github.com/KawahPutih/backend/pkg/apperror"

var (
	ErrBookingNotFound    = apperror.NotFound("Booking tidak ditemukan.").WithCode("RSCH_001")
	ErrForbidden          = apperror.Forbidden("Booking ini bukan milik Anda.").WithCode("RSCH_002")
	ErrBookingNotEligible = apperror.Conflict("Jadwal ulang hanya berlaku untuk booking yang sudah dibayar atau dikonfirmasi.").WithCode("RSCH_003")
	ErrTooLate            = apperror.Conflict("Jadwal ulang hanya dapat dilakukan sebelum tanggal kunjungan.").WithCode("RSCH_004")
	ErrInvalidDate        = apperror.Validation("Tanggal kunjungan baru tidak valid.").WithCode("RSCH_005")
	ErrSameDate           = apperror.Conflict("Tanggal kunjungan baru sama dengan tanggal saat ini.").WithCode("RSCH_006")
	ErrLimitReached       = apperror.Conflict("Batas maksimal 2 kali jadwal ulang sudah tercapai.").WithCode("RSCH_007")
	ErrTicketUnavailable  = apperror.Conflict("Tiket pada tanggal baru tidak tersedia.").WithCode("RSCH_008")
	ErrNotFound           = apperror.NotFound("Pengajuan jadwal ulang tidak ditemukan.").WithCode("RSCH_009")
	ErrInvalidStatus      = apperror.Validation("Status jadwal ulang tidak valid.").WithCode("RSCH_010")
	ErrInvalidTransition  = apperror.Conflict("Status jadwal ulang tidak dapat diubah ke tahap tersebut.").WithCode("RSCH_011")
	ErrAlreadyPending     = apperror.Conflict("Booking ini masih memiliki pengajuan jadwal ulang yang menunggu persetujuan.").WithCode("RSCH_012")
	ErrBookingChanged     = apperror.Conflict("Tanggal booking sudah berubah. Muat ulang data lalu coba kembali.").WithCode("RSCH_013")
	ErrNewDatePassed      = apperror.Conflict("Tanggal kunjungan baru sudah lewat sehingga tidak dapat disetujui.").WithCode("RSCH_014")
)
