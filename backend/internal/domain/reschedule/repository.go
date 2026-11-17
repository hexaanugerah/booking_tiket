package reschedule

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
)

type Item struct {
	TicketID   string
	TicketName string
	Quantity   int
}

type BookingSnapshot struct {
	ID          string
	UserID      string
	BookingCode string
	VisitorName string
	Email       string
	VisitDate   string
	Status      string
	Items       []Item
}

type TicketCapacity struct {
	ID       string
	Capacity int
	IsActive bool
}

type Record struct {
	Reschedule
	BookingCode string
}

type Repository interface {
	FindBooking(ctx context.Context, bookingID string) (*BookingSnapshot, error)
	CountByBooking(ctx context.Context, bookingID string) (int64, error)
	ListByBooking(ctx context.Context, bookingID string) ([]Reschedule, error)
	ListByUser(ctx context.Context, userID string, offset, limit int) ([]Record, int64, error)
	TicketCapacities(ctx context.Context, ticketIDs []string) (map[string]TicketCapacity, error)
	SoldByTicketOn(ctx context.Context, date, excludeBookingID string) (map[string]int64, error)
	CreateAndApply(ctx context.Context, item *Reschedule) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) FindBooking(ctx context.Context, bookingID string) (*BookingSnapshot, error) {
	var row BookingSnapshot
	err := r.db.WithContext(ctx).
		Table("bookings").
		Select("id, user_id, booking_code, visitor_name, email, visit_date, status").
		Where("id = ? AND deleted_at IS NULL", bookingID).
		Take(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var items []Item
	err = r.db.WithContext(ctx).
		Table("booking_items").
		Select("ticket_id, ticket_name, quantity").
		Where("booking_id = ? AND deleted_at IS NULL", bookingID).
		Scan(&items).Error
	if err != nil {
		return nil, err
	}
	row.Items = items
	return &row, nil
}

func (r *repository) CountByBooking(ctx context.Context, bookingID string) (int64, error) {
	var total int64
	err := r.db.WithContext(ctx).
		Model(&Reschedule{}).
		Where("booking_id = ?", bookingID).
		Count(&total).Error
	return total, err
}

func (r *repository) ListByBooking(ctx context.Context, bookingID string) ([]Reschedule, error) {
	var items []Reschedule
	err := r.db.WithContext(ctx).
		Where("booking_id = ?", bookingID).
		Order("created_at DESC").
		Find(&items).Error
	return items, err
}

// ListByUser mengambil seluruh riwayat jadwal ulang milik pengunjung,
// lengkap dengan kode booking untuk ditampilkan di panel pengunjung.
func (r *repository) ListByUser(ctx context.Context, userID string, offset, limit int) ([]Record, int64, error) {
	var total int64
	if err := r.db.WithContext(ctx).
		Model(&Reschedule{}).
		Where("user_id = ?", userID).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var items []Record
	err := r.db.WithContext(ctx).
		Table("booking_reschedules").
		Select("booking_reschedules.*, bookings.booking_code AS booking_code").
		Joins("LEFT JOIN bookings ON bookings.id = booking_reschedules.booking_id AND bookings.deleted_at IS NULL").
		Where("booking_reschedules.user_id = ? AND booking_reschedules.deleted_at IS NULL", userID).
		Order("booking_reschedules.created_at DESC").
		Offset(offset).
		Limit(limit).
		Scan(&items).Error
	return items, total, err
}

func (r *repository) TicketCapacities(ctx context.Context, ticketIDs []string) (map[string]TicketCapacity, error) {
	var rows []TicketCapacity
	if err := r.db.WithContext(ctx).
		Table("ticket_types").
		Select("id, capacity, is_active").
		Where("id IN ? AND deleted_at IS NULL", ticketIDs).
		Scan(&rows).Error; err != nil {
		return nil, err
	}
	out := make(map[string]TicketCapacity, len(rows))
	for _, row := range rows {
		out[row.ID] = row
	}
	return out, nil
}

func (r *repository) SoldByTicketOn(ctx context.Context, date, excludeBookingID string) (map[string]int64, error) {
	type soldRow struct {
		TicketID string
		Total    int64
	}
	var rows []soldRow

	query := r.db.WithContext(ctx).
		Table("booking_items").
		Select("booking_items.ticket_id AS ticket_id, COALESCE(SUM(booking_items.quantity), 0) AS total").
		Joins("JOIN bookings ON bookings.id = booking_items.booking_id").
		Where("bookings.visit_date = ?", date).
		Where("bookings.status IN ?", []string{"PENDING", "PAID", "CONFIRMED", "USED"}).
		Where("bookings.deleted_at IS NULL AND booking_items.deleted_at IS NULL")
	if excludeBookingID != "" {
		query = query.Where("bookings.id <> ?", excludeBookingID)
	}

	if err := query.Group("booking_items.ticket_id").Scan(&rows).Error; err != nil {
		return nil, err
	}
	out := make(map[string]int64, len(rows))
	for _, row := range rows {
		out[row.TicketID] = row.Total
	}
	return out, nil
}

func (r *repository) CreateAndApply(ctx context.Context, item *Reschedule) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(item).Error; err != nil {
			return err
		}
		result := tx.Table("bookings").
			Where("id = ? AND user_id = ? AND status IN ?", item.BookingID, item.UserID, []string{"PAID", "CONFIRMED"}).
			Updates(map[string]interface{}{
				"visit_date": item.NewVisitDate,
				"updated_at": time.Now(),
			})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return ErrBookingNotEligible
		}
		// E-tiket menyimpan tanggal kunjungan sendiri, jadi ikut diperbarui
		// agar tanggal pada tiket tidak berbeda dengan booking.
		return tx.Table("tickets").
			Where("booking_id = ?", item.BookingID).
			Updates(map[string]interface{}{
				"visit_date": item.NewVisitDate,
				"updated_at": time.Now(),
			}).Error
	})
}
