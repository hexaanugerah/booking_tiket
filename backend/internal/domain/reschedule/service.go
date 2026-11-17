package reschedule

import (
	"context"
	"fmt"
	"html"
	"strings"
	"time"

	"github.com/KawahPutih/backend/pkg/apperror"
	"github.com/KawahPutih/backend/pkg/mailer"
)

const maxReschedulesPerBooking = 2

type Service struct {
	repo   Repository
	mailer mailer.Mailer
}

func NewService(repo Repository, m mailer.Mailer) *Service {
	return &Service{repo: repo, mailer: m}
}

// Create menyimpan pengajuan jadwal ulang. Tanggal booking tidak langsung
// berubah; perubahan baru terjadi setelah admin menyetujui pengajuan.
func (s *Service) Create(ctx context.Context, userID, bookingID string, req CreateRequest) (*Response, error) {
	req.Normalize()
	if userID == "" {
		return nil, apperror.Unauthorized("Silakan login terlebih dahulu.")
	}

	booking, err := s.repo.FindBooking(ctx, bookingID)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	if booking == nil {
		return nil, ErrBookingNotFound
	}
	if booking.UserID != userID {
		return nil, ErrForbidden
	}
	if booking.Status != "PAID" && booking.Status != "CONFIRMED" {
		return nil, ErrBookingNotEligible
	}

	today := time.Now().Format("2006-01-02")
	if booking.VisitDate <= today {
		return nil, ErrTooLate
	}

	newDate, err := time.Parse("2006-01-02", req.NewVisitDate)
	if err != nil || len(req.NewVisitDate) != 10 {
		return nil, ErrInvalidDate
	}
	if newDate.Format("2006-01-02") <= today {
		return nil, ErrInvalidDate
	}
	if req.NewVisitDate == booking.VisitDate {
		return nil, ErrSameDate
	}

	pending, err := s.repo.HasPendingByBooking(ctx, booking.ID)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	if pending {
		return nil, ErrAlreadyPending
	}

	count, err := s.repo.CountActiveByBooking(ctx, booking.ID)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	if count >= maxReschedulesPerBooking {
		return nil, ErrLimitReached
	}

	if err := s.checkCapacity(ctx, booking, req.NewVisitDate); err != nil {
		return nil, err
	}

	item := &Reschedule{
		BookingID:    booking.ID,
		UserID:       userID,
		OldVisitDate: booking.VisitDate,
		NewVisitDate: req.NewVisitDate,
		Reason:       req.Reason,
		Status:       StatusPending,
	}
	if err := s.repo.Create(ctx, item); err != nil {
		return nil, apperror.Internal(err)
	}

	res := ToResponse(recordFromBooking(*item, booking))
	return &res, nil
}

// ListMine mengambil seluruh riwayat pengajuan milik pengunjung yang login.
func (s *Service) ListMine(ctx context.Context, userID string) ([]Response, error) {
	if userID == "" {
		return nil, apperror.Unauthorized("Silakan login terlebih dahulu.")
	}

	items, _, err := s.repo.ListByUser(ctx, userID, 0, 100)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	return ToResponseList(items), nil
}

func (s *Service) ListByBooking(ctx context.Context, userID, bookingID string) ([]Response, error) {
	booking, err := s.repo.FindBooking(ctx, bookingID)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	if booking == nil {
		return nil, ErrBookingNotFound
	}
	if booking.UserID != userID {
		return nil, ErrForbidden
	}

	items, err := s.repo.ListByBooking(ctx, bookingID)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	return ToResponseList(items), nil
}

// ListAll mengambil seluruh pengajuan untuk admin, opsional per status.
func (s *Service) ListAll(ctx context.Context, status string, page, limit int) ([]Response, int64, error) {
	status = strings.ToUpper(strings.TrimSpace(status))
	if status != "" && status != StatusPending && status != StatusApproved && status != StatusRejected {
		return nil, 0, ErrInvalidStatus
	}
	page, limit = normalizePaging(page, limit)
	items, total, err := s.repo.ListAll(ctx, status, (page-1)*limit, limit)
	if err != nil {
		return nil, 0, apperror.Internal(err)
	}
	return ToResponseList(items), total, nil
}

func (s *Service) Summary(ctx context.Context) (SummaryResponse, error) {
	row, err := s.repo.Summary(ctx)
	if err != nil {
		return SummaryResponse{}, apperror.Internal(err)
	}
	return SummaryResponse{
		Pending:  row.Pending,
		Approved: row.Approved,
		Rejected: row.Rejected,
	}, nil
}

