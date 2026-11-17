package reschedule

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Base struct {
	ID        string         `gorm:"type:char(36);primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (b *Base) BeforeCreate(tx *gorm.DB) error {
	if b.ID == "" {
		b.ID = uuid.NewString()
	}
	return nil
}

const (
	// StatusPending menandai pengajuan jadwal ulang yang menunggu keputusan admin.
	StatusPending = "PENDING"
	// StatusApproved menandai pengajuan yang disetujui; tanggal booking sudah diubah.
	StatusApproved = "APPROVED"
	// StatusRejected menandai pengajuan yang ditolak; tanggal booking tidak berubah.
	StatusRejected = "REJECTED"
)

// Reschedule menyimpan riwayat pengajuan perubahan tanggal kunjungan.
// Tanggal booking baru benar-benar diubah saat admin menyetujui pengajuan.
// OldVisitDate/NewVisitDate memakai format "2006-01-02".
type Reschedule struct {
	Base
	BookingID    string     `gorm:"type:char(36);index;not null" json:"booking_id"`
	UserID       string     `gorm:"type:char(36);index;not null" json:"user_id"`
	OldVisitDate string     `gorm:"type:varchar(10);not null" json:"old_visit_date"`
	NewVisitDate string     `gorm:"type:varchar(10);not null" json:"new_visit_date"`
	Reason       string     `gorm:"type:varchar(500);not null" json:"reason"`
	Status       string     `gorm:"type:varchar(20);index;not null;default:PENDING" json:"status"`
	AdminNote    string     `gorm:"type:varchar(500)" json:"admin_note,omitempty"`
	ProcessedBy  string     `gorm:"type:char(36);index" json:"processed_by,omitempty"`
	ProcessedAt  *time.Time `json:"processed_at,omitempty"`
}

func (Reschedule) TableName() string { return "booking_reschedules" }
