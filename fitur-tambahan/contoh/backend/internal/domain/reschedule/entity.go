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
	StatusApproved = "APPROVED"
)

// Reschedule menyimpan riwayat perubahan tanggal kunjungan.
// OldVisitDate/NewVisitDate memakai format "2006-01-02".
type Reschedule struct {
	Base
	BookingID    string `gorm:"type:char(36);index;not null" json:"booking_id"`
	UserID       string `gorm:"type:char(36);index;not null" json:"user_id"`
	OldVisitDate string `gorm:"type:varchar(10);not null" json:"old_visit_date"`
	NewVisitDate string `gorm:"type:varchar(10);not null" json:"new_visit_date"`
	Reason       string `gorm:"type:varchar(500);not null" json:"reason"`
	Status       string `gorm:"type:varchar(20);index;not null;default:APPROVED" json:"status"`
}

func (Reschedule) TableName() string { return "booking_reschedules" }
