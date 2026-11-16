package review

import (
	"context"
	"strings"

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

	review := &Review{
		UserID:        userID,
		BookingID:     &req.BookingID,
		DestinationID: req.DestinationID,
		Rating:        req.Rating,
		Comment:       req.Comment,
		Status:        StatusPending,
		IsApproved:    false,
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
	if review.Status != StatusPending {
		return nil, ErrNotEditable
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
	if review.Status == StatusApproved {
		return ErrNotEditable
	}
	if err := s.repo.Delete(ctx, id); err != nil {
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

func (s *Service) ListApproved(ctx context.Context, page, limit int) ([]Review, int64, error) {
	page, limit = normalizePaging(page, limit)
	items, total, err := s.repo.ListApproved(ctx, (page-1)*limit, limit)
	if err != nil {
		return nil, 0, apperror.Internal(err)
	}
	return items, total, nil
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
