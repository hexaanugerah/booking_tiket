package review

import (
	"context"
	"errors"

	"gorm.io/gorm"
)

type BookingSnapshot struct {
	ID     string
	UserID string
	Status string
}

type Repository interface {
	Create(ctx context.Context, review *Review) error
	FindByID(ctx context.Context, id string) (*Review, error)
	FindByBookingID(ctx context.Context, bookingID string) (*Review, error)
	ListByUser(ctx context.Context, userID string, offset, limit int) ([]Review, int64, error)
	ListAll(ctx context.Context, status string, offset, limit int) ([]Review, int64, error)
	ListApproved(ctx context.Context, offset, limit int) ([]Review, int64, error)
	Summary(ctx context.Context) (float64, int64, error)
	Update(ctx context.Context, review *Review) error
	Delete(ctx context.Context, id string) error
	DestinationExists(ctx context.Context, id string) (bool, error)
	FindBooking(ctx context.Context, bookingID string) (*BookingSnapshot, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, review *Review) error {
	return r.db.WithContext(ctx).Create(review).Error
}

func (r *repository) FindByID(ctx context.Context, id string) (*Review, error) {
	var item Review
	err := r.db.WithContext(ctx).First(&item, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *repository) FindByBookingID(ctx context.Context, bookingID string) (*Review, error) {
	var item Review
	err := r.db.WithContext(ctx).Where("booking_id = ?", bookingID).First(&item).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *repository) ListByUser(ctx context.Context, userID string, offset, limit int) ([]Review, int64, error) {
	q := r.db.WithContext(ctx).Model(&Review{}).Where("user_id = ?", userID)
	return r.list(q, offset, limit)
}

func (r *repository) ListAll(ctx context.Context, status string, offset, limit int) ([]Review, int64, error) {
	q := r.db.WithContext(ctx).Model(&Review{})
	if status != "" {
		q = q.Where("status = ?", status)
	}
	return r.list(q, offset, limit)
}

func (r *repository) ListApproved(ctx context.Context, offset, limit int) ([]Review, int64, error) {
	q := r.db.WithContext(ctx).Model(&Review{}).Where("is_approved = ?", true)
	return r.list(q, offset, limit)
}

// Summary menghitung rata-rata rating dari seluruh ulasan yang tayang.
func (r *repository) Summary(ctx context.Context) (float64, int64, error) {
	var row struct {
		Average float64
		Total   int64
	}
	err := r.db.WithContext(ctx).
		Model(&Review{}).
		Select("COALESCE(AVG(rating), 0) AS average, COUNT(*) AS total").
		Where("is_approved = ?", true).
		Scan(&row).Error
	return row.Average, row.Total, err
}

func (r *repository) list(q *gorm.DB, offset, limit int) ([]Review, int64, error) {
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var items []Review
	err := q.Order("created_at DESC").Offset(offset).Limit(limit).Find(&items).Error
	return items, total, err
}

func (r *repository) Update(ctx context.Context, review *Review) error {
	return r.db.WithContext(ctx).Save(review).Error
}

func (r *repository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&Review{}, "id = ?", id).Error
}

func (r *repository) DestinationExists(ctx context.Context, id string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Table("destinations").Where("id = ?", id).Count(&count).Error
	return count > 0, err
}

func (r *repository) FindBooking(ctx context.Context, bookingID string) (*BookingSnapshot, error) {
	var row BookingSnapshot
	err := r.db.WithContext(ctx).
		Table("bookings").
		Select("id, user_id, status").
		Where("id = ?", bookingID).
		Take(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}
