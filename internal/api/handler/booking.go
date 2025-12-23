package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/yoga/knowledge-base/internal/service/booking"
	"go.uber.org/zap"
)

// BookingHandler 预订处理器
type BookingHandler struct {
	service *booking.Service
	logger  *zap.Logger
}

// NewBookingHandler 创建预订处理器
func NewBookingHandler(service *booking.Service, logger *zap.Logger) *BookingHandler {
	return &BookingHandler{
		service: service,
		logger:  logger,
	}
}

// ListClasses 列出课程
func (h *BookingHandler) ListClasses(c *gin.Context) {
	// 获取查询参数
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	var startTime, endTime *time.Time

	// 如果没有指定日期，默认查询本周
	if startDate == "" && endDate == "" {
		now := time.Now()
		// 本周一
		weekday := int(now.Weekday())
		if weekday == 0 {
			weekday = 7
		}
		monday := now.AddDate(0, 0, -weekday+1)
		monday = time.Date(monday.Year(), monday.Month(), monday.Day(), 0, 0, 0, 0, monday.Location())
		startTime = &monday
		// 本周日
		sunday := monday.AddDate(0, 0, 6)
		sunday = time.Date(sunday.Year(), sunday.Month(), sunday.Day(), 23, 59, 59, 0, sunday.Location())
		endTime = &sunday
	} else {
		if startDate != "" {
			t, err := time.Parse("2006-01-02", startDate)
			if err == nil {
				t = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
				startTime = &t
			}
		}
		if endDate != "" {
			t, err := time.Parse("2006-01-02", endDate)
			if err == nil {
				t = time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 0, t.Location())
				endTime = &t
			}
		}
	}

	classes, err := h.service.ListClasses(c.Request.Context(), startTime, endTime, limit, offset)
	if err != nil {
		h.logger.Error("查询课程列表失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询课程列表失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"items": classes,
		"limit": limit,
		"offset": offset,
	})
}

// GetClass 获取课程详情
func (h *BookingHandler) GetClass(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的课程ID"})
		return
	}

	class, err := h.service.GetClass(c.Request.Context(), id)
	if err != nil {
		h.logger.Error("查询课程失败", zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "课程不存在"})
		return
	}

	c.JSON(http.StatusOK, class)
}

// BookClass 预订课程
func (h *BookingHandler) BookClass(c *gin.Context) {
	var req struct {
		UserID   string `json:"user_id" binding:"required"`
		UserName string `json:"user_name"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	classID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的课程ID"})
		return
	}

	booking, err := h.service.BookClass(c.Request.Context(), classID, req.UserID, req.UserName)
	if err != nil {
		h.logger.Error("预订课程失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "预订课程失败"})
		return
	}

	c.JSON(http.StatusCreated, booking)
}

// CancelBooking 取消预订
func (h *BookingHandler) CancelBooking(c *gin.Context) {
	bookingID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的预订ID"})
		return
	}

	if err := h.service.CancelBooking(c.Request.Context(), bookingID); err != nil {
		h.logger.Error("取消预订失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "取消预订失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "取消预订成功"})
}

// ListUserBookings 列出用户预订
func (h *BookingHandler) ListUserBookings(c *gin.Context) {
	userID := c.Query("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id参数必需"})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	bookings, err := h.service.ListUserBookings(c.Request.Context(), userID, limit, offset)
	if err != nil {
		h.logger.Error("查询用户预订失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询用户预订失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"items": bookings,
		"limit": limit,
		"offset": offset,
	})
}

// CreateReview 创建评价
func (h *BookingHandler) CreateReview(c *gin.Context) {
	var req struct {
		UserID   string   `json:"user_id" binding:"required"`
		UserName string   `json:"user_name"`
		Rating   int      `json:"rating" binding:"required,min=1,max=5"`
		Content  string   `json:"content"`
		Images   []string `json:"images"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	classID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的课程ID"})
		return
	}

	review, err := h.service.CreateReview(c.Request.Context(), classID, req.UserID, req.UserName, req.Rating, req.Content, req.Images)
	if err != nil {
		h.logger.Error("创建评价失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建评价失败"})
		return
	}

	c.JSON(http.StatusCreated, review)
}

// ListClassReviews 列出课程评价
func (h *BookingHandler) ListClassReviews(c *gin.Context) {
	classID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的课程ID"})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	reviews, err := h.service.ListClassReviews(c.Request.Context(), classID, limit, offset)
	if err != nil {
		h.logger.Error("查询课程评价失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询课程评价失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"items": reviews,
		"limit": limit,
		"offset": offset,
	})
}

