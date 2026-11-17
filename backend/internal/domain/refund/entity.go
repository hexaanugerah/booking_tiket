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

const (
	// InitiatedByUser menandai refund yang diajukan sendiri oleh pengunjung.
	InitiatedByUser = "USER"
	// InitiatedByAdmin menandai refund yang diproses langsung oleh admin.
	InitiatedByAdmin = "ADMIN"

	// UserRefundDeductionPercent adalah potongan biaya administrasi untuk
	// refund yang diajukan pengunjung. Refund langsung oleh admin tidak
	// dikenakan potongan.
	UserRefundDeductionPercent = 10
)

const (
	// DestinationTypeBank menandai refund dikirim ke rekening bank.
	DestinationTypeBank = "BANK"
	// DestinationTypeEWallet menandai refund dikirim ke e-wallet (DANA, OVO, dll).
	DestinationTypeEWallet = "EWALLET"
)

// Refund adalah pengembalian dana untuk satu booking.
// Amount adalah nominal bersih yang dikembalikan, sedangkan OriginalAmount
// adalah total pembayaran booking sebelum potongan.
type Refund struct {
	Base
	BookingID        string     `gorm:"type:char(36);index;not null" json:"booking_id"`
	UserID           string     `gorm:"type:char(36);index;not null" json:"user_id"`
	InitiatedBy      string     `gorm:"type:varchar(10);not null;default:USER" json:"initiated_by"`
	DestinationType  string     `gorm:"type:varchar(10);not null;default:BANK" json:"destination_type"`
	OriginalAmount   int64      `gorm:"not null;default:0" json:"original_amount"`
	DeductionPercent int        `gorm:"not null;default:0" json:"deduction_percent"`
	DeductionAmount  int64      `gorm:"not null;default:0" json:"deduction_amount"`
	Amount           int64      `gorm:"not null;default:0" json:"amount"`
	Reason           string     `gorm:"type:varchar(500);not null" json:"reason"`
	BankName         string     `gorm:"type:varchar(100);not null" json:"bank_name"`
	AccountName      string     `gorm:"type:varchar(150);not null" json:"account_name"`
	AccountNumber    string     `gorm:"type:varchar(50);not null" json:"account_number"`
	Status           string     `gorm:"type:varchar(20);index;not null;default:PENDING" json:"status"`
	AdminNote        string     `gorm:"type:varchar(500)" json:"admin_note,omitempty"`
	ProcessedBy      string     `gorm:"type:char(36);index" json:"processed_by,omitempty"`
	ProcessedAt      *time.Time `json:"processed_at,omitempty"`
}

func (Refund) TableName() string { return "refunds" }
