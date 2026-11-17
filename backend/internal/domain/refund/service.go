package refund

import (
	"context"
	"strings"
	"time"

	"github.com/KawahPutih/backend/pkg/apperror"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// Create membuat pengajuan refund baru. Nominal selalu diambil dari total
// booking di server, bukan dari input pengguna.
func (s *Service) Create(ctx context.Context, userID string, req CreateRefundRequest) (*Response, error) {
	req.Normalize()
	if userID == "" {
		return nil, apperror.Unauthorized("Silakan login terlebih dahulu.")
	}

	booking, err := s.repo.FindBooking(ctx, req.BookingID)
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
	if booking.VisitDate <= time.Now().Format("2006-01-02") {
		return nil, ErrVisitDatePassed
	}

	existing, err := s.repo.FindActiveByBooking(ctx, req.BookingID)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	if existing != nil {
		return nil, ErrAlreadyRequested
	}

	netAmount, deductionAmount, deductionPercent := calculateAmounts(booking.Total, InitiatedByUser)

	destinationType := req.DestinationType
	if destinationType == "" {
		destinationType = DestinationTypeBank
	}

	refund := &Refund{
		BookingID:        booking.ID,
		UserID:           userID,
		InitiatedBy:      InitiatedByUser,
		DestinationType:  destinationType,
		Amount:           netAmount,
		OriginalAmount:   booking.Total,
		DeductionPercent: deductionPercent,
		DeductionAmount:  deductionAmount,
		Reason:           req.Reason,
		BankName:         req.BankName,
		AccountName:      req.AccountName,
		AccountNumber:    req.AccountNumber,
		Status:           StatusPending,
	}
	if err := s.repo.Create(ctx, refund); err != nil {
		return nil, apperror.Internal(err)
	}

	record := Record{
		Refund:        *refund,
		BookingCode:   booking.BookingCode,
		VisitorName:   booking.VisitorName,
		Email:         booking.Email,
		VisitDate:     booking.VisitDate,
		BookingStatus: booking.Status,
	}
	res := ToResponse(record)
	return &res, nil
}

// CreateByAdmin memproses refund langsung oleh admin. Nominal yang
// dikembalikan adalah total booking tanpa potongan, dan booking langsung
// dibatalkan dalam transaksi yang sama.
func (s *Service) CreateByAdmin(ctx context.Context, adminID string, req CreateAdminRefundRequest) (*Response, error) {
	req.Normalize()
	if adminID == "" {
		return nil, apperror.Unauthorized("Silakan login terlebih dahulu.")
	}

	booking, err := s.repo.FindBooking(ctx, req.BookingID)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	if booking == nil {
		return nil, ErrBookingNotFound
	}
	if booking.Status != "PAID" && booking.Status != "CONFIRMED" {
		return nil, ErrBookingNotEligible
	}

	existing, err := s.repo.FindActiveByBooking(ctx, req.BookingID)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	if existing != nil {
		return nil, ErrAlreadyRequested
	}

	netAmount, deductionAmount, deductionPercent := calculateAmounts(booking.Total, InitiatedByAdmin)

	reason := req.Reason
	if reason == "" {
		reason = "Refund langsung oleh admin."
	}
	destinationType := req.DestinationType
	if destinationType == "" {
		destinationType = DestinationTypeBank
	}
	now := time.Now()
	refund := &Refund{
		BookingID:        booking.ID,
		UserID:           booking.UserID,
		InitiatedBy:      InitiatedByAdmin,
		DestinationType:  destinationType,
		Amount:           netAmount,
		OriginalAmount:   booking.Total,
		DeductionPercent: deductionPercent,
		DeductionAmount:  deductionAmount,
		Reason:           reason,
		BankName:         req.BankName,
		AccountName:      req.AccountName,
		AccountNumber:    req.AccountNumber,
		Status:           StatusRefunded,
		AdminNote:        req.AdminNote,
		ProcessedBy:      adminID,
		ProcessedAt:      &now,
	}
	if err := s.repo.RefundBooking(ctx, refund); err != nil {
		return nil, apperror.Internal(err)
	}

	record := Record{
		Refund:        *refund,
		BookingCode:   booking.BookingCode,
		VisitorName:   booking.VisitorName,
		Email:         booking.Email,
		VisitDate:     booking.VisitDate,
		BookingStatus: "CANCELLED",
	}
	res := ToResponse(record)
	return &res, nil
}

func (s *Service) ListMine(ctx context.Context, userID string, page, limit int) ([]Response, int64, error) {
	page, limit = normalizePaging(page, limit)
	items, total, err := s.repo.ListByUser(ctx, userID, (page-1)*limit, limit)
	if err != nil {
		return nil, 0, apperror.Internal(err)
	}
	return ToResponseList(items), total, nil
}

func (s *Service) ListAll(ctx context.Context, status string, page, limit int) ([]Response, int64, error) {
	status = strings.ToUpper(strings.TrimSpace(status))
	if status != "" &&
		status != StatusPending && status != StatusApproved &&
		status != StatusRejected && status != StatusRefunded {
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
		Pending:        row.Pending,
		Approved:       row.Approved,
		Refunded:       row.Refunded,
		RefundedAmount: row.RefundedAmount,
	}, nil
}

// SetStatus memproses pengajuan refund oleh admin.
// Alur yang diizinkan: PENDING -> APPROVED/REJECTED, APPROVED -> REFUNDED.
func (s *Service) SetStatus(ctx context.Context, id, adminID string, req UpdateRefundStatusRequest) (*Response, error) {
	req.Normalize()
	if req.Status != StatusApproved && req.Status != StatusRejected && req.Status != StatusRefunded {
		return nil, ErrInvalidStatus
	}

	record, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	if record == nil {
		return nil, ErrNotFound
	}

	switch record.Status {
	case StatusPending:
		if req.Status != StatusApproved && req.Status != StatusRejected {
			return nil, ErrInvalidTransition
		}
	case StatusApproved:
		if req.Status != StatusRefunded {
			return nil, ErrInvalidTransition
		}
	default:
		return nil, ErrInvalidTransition
	}

	now := time.Now()
	record.Status = req.Status
	record.AdminNote = req.AdminNote
	record.ProcessedBy = adminID
	record.ProcessedAt = &now

	if req.Status == StatusRefunded {
		if err := s.repo.RefundBooking(ctx, &record.Refund); err != nil {
			return nil, apperror.Internal(err)
		}
		record.BookingStatus = "CANCELLED"
	} else {
		if err := s.repo.Update(ctx, &record.Refund); err != nil {
			return nil, apperror.Internal(err)
		}
	}

	res := ToResponse(*record)
	return &res, nil
}

// calculateAmounts menghitung nominal bersih dan potongan refund berdasarkan
// sumber pengajuan. Pengunjung dikenakan potongan 10%, admin tidak.
func calculateAmounts(total int64, initiatedBy string) (net, deduction int64, percent int) {
	if initiatedBy == InitiatedByUser {
		percent = UserRefundDeductionPercent
		deduction = total * int64(percent) / 100
		return total - deduction, deduction, percent
	}
	return total, 0, 0
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
