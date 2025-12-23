package booking

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/yoga/knowledge-base/internal/domain/booking"
	"github.com/yoga/knowledge-base/pkg/observability"
	"go.uber.org/zap"
)

// Repository 预订仓储接口
type Repository interface {
	CreateClass(ctx context.Context, class *booking.Class) error
	GetClass(ctx context.Context, id uuid.UUID) (*booking.Class, error)
	ListClasses(ctx context.Context, startTime, endTime *time.Time, limit, offset int) ([]*booking.Class, error)
	UpdateClass(ctx context.Context, class *booking.Class) error
	CreateBooking(ctx context.Context, b *booking.Booking) error
	GetBooking(ctx context.Context, id uuid.UUID) (*booking.Booking, error)
	ListUserBookings(ctx context.Context, userID string, limit, offset int) ([]*booking.Booking, error)
	CancelBooking(ctx context.Context, id uuid.UUID) error
	CreateReview(ctx context.Context, review *booking.Review) error
	ListClassReviews(ctx context.Context, classID uuid.UUID, limit, offset int) ([]*booking.Review, error)
	GetUserReview(ctx context.Context, classID uuid.UUID, userID string) (*booking.Review, error)
}

// Service 定课服务
type Service struct {
	repo   Repository
	logger *zap.Logger
}

// NewService 创建定课服务
func NewService(repo Repository, logger *zap.Logger) *Service {
	return &Service{
		repo:   repo,
		logger: logger,
	}
}

// CreateClass 创建课程
func (s *Service) CreateClass(ctx context.Context, name, description, instructor string, startTime, endTime time.Time, capacity int) (*booking.Class, error) {
	ctx, span := observability.StartSpan(ctx, "booking-service", "CreateClass")
	defer span.End()

	class := &booking.Class{
		Name:        name,
		Description: description,
		Instructor:  instructor,
		StartTime:   startTime,
		EndTime:     endTime,
		Capacity:    capacity,
		Status:      "scheduled",
	}

	if err := s.repo.CreateClass(ctx, class); err != nil {
		return nil, fmt.Errorf("创建课程失败: %w", err)
	}

	return class, nil
}

// GetClass 获取课程
func (s *Service) GetClass(ctx context.Context, id uuid.UUID) (*booking.Class, error) {
	ctx, span := observability.StartSpan(ctx, "booking-service", "GetClass")
	defer span.End()

	return s.repo.GetClass(ctx, id)
}

// ListClasses 列出课程
func (s *Service) ListClasses(ctx context.Context, startTime, endTime *time.Time, limit, offset int) ([]*booking.Class, error) {
	ctx, span := observability.StartSpan(ctx, "booking-service", "ListClasses")
	defer span.End()

	return s.repo.ListClasses(ctx, startTime, endTime, limit, offset)
}

// BookClass 预订课程
func (s *Service) BookClass(ctx context.Context, classID uuid.UUID, userID, userName string) (*booking.Booking, error) {
	ctx, span := observability.StartSpan(ctx, "booking-service", "BookClass")
	defer span.End()

	b := &booking.Booking{
		ClassID:  classID,
		UserID:   userID,
		UserName: userName,
		Status:   "confirmed",
	}

	if err := s.repo.CreateBooking(ctx, b); err != nil {
		return nil, fmt.Errorf("预订课程失败: %w", err)
	}

	return b, nil
}

// CancelBooking 取消预订
func (s *Service) CancelBooking(ctx context.Context, bookingID uuid.UUID) error {
	ctx, span := observability.StartSpan(ctx, "booking-service", "CancelBooking")
	defer span.End()

	return s.repo.CancelBooking(ctx, bookingID)
}

// ListUserBookings 列出用户预订
func (s *Service) ListUserBookings(ctx context.Context, userID string, limit, offset int) ([]*booking.Booking, error) {
	ctx, span := observability.StartSpan(ctx, "booking-service", "ListUserBookings")
	defer span.End()

	return s.repo.ListUserBookings(ctx, userID, limit, offset)
}

// CreateReview 创建评价
func (s *Service) CreateReview(ctx context.Context, classID uuid.UUID, userID, userName string, rating int, content string, images []string) (*booking.Review, error) {
	ctx, span := observability.StartSpan(ctx, "booking-service", "CreateReview")
	defer span.End()

	review := &booking.Review{
		ClassID:  classID,
		UserID:   userID,
		UserName: userName,
		Rating:   rating,
		Content:  content,
		Images:   images,
	}

	if !review.IsValidRating() {
		return nil, fmt.Errorf("评分必须在1-5之间")
	}

	if err := s.repo.CreateReview(ctx, review); err != nil {
		return nil, fmt.Errorf("创建评价失败: %w", err)
	}

	return review, nil
}

// ListClassReviews 列出课程评价
func (s *Service) ListClassReviews(ctx context.Context, classID uuid.UUID, limit, offset int) ([]*booking.Review, error) {
	ctx, span := observability.StartSpan(ctx, "booking-service", "ListClassReviews")
	defer span.End()

	return s.repo.ListClassReviews(ctx, classID, limit, offset)
}

// GetUserReview 获取用户对课程的评价
func (s *Service) GetUserReview(ctx context.Context, classID uuid.UUID, userID string) (*booking.Review, error) {
	ctx, span := observability.StartSpan(ctx, "booking-service", "GetUserReview")
	defer span.End()

	return s.repo.GetUserReview(ctx, classID, userID)
}
