package review

import (
	"context"
	"strings"
	"time"

	"github.com/KawahPutih/backend/pkg/apperror"
)

var eligibleBookingStatuses = map[string]struct{}{
	"PAID":      {},
	"CONFIRMED": {},
	"USED":      {},
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, userID string, req CreateReviewRequest) (*Review, error) {
	req.Normalize()
	if userID == "" {
		return nil, apperror.Unauthorized("Silakan login terlebih dahulu.")
	}
	if err := ValidateBookingID(req.BookingID); err != nil {
		return nil, err
	}
	if err := ValidateRating(req.Rating); err != nil {
		return nil, err
	}
	if err := ValidateComment(req.Comment); err != nil {
		return nil, err
	}

	if req.DestinationID != "" {
		exists, err := s.repo.DestinationExists(ctx, req.DestinationID)
		if err != nil {
			return nil, apperror.Internal(err)
		}
		if !exists {
			return nil, ErrDestinationNotFound
		}
	}

	booking, err := s.repo.FindBooking(ctx, req.BookingID)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	if booking == nil {
		return nil, ErrBookingNotFound
	}
	if booking.UserID != userID {
		return nil, ErrBookingNotOwned
	}
	if _, ok := eligibleBookingStatuses[strings.ToUpper(booking.Status)]; !ok {
		return nil, ErrBookingNotEligible
	}

	existing, err := s.repo.FindByBookingID(ctx, req.BookingID)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	if existing != nil {
		return nil, ErrAlreadyReviewed
	}

	deleted, err := s.repo.FindByBookingIDUnscoped(ctx, req.BookingID)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	if deleted != nil {
		if !deleted.DeletedAt.Valid {
			return nil, ErrAlreadyReviewed
		}
		deleted.UserID = userID
		deleted.DestinationID = req.DestinationID
		deleted.Rating = req.Rating
		deleted.Comment = req.Comment
		deleted.Status = StatusApproved
		deleted.IsApproved = true
		deleted.CreatedAt = time.Now()
		if err := s.repo.Restore(ctx, deleted); err != nil {
			return nil, apperror.Internal(err)
		}
		return deleted, nil
	}

	review := &Review{
		UserID:        userID,
		BookingID:     &req.BookingID,
		DestinationID: req.DestinationID,
		Rating:        req.Rating,
		Comment:       req.Comment,
		Status:        StatusApproved,
		IsApproved:    true,
	}
	if err := s.repo.Create(ctx, review); err != nil {
		return nil, apperror.Internal(err)
	}
	return review, nil
}

func (s *Service) Update(ctx context.Context, userID, id string, req UpdateReviewRequest) (*Review, error) {
	req.Normalize()
	if err := ValidateRating(req.Rating); err != nil {
		return nil, err
	}
	if err := ValidateComment(req.Comment); err != nil {
		return nil, err
	}

	review, err := s.findOwned(ctx, userID, id)
	if err != nil {
		return nil, err
	}

	review.Rating = req.Rating
	review.Comment = req.Comment
	if err := s.repo.Update(ctx, review); err != nil {
		return nil, apperror.Internal(err)
	}
	return review, nil
}

func (s *Service) Delete(ctx context.Context, userID, id string) error {
	review, err := s.findOwned(ctx, userID, id)
	if err != nil {
		return err
	}
	if err := s.repo.Delete(ctx, review.ID); err != nil {
		return apperror.Internal(err)
	}
	return nil
}

func (s *Service) Moderate(ctx context.Context, id string, req ModerateReviewRequest) (*Review, error) {
	status := strings.ToUpper(strings.TrimSpace(req.Status))
	if err := ValidateStatus(status); err != nil {
		return nil, err
	}

	review, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	if review == nil {
		return nil, ErrNotFound
	}

	review.Status = status
	review.IsApproved = status == StatusApproved
	if err := s.repo.Update(ctx, review); err != nil {
		return nil, apperror.Internal(err)
	}
	return review, nil
}

func (s *Service) ListMine(ctx context.Context, userID string, page, limit int) ([]Review, int64, error) {
	page, limit = normalizePaging(page, limit)
	items, total, err := s.repo.ListByUser(ctx, userID, (page-1)*limit, limit)
	if err != nil {
		return nil, 0, apperror.Internal(err)
	}
	return items, total, nil
}

func (s *Service) ListAll(ctx context.Context, status string, page, limit int) ([]Review, int64, error) {
	status = strings.ToUpper(strings.TrimSpace(status))
	if status != "" {
		if status != StatusPending && status != StatusApproved && status != StatusRejected {
			return nil, 0, ErrInvalidStatus
		}
	}
	page, limit = normalizePaging(page, limit)
	items, total, err := s.repo.ListAll(ctx, status, (page-1)*limit, limit)
	if err != nil {
		return nil, 0, apperror.Internal(err)
	}
	return items, total, nil
}

// ListApproved mendukung filter rating 1-5 dan urutan tertentu.
// rating = 0 berarti semua rating; sort default "newest".
func (s *Service) ListApproved(ctx context.Context, rating int, sort string, page, limit int) ([]Review, int64, error) {
	if rating < 0 || rating > 5 {
		return nil, 0, ErrInvalidRatingFilter
	}

	sort = strings.ToLower(strings.TrimSpace(sort))
	switch sort {
	case "":
		sort = SortNewest
	case SortNewest, SortRatingDesc, SortRatingAsc:
	default:
		return nil, 0, ErrInvalidSort
	}

	page, limit = normalizePaging(page, limit)
	items, total, err := s.repo.ListApproved(ctx, ListApprovedOptions{
		Rating: rating,
		Sort:   sort,
	}, (page-1)*limit, limit)
	if err != nil {
		return nil, 0, apperror.Internal(err)
	}
	return items, total, nil
}

func (s *Service) Summary(ctx context.Context) (float64, int64, error) {
	average, total, err := s.repo.Summary(ctx)
	if err != nil {
		return 0, 0, apperror.Internal(err)
	}
	return average, total, nil
}

func (s *Service) findOwned(ctx context.Context, userID, id string) (*Review, error) {
	review, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	if review == nil {
		return nil, ErrNotFound
	}
	if review.UserID != userID {
		return nil, ErrForbidden
	}
	return review, nil
}

func normalizePaging(page, limit int) (int, int) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	return page, limit
}
