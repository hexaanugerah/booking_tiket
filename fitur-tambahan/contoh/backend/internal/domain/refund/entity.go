package refund

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
	StatusPending  = "PENDING"
	StatusApproved = "APPROVED"
	StatusRejected = "REJECTED"
	StatusRefunded = "REFUNDED"
)

// Refund adalah pengajuan pengembalian dana untuk satu booking.
// Nominal disimpan sebagai snapshot total booking saat pengajuan dibuat.
type Refund struct {
	Base
	BookingID     string     `gorm:"type:char(36);index;not null" json:"booking_id"`
	UserID        string     `gorm:"type:char(36);index;not null" json:"user_id"`
	Amount        int64      `gorm:"not null;default:0" json:"amount"`
	Reason        string     `gorm:"type:varchar(500);not null" json:"reason"`
	BankName      string     `gorm:"type:varchar(100);not null" json:"bank_name"`
	AccountName   string     `gorm:"type:varchar(150);not null" json:"account_name"`
	AccountNumber string     `gorm:"type:varchar(50);not null" json:"account_number"`
	Status        string     `gorm:"type:varchar(20);index;not null;default:PENDING" json:"status"`
	AdminNote     string     `gorm:"type:varchar(500)" json:"admin_note,omitempty"`
	ProcessedBy   string     `gorm:"type:char(36);index" json:"processed_by,omitempty"`
	ProcessedAt   *time.Time `json:"processed_at,omitempty"`
}

func (Refund) TableName() string { return "refunds" }
