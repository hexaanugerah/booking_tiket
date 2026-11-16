package review

import (
	"strings"
	"testing"

	"github.com/KawahPutih/backend/pkg/response"
)

func TestValidateRating(t *testing.T) {
	cases := []struct {
		name    string
		rating  int
		wantErr bool
	}{
		{name: "rating 0 ditolak", rating: 0, wantErr: true},
		{name: "rating 1 diterima", rating: 1},
		{name: "rating 5 diterima", rating: 5},
		{name: "rating 6 ditolak", rating: 6, wantErr: true},
		{name: "rating negatif ditolak", rating: -1, wantErr: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateRating(tc.rating)
			if (err != nil) != tc.wantErr {
				t.Fatalf("ValidateRating(%d) error = %v, wantErr %v", tc.rating, err, tc.wantErr)
			}
		})
	}
}

func TestValidateComment(t *testing.T) {
	cases := []struct {
		name    string
		comment string
		wantErr bool
	}{
		{name: "kosong ditolak", comment: "", wantErr: true},
		{name: "spasi saja ditolak", comment: "     ", wantErr: true},
		{name: "9 karakter ditolak", comment: strings.Repeat("a", 9), wantErr: true},
		{name: "10 karakter diterima", comment: strings.Repeat("a", 10)},
		{name: "spasi tepi tidak dihitung", comment: "  halo dunia  "},
		{name: "1000 karakter diterima", comment: strings.Repeat("a", 1000)},
		{name: "1001 karakter ditolak", comment: strings.Repeat("a", 1001), wantErr: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateComment(tc.comment)
			if (err != nil) != tc.wantErr {
				t.Fatalf("ValidateComment(panjang %d) error = %v, wantErr %v", len(tc.comment), err, tc.wantErr)
			}
		})
	}
}

func TestCreateReviewRequestValidation(t *testing.T) {
	const validBookingID = "550e8400-e29b-41d4-a716-446655440000"

	valid := CreateReviewRequest{
		BookingID: validBookingID,
		Rating:    5,
		Comment:   "Pengalaman yang menyenangkan.",
	}
	if errs := Validate(valid); errs != nil {
		t.Fatalf("request valid seharusnya lolos, dapat error: %+v", errs)
	}

	invalidRating := valid
	invalidRating.Rating = 0
	if errs := Validate(invalidRating); !hasFieldError(errs, "rating") {
		t.Fatalf("rating 0 seharusnya menghasilkan error field rating, dapat: %+v", errs)
	}

	shortComment := valid
	shortComment.Comment = "pendek"
	if errs := Validate(shortComment); !hasFieldError(errs, "comment") {
		t.Fatalf("ulasan pendek seharusnya menghasilkan error field comment, dapat: %+v", errs)
	}

	missingBooking := valid
	missingBooking.BookingID = ""
	if errs := Validate(missingBooking); !hasFieldError(errs, "booking_id") {
		t.Fatalf("booking kosong seharusnya menghasilkan error field booking_id, dapat: %+v", errs)
	}

	invalidBooking := valid
	invalidBooking.BookingID = "bukan-uuid"
	if errs := Validate(invalidBooking); !hasFieldError(errs, "booking_id") {
		t.Fatalf("booking bukan UUID seharusnya ditolak, dapat: %+v", errs)
	}
}

func TestModerateReviewRequestNormalizeAndValidate(t *testing.T) {
	req := ModerateReviewRequest{Status: " approved "}
	req.Normalize()
	if req.Status != StatusApproved {
		t.Fatalf("status harus dinormalisasi menjadi APPROVED, dapat %q", req.Status)
	}
	if errs := Validate(req); errs != nil {
		t.Fatalf("status APPROVED seharusnya valid, dapat error: %+v", errs)
	}

	invalid := ModerateReviewRequest{Status: StatusPending}
	if errs := Validate(invalid); !hasFieldError(errs, "status") {
		t.Fatalf("status PENDING seharusnya ditolak, dapat: %+v", errs)
	}
}

func hasFieldError(errs []response.FieldError, field string) bool {
	for _, err := range errs {
		if err.Field == field {
			return true
		}
	}
	return false
}
