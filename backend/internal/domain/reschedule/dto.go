package reschedule

import (
	"strings"
	"time"
)

type CreateRequest struct {
	NewVisitDate string `json:"new_visit_date" validate:"required,len=10"`
	Reason       string `json:"reason" validate:"required,min=10,max=500"`
}

func (r *CreateRequest) Normalize() {
	r.NewVisitDate = strings.TrimSpace(r.NewVisitDate)
	r.Reason = strings.TrimSpace(r.Reason)
}

// UpdateStatusRequest dipakai admin untuk menyetujui atau menolak pengajuan.
type UpdateStatusRequest struct {
	Status    string `json:"status" validate:"required,oneof=APPROVED REJECTED"`
	AdminNote string `json:"admin_note" validate:"omitempty,max=500"`
}

func (r *UpdateStatusRequest) Normalize() {
	r.Status = strings.ToUpper(strings.TrimSpace(r.Status))
	r.AdminNote = strings.TrimSpace(r.AdminNote)
}

type SummaryResponse struct {
	Pending  int64 `json:"pending"`
	Approved int64 `json:"approved"`
	Rejected int64 `json:"rejected"`
}

type Response struct {
	ID           string `json:"id"`
	BookingID    string `json:"booking_id"`
	BookingCode  string `json:"booking_code"`
	UserID       string `json:"user_id"`
	OldVisitDate string `json:"old_visit_date"`
	NewVisitDate string `json:"new_visit_date"`
	Reason       string `json:"reason"`
	Status       string `json:"status"`
	AdminNote    string `json:"admin_note,omitempty"`
	ProcessedBy  string `json:"processed_by,omitempty"`
	ProcessedAt  string `json:"processed_at,omitempty"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`

	// Data booking untuk ditampilkan di tabel admin tanpa request tambahan.
	VisitorName   string `json:"visitor_name"`
	Email         string `json:"email"`
	VisitDate     string `json:"visit_date"`
	BookingStatus string `json:"booking_status"`
}

func ToResponse(record Record) Response {
	item := record.Reschedule
	res := Response{
		ID:            item.ID,
		BookingID:     item.BookingID,
		BookingCode:   record.BookingCode,
		UserID:        item.UserID,
		OldVisitDate:  item.OldVisitDate,
		NewVisitDate:  item.NewVisitDate,
		Reason:        item.Reason,
		Status:        item.Status,
		AdminNote:     item.AdminNote,
		ProcessedBy:   item.ProcessedBy,
		VisitorName:   record.VisitorName,
		Email:         record.Email,
		VisitDate:     record.VisitDate,
		BookingStatus: record.BookingStatus,
		CreatedAt:     item.CreatedAt.Format(time.RFC3339),
		UpdatedAt:     item.UpdatedAt.Format(time.RFC3339),
	}
	if item.ProcessedAt != nil {
		res.ProcessedAt = item.ProcessedAt.Format(time.RFC3339)
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

func recordFromBooking(item Reschedule, booking *BookingSnapshot) Record {
	record := Record{Reschedule: item}
	if booking != nil {
		record.BookingCode = booking.BookingCode
		record.VisitorName = booking.VisitorName
		record.Email = booking.Email
		record.VisitDate = booking.VisitDate
		record.BookingStatus = booking.Status
	}
	return record
}
