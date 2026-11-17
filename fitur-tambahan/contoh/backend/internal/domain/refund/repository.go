package refund

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
)

type BookingSnapshot struct {
	ID          string
	UserID      string
	BookingCode string
	VisitorName string
	Email       string
	VisitDate   string
	Status      string
	Total       int64
}

type Record struct {
	Refund
	BookingCode   string
	VisitorName   string
	Email         string
	VisitDate     string
	BookingStatus string
}

type Summary struct {
	Pending        int64
	Approved       int64
	Refunded       int64
	RefundedAmount int64
}

type Repository interface {
	Create(ctx context.Context, refund *Refund) error
	FindByID(ctx context.Context, id string) (*Record, error)
	FindActiveByBooking(ctx context.Context, bookingID string) (*Refund, error)
	ListByUser(ctx context.Context, userID string, offset, limit int) ([]Record, int64, error)
	ListAll(ctx context.Context, status string, offset, limit int) ([]Record, int64, error)
	Summary(ctx context.Context) (Summary, error)
	FindBooking(ctx context.Context, bookingID string) (*BookingSnapshot, error)
	Update(ctx context.Context, refund *Refund) error
	RefundBooking(ctx context.Context, refund *Refund) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

const recordSelect = `refunds.id, refunds.created_at, refunds.updated_at, refunds.deleted_at,
	refunds.booking_id, refunds.user_id, refunds.amount, refunds.reason, refunds.bank_name,
	refunds.account_name, refunds.account_number, refunds.status, refunds.admin_note,
	refunds.processed_by, refunds.processed_at,
	bookings.booking_code, bookings.visitor_name, bookings.email,
	bookings.visit_date, bookings.status AS booking_status`

func (r *repository) recordQuery(ctx context.Context) *gorm.DB {
	return r.db.WithContext(ctx).
		Table("refunds").
		Joins("LEFT JOIN bookings ON bookings.id = refunds.booking_id AND bookings.deleted_at IS NULL").
		Where("refunds.deleted_at IS NULL")
}

func (r *repository) Create(ctx context.Context, refund *Refund) error {
	return r.db.WithContext(ctx).Create(refund).Error
}

func (r *repository) FindByID(ctx context.Context, id string) (*Record, error) {
	var row Record
	err := r.recordQuery(ctx).
		Select(recordSelect).
		Where("refunds.id = ?", id).
		Take(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *repository) FindActiveByBooking(ctx context.Context, bookingID string) (*Refund, error) {
	var item Refund
	err := r.db.WithContext(ctx).
		Where("booking_id = ? AND status IN ?", bookingID, []string{StatusPending, StatusApproved, StatusRefunded}).
		First(&item).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *repository) ListByUser(ctx context.Context, userID string, offset, limit int) ([]Record, int64, error) {
	return r.list(ctx, "refunds.user_id = ?", []interface{}{userID}, offset, limit)
}

func (r *repository) ListAll(ctx context.Context, status string, offset, limit int) ([]Record, int64, error) {
	if status == "" {
		return r.list(ctx, "", nil, offset, limit)
	}
	return r.list(ctx, "refunds.status = ?", []interface{}{status}, offset, limit)
}

func (r *repository) list(ctx context.Context, where string, args []interface{}, offset, limit int) ([]Record, int64, error) {
	countQuery := r.db.WithContext(ctx).Table("refunds").Where("refunds.deleted_at IS NULL")
	if where != "" {
		countQuery = countQuery.Where(where, args...)
	}
	var total int64
	if err := countQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	query := r.recordQuery(ctx)
	if where != "" {
		query = query.Where(where, args...)
	}
	var items []Record
	err := query.
		Select(recordSelect).
		Order("refunds.created_at DESC").
		Offset(offset).
		Limit(limit).
		Scan(&items).Error
	return items, total, err
}

func (r *repository) Summary(ctx context.Context) (Summary, error) {
	var row Summary
	err := r.db.WithContext(ctx).
		Table("refunds").
		Select(`COALESCE(SUM(status = 'PENDING'), 0) AS pending,
			COALESCE(SUM(status = 'APPROVED'), 0) AS approved,
			COALESCE(SUM(status = 'REFUNDED'), 0) AS refunded,
			COALESCE(SUM(CASE WHEN status = 'REFUNDED' THEN amount ELSE 0 END), 0) AS refunded_amount`).
		Where("deleted_at IS NULL").
		Scan(&row).Error
	return row, err
}

func (r *repository) FindBooking(ctx context.Context, bookingID string) (*BookingSnapshot, error) {
	var row BookingSnapshot
	err := r.db.WithContext(ctx).
		Table("bookings").
		Select("id, user_id, booking_code, visitor_name, email, visit_date, status, total").
		Where("id = ? AND deleted_at IS NULL", bookingID).
		Take(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *repository) Update(ctx context.Context, refund *Refund) error {
	return r.db.WithContext(ctx).Save(refund).Error
}

// RefundBooking menyimpan status refund dan, dalam transaksi yang sama,
// menandai payment/booking/e-ticket sebagai refunded/cancelled.
func (r *repository) RefundBooking(ctx context.Context, refund *Refund) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(refund).Error; err != nil {
			return err
		}
		now := time.Now()
		if err := tx.Table("payments").
			Where("booking_id = ?", refund.BookingID).
			Updates(map[string]interface{}{
				"status":            "REFUNDED",
				"settlement_status": "REFUNDED",
				"updated_at":        now,
			}).Error; err != nil {
			return err
		}
		if err := tx.Table("bookings").
			Where("id = ?", refund.BookingID).
			Updates(map[string]interface{}{
				"status":     "CANCELLED",
				"updated_at": now,
			}).Error; err != nil {
			return err
		}
		return tx.Table("tickets").
			Where("booking_id = ?", refund.BookingID).
			Updates(map[string]interface{}{
				"status":     "CANCELLED",
				"updated_at": now,
			}).Error
	})
}
