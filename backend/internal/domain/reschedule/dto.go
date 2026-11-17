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

type Response struct {
	ID           string `json:"id"`
	BookingID    string `json:"booking_id"`
	BookingCode  string `json:"booking_code"`
	OldVisitDate string `json:"old_visit_date"`
	NewVisitDate string `json:"new_visit_date"`
	Reason       string `json:"reason"`
	Status       string `json:"status"`
	CreatedAt    string `json:"created_at"`
}

func ToResponse(item Reschedule, booking *BookingSnapshot) Response {
	res := Response{
		ID:           item.ID,
		BookingID:    item.BookingID,
		OldVisitDate: item.OldVisitDate,
		NewVisitDate: item.NewVisitDate,
		Reason:       item.Reason,
		Status:       item.Status,
		CreatedAt:    item.CreatedAt.Format(time.RFC3339),
	}
	if booking != nil {
		res.BookingCode = booking.BookingCode
	}
	return res
}

func ToResponseList(items []Reschedule, booking *BookingSnapshot) []Response {
	out := make([]Response, 0, len(items))
	for _, item := range items {
		out = append(out, ToResponse(item, booking))
	}
	return out
}

// toRecordResponse melengkapi response dengan kode booking dari join.
func toRecordResponse(item Record) Response {
	res := ToResponse(item.Reschedule, nil)
	res.BookingCode = item.BookingCode
	return res
}