// SetStatus memproses pengajuan jadwal ulang. Alur yang diizinkan hanya
// PENDING -> APPROVED atau PENDING -> REJECTED. Saat disetujui, tanggal booking
// dan e-tiket diperbarui lalu pengunjung menerima email notifikasi.
func (s *Service) SetStatus(ctx context.Context, id, adminID string, req UpdateStatusRequest) (*Response, error) {
	req.Normalize()
	if req.Status != StatusApproved && req.Status != StatusRejected {
		return nil, ErrInvalidStatus
	}

	record, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	if record == nil {
		return nil, ErrNotFound
	}
	if record.Status != StatusPending {
		return nil, ErrInvalidTransition
	}

	now := time.Now()
	record.AdminNote = req.AdminNote
	record.ProcessedBy = adminID
	record.ProcessedAt = &now

	if req.Status == StatusRejected {
		record.Status = StatusRejected
		if err := s.repo.Update(ctx, &record.Reschedule); err != nil {
			return nil, apperror.Internal(err)
		}
		res := ToResponse(*record)
		return &res, nil
	}

	booking, err := s.repo.FindBooking(ctx, record.BookingID)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	if booking == nil {
		return nil, ErrBookingNotFound
	}
	if booking.Status != "PAID" && booking.Status != "CONFIRMED" {
		return nil, ErrBookingNotEligible
	}
	if booking.VisitDate != record.OldVisitDate {
		return nil, ErrBookingChanged
	}
	if record.NewVisitDate <= now.Format("2006-01-02") {
		return nil, ErrNewDatePassed
	}
	if err := s.checkCapacity(ctx, booking, record.NewVisitDate); err != nil {
		return nil, err
	}

	record.Status = StatusApproved
	if err := s.repo.Approve(ctx, &record.Reschedule); err != nil {
		if err == ErrInvalidTransition || err == ErrBookingChanged {
			return nil, err
		}
		return nil, apperror.Internal(err)
	}

	record.VisitDate = record.NewVisitDate
	s.sendApprovedEmail(ctx, booking, &record.Reschedule)

	res := ToResponse(*record)
	return &res, nil
}

// checkCapacity memastikan tiket yang dipesan masih tersedia pada tanggal baru.
func (s *Service) checkCapacity(ctx context.Context, booking *BookingSnapshot, newDate string) error {
	if len(booking.Items) == 0 {
		return nil
	}

	ticketIDs := make([]string, 0, len(booking.Items))
	for _, item := range booking.Items {
		ticketIDs = append(ticketIDs, item.TicketID)
	}

	capacities, err := s.repo.TicketCapacities(ctx, ticketIDs)
	if err != nil {
		return apperror.Internal(err)
	}
	sold, err := s.repo.SoldByTicketOn(ctx, newDate, booking.ID)
	if err != nil {
		return apperror.Internal(err)
	}

	for _, item := range booking.Items {
		capacity, ok := capacities[item.TicketID]
		if !ok || !capacity.IsActive {
			return ErrTicketUnavailable
		}
		if sold[item.TicketID]+int64(item.Quantity) > int64(capacity.Capacity) {
			return ErrTicketUnavailable
		}
	}
	return nil
}

// sendApprovedEmail memberi tahu pengunjung bahwa tanggal kunjungan sudah
// berubah. Kegagalan kirim email tidak membatalkan perubahan yang sudah
// tersimpan; mailer dev mode hanya menuliskan isi email ke log.
func (s *Service) sendApprovedEmail(ctx context.Context, booking *BookingSnapshot, item *Reschedule) {
	if s.mailer == nil || booking.Email == "" {
		return
	}

	subject := fmt.Sprintf("Jadwal Kunjungan %s Disetujui", booking.BookingCode)
	body := fmt.Sprintf(
		`<p>Halo %s,</p>
<p>Pengajuan jadwal ulang untuk booking <strong>%s</strong> telah disetujui.</p>
<table cellpadding="4" style="border-collapse:collapse">
  <tr><td>Tanggal lama</td><td><strong>%s</strong></td></tr>
  <tr><td>Tanggal baru</td><td><strong>%s</strong></td></tr>
</table>
<p>E-tiket Anda sudah diperbarui dan berlaku pada tanggal baru tersebut.</p>
<p>Terima kasih.</p>`,
		html.EscapeString(booking.VisitorName),
		html.EscapeString(booking.BookingCode),
		html.EscapeString(item.OldVisitDate),
		html.EscapeString(item.NewVisitDate),
	)
	_ = s.mailer.Send(ctx, booking.Email, subject, body)
}

func normalizePaging(page, limit int) (int, int) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	return page, limit
}
