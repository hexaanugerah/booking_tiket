package refund

import (
	"strings"
	"time"
)

type CreateRefundRequest struct {
	BookingID     string `json:"booking_id" validate:"required,uuid4"`
	Reason        string `json:"reason" validate:"required,min=10,max=500"`
	BankName      string `json:"bank_name" validate:"required,min=2,max=100"`
	AccountName   string `json:"account_name" validate:"required,min=2,max=150"`
	AccountNumber string `json:"account_number" validate:"required,min=4,max=50"`
}

func (r *CreateRefundRequest) Normalize() {
	r.BookingID = strings.TrimSpace(r.BookingID)
	r.Reason = strings.TrimSpace(r.Reason)
	r.BankName = strings.TrimSpace(r.BankName)
	r.AccountName = strings.TrimSpace(r.AccountName)
	r.AccountNumber = strings.TrimSpace(r.AccountNumber)
}

type UpdateRefundStatusRequest struct {
	Status    string `json:"status" validate:"required,oneof=APPROVED REJECTED REFUNDED"`
	AdminNote string `json:"admin_note" validate:"omitempty,max=500"`
}

func (r *UpdateRefundStatusRequest) Normalize() {
	r.Status = strings.ToUpper(strings.TrimSpace(r.Status))
	r.AdminNote = strings.TrimSpace(r.AdminNote)
}

type SummaryResponse struct {
	Pending        int64 `json:"pending"`
	Approved       int64 `json:"approved"`
	Refunded       int64 `json:"refunded"`
	RefundedAmount int64 `json:"refunded_amount"`
}

type Response struct {
	ID            string `json:"id"`
	BookingID     string `json:"booking_id"`
	UserID        string `json:"user_id"`
	Amount        int64  `json:"amount"`
	Reason        string `json:"reason"`
	BankName      string `json:"bank_name"`
	AccountName   string `json:"account_name"`
	AccountNumber string `json:"account_number"`
	Status        string `json:"status"`
	AdminNote     string `json:"admin_note,omitempty"`
	ProcessedBy   string `json:"processed_by,omitempty"`
	ProcessedAt   string `json:"processed_at,omitempty"`

	// Data booking untuk memudahkan tabel admin tanpa request tambahan.
	BookingCode   string `json:"booking_code"`
	VisitorName   string `json:"visitor_name"`
	Email         string `json:"email"`
	VisitDate     string `json:"visit_date"`
	BookingStatus string `json:"booking_status"`

	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

func ToResponse(record Record) Response {
	r := record.Refund
	res := Response{
		ID:            r.ID,
		BookingID:     r.BookingID,
		UserID:        r.UserID,
		Amount:        r.Amount,
		Reason:        r.Reason,
		BankName:      r.BankName,
		AccountName:   r.AccountName,
		AccountNumber: r.AccountNumber,
		Status:        r.Status,
		AdminNote:     r.AdminNote,
		ProcessedBy:   r.ProcessedBy,
		BookingCode:   record.BookingCode,
		VisitorName:   record.VisitorName,
		Email:         record.Email,
		VisitDate:     record.VisitDate,
		BookingStatus: record.BookingStatus,
		CreatedAt:     r.CreatedAt.Format(time.RFC3339),
		UpdatedAt:     r.UpdatedAt.Format(time.RFC3339),
	}
	if r.ProcessedAt != nil {
		res.ProcessedAt = r.ProcessedAt.Format(time.RFC3339)
	}
	return res
}

func ToResponseList(records []Record) []Response {
	out := make([]Response, 0, len(records))
	for _, record := range records {
		out = append(out, ToResponse(record))
	}
	return out
}
