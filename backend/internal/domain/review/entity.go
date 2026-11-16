package review

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

type Review struct {
	Base
	UserID        string  `gorm:"type:char(36);index" json:"user_id"`
	BookingID     *string `gorm:"type:char(36);uniqueIndex" json:"booking_id,omitempty"`
	DestinationID string  `gorm:"type:char(36);index" json:"destination_id,omitempty"`
	Rating        int     `gorm:"not null;default:0" json:"rating"`
	Comment       string  `gorm:"type:text" json:"comment"`
	Status        string  `gorm:"type:varchar(20);index;default:PENDING" json:"status"`
	IsApproved    bool    `gorm:"default:false;index" json:"is_approved"`
}

func (Review) TableName() string { return "reviews" }
