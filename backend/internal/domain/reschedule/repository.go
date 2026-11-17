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
	Items       []Item `gorm:"-"`
}

type TicketCapacity struct {
	ID       string
	Capacity int
	IsActive bool
}

// Record menggabungkan pengajuan dengan data booking untuk tampilan admin.
type Record struct {
	Reschedule
	BookingCode   string
	VisitorName   string
	Email         string
	VisitDate     string
	BookingStatus string
}

type Summary struct {
	Pending  int64
	Approved int64
	Rejected int64
}

type Repository interface {
	FindBooking(ctx context.Context, bookingID string) (*BookingSnapshot, error)
	CountActiveByBooking(ctx context.Context, bookingID string) (int64, error)
	HasPendingByBooking(ctx context.Context, bookingID string) (bool, error)
	FindByID(ctx context.Context, id string) (*Record, error)
	ListByBooking(ctx context.Context, bookingID string) ([]Record, error)
	ListByUser(ctx context.Context, userID string, offset, limit int) ([]Record, int64, error)
	ListAll(ctx context.Context, status string, offset, limit int) ([]Record, int64, error)
	Summary(ctx context.Context) (Summary, error)
	TicketCapacities(ctx context.Context, ticketIDs []string) (map[string]TicketCapacity, error)
	SoldByTicketOn(ctx context.Context, date, excludeBookingID string) (map[string]int64, error)
	Create(ctx context.Context, item *Reschedule) error
	Update(ctx context.Context, item *Reschedule) error
	Approve(ctx context.Context, item *Reschedule) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

const recordSelect = `booking_reschedules.*, bookings.booking_code AS booking_code,
	bookings.visitor_name AS visitor_name, bookings.email AS email,
	bookings.visit_date AS visit_date, bookings.status AS booking_status`

func (r *repository) recordQuery(ctx context.Context) *gorm.DB {
	return r.db.WithContext(ctx).
		Table("booking_reschedules").
		Joins("LEFT JOIN bookings ON bookings.id = booking_reschedules.booking_id AND bookings.deleted_at IS NULL").
		Where("booking_reschedules.deleted_at IS NULL")
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

// CountActiveByBooking menghitung pengajuan yang masih menunggu atau sudah
// disetujui. Pengajuan yang ditolak tidak memakai jatah jadwal ulang.
func (r *repository) CountActiveByBooking(ctx context.Context, bookingID string) (int64, error) {
	var total int64
	err := r.db.WithContext(ctx).
		Model(&Reschedule{}).
		Where("booking_id = ? AND status IN ?", bookingID, []string{StatusPending, StatusApproved}).
		Count(&total).Error
	return total, err
}

func (r *repository) HasPendingByBooking(ctx context.Context, bookingID string) (bool, error) {
	var total int64
	err := r.db.WithContext(ctx).
		Model(&Reschedule{}).
		Where("booking_id = ? AND status = ?", bookingID, StatusPending).
		Count(&total).Error
	return total > 0, err
}

func (r *repository) FindByID(ctx context.Context, id string) (*Record, error) {
	var row Record
	err := r.recordQuery(ctx).
		Select(recordSelect).
		Where("booking_reschedules.id = ?", id).
		Take(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *repository) ListByBooking(ctx context.Context, bookingID string) ([]Record, error) {
	var items []Record
	err := r.recordQuery(ctx).
		Select(recordSelect).
		Where("booking_reschedules.booking_id = ?", bookingID).
		Order("booking_reschedules.created_at DESC").
		Scan(&items).Error
	return items, err
}

// ListByUser mengambil seluruh riwayat pengajuan milik pengunjung, lengkap
// dengan kode booking untuk ditampilkan di panel pengunjung.
func (r *repository) ListByUser(ctx context.Context, userID string, offset, limit int) ([]Record, int64, error) {
	var total int64
	if err := r.db.WithContext(ctx).
		Model(&Reschedule{}).
		Where("user_id = ?", userID).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var items []Record
	err := r.recordQuery(ctx).
		Select(recordSelect).
		Where("booking_reschedules.user_id = ?", userID).
		Order("booking_reschedules.created_at DESC").
		Offset(offset).
		Limit(limit).
		Scan(&items).Error
	return items, total, err
}

func (r *repository) ListAll(ctx context.Context, status string, offset, limit int) ([]Record, int64, error) {
	countQuery := r.db.WithContext(ctx).
		Table("booking_reschedules").
		Where("booking_reschedules.deleted_at IS NULL")
	query := r.recordQuery(ctx)
	if status != "" {
		countQuery = countQuery.Where("booking_reschedules.status = ?", status)
		query = query.Where("booking_reschedules.status = ?", status)
	}

	var total int64
	if err := countQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var items []Record
	err := query.
		Select(recordSelect).
		Order("booking_reschedules.created_at DESC").
		Offset(offset).
		Limit(limit).
		Scan(&items).Error
	return items, total, err
}

func (r *repository) Summary(ctx context.Context) (Summary, error) {
	var row Summary
	err := r.db.WithContext(ctx).
		Table("booking_reschedules").
		Select(`COALESCE(SUM(CASE WHEN status = 'PENDING' THEN 1 ELSE 0 END), 0) AS pending,
			COALESCE(SUM(CASE WHEN status = 'APPROVED' THEN 1 ELSE 0 END), 0) AS approved,
			COALESCE(SUM(CASE WHEN status = 'REJECTED' THEN 1 ELSE 0 END), 0) AS rejected`).
		Where("deleted_at IS NULL").
		Scan(&row).Error
	return row, err
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

func (r *repository) Create(ctx context.Context, item *Reschedule) error {
	return r.db.WithContext(ctx).Create(item).Error
}

func (r *repository) Update(ctx context.Context, item *Reschedule) error {
	return r.db.WithContext(ctx).Save(item).Error
}

// Approve menerapkan tanggal baru ke booking dan e-tiket dalam satu transaksi.
// Pengajuan hanya bisa disetujui bila statusnya masih PENDING dan tanggal
// booking belum berubah sejak pengajuan dibuat.
func (r *repository) Approve(ctx context.Context, item *Reschedule) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		now := time.Now()

		result := tx.Model(&Reschedule{}).
			Where("id = ? AND status = ?", item.ID, StatusPending).
			Updates(map[string]interface{}{
				"status":       StatusApproved,
				"admin_note":   item.AdminNote,
				"processed_by": item.ProcessedBy,
				"processed_at": now,
				"updated_at":   now,
			})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return ErrInvalidTransition
		}

		result = tx.Table("bookings").
			Where("id = ? AND visit_date = ? AND status IN ?", item.BookingID, item.OldVisitDate, []string{"PAID", "CONFIRMED"}).
			Updates(map[string]interface{}{
				"visit_date": item.NewVisitDate,
				"updated_at": now,
			})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return ErrBookingChanged
		}

		// E-tiket menyimpan tanggal kunjungan sendiri, jadi ikut diperbarui
		// agar tanggal pada tiket tidak berbeda dengan booking.
		return tx.Table("tickets").
			Where("booking_id = ?", item.BookingID).
			Updates(map[string]interface{}{
				"visit_date": item.NewVisitDate,
				"updated_at": now,
			}).Error
	})
}
