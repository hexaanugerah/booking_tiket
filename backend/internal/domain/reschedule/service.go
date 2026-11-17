package reschedule

import (
	"context"
	"time"

	"github.com/KawahPutih/backend/pkg/apperror"
)

const maxReschedulesPerBooking = 2

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// Create mengubah visit_date booking dan mencatat riwayat jadwal ulang.
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

	count, err := s.repo.CountByBooking(ctx, booking.ID)
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
		Status:       StatusApproved,
	}
	if err := s.repo.CreateAndApply(ctx, item); err != nil {
		return nil, apperror.Internal(err)
	}

	res := ToResponse(*item, booking)
	return &res, nil
}

// ListMine mengambil seluruh riwayat jadwal ulang milik pengunjung yang login.
func (s *Service) ListMine(ctx context.Context, userID string) ([]Response, error) {
	if userID == "" {
		return nil, apperror.Unauthorized("Silakan login terlebih dahulu.")
	}

	items, _, err := s.repo.ListByUser(ctx, userID, 0, 100)
	if err != nil {
		return nil, apperror.Internal(err)
	}

	out := make([]Response, 0, len(items))
	for _, item := range items {
		out = append(out, toRecordResponse(item))
	}
	return out, nil
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
	return ToResponseList(items, booking), nil
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
