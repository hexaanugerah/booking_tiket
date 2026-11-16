package review

import "time"

const (
	StatusPending  = "PENDING"
	StatusApproved = "APPROVED"
	StatusRejected = "REJECTED"
)

type CreateReviewRequest struct {
	DestinationID string `json:"destination_id" validate:"omitempty,uuid4"`
	BookingID     string `json:"booking_id" validate:"required,uuid4"`
	Rating        int    `json:"rating" validate:"required,min=1,max=5"`
	Comment       string `json:"comment" validate:"required,min=10,max=1000"`
}

func (r *CreateReviewRequest) Normalize() {
	r.BookingID = trimSpace(r.BookingID)
	r.DestinationID = trimSpace(r.DestinationID)
	r.Comment = NormalizeComment(r.Comment)
}

type UpdateReviewRequest struct {
	Rating  int    `json:"rating" validate:"required,min=1,max=5"`
	Comment string `json:"comment" validate:"required,min=10,max=1000"`
}

func (r *UpdateReviewRequest) Normalize() {
	r.Comment = NormalizeComment(r.Comment)
}

type ModerateReviewRequest struct {
	Status string `json:"status" validate:"required,oneof=APPROVED REJECTED"`
}

type Response struct {
	ID            string `json:"id"`
	UserID        string `json:"user_id"`
	BookingID     string `json:"booking_id,omitempty"`
	DestinationID string `json:"destination_id,omitempty"`
	Rating        int    `json:"rating"`
	Comment       string `json:"comment"`
	Status        string `json:"status"`
	IsApproved    bool   `json:"is_approved"`
	CreatedAt     string `json:"created_at"`
	UpdatedAt     string `json:"updated_at"`
}

func ToResponse(r *Review) *Response {
	if r == nil {
		return nil
	}
	bookingID := ""
	if r.BookingID != nil {
		bookingID = *r.BookingID
	}
	return &Response{
		ID:            r.ID,
		UserID:        r.UserID,
		BookingID:     bookingID,
		DestinationID: r.DestinationID,
		Rating:        r.Rating,
		Comment:       r.Comment,
		Status:        r.Status,
		IsApproved:    r.IsApproved,
		CreatedAt:     r.CreatedAt.Format(time.RFC3339),
		UpdatedAt:     r.UpdatedAt.Format(time.RFC3339),
	}
}

func ToResponseList(items []Review) []Response {
	out := make([]Response, 0, len(items))
	for i := range items {
		out = append(out, *ToResponse(&items[i]))
	}
	return out
}
