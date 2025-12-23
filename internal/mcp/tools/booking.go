package tools

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/yoga/knowledge-base/internal/service/booking"
	"github.com/yoga/knowledge-base/pkg/mcp"
	"go.uber.org/zap"
)

// RegisterBookingTools 注册定课相关工具
func RegisterBookingTools(server mcp.Server, bookingSvc *booking.Service, logger *zap.Logger) {
	// 查询课程表工具
	server.RegisterTool(mcp.Tool{
		Name:        "query_schedule",
		Description: "查询课程表，可以按日期范围查询",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"start_date": map[string]interface{}{
					"type":        "string",
					"description": "开始日期，格式：YYYY-MM-DD",
				},
				"end_date": map[string]interface{}{
					"type":        "string",
					"description": "结束日期，格式：YYYY-MM-DD",
				},
			},
		},
	}, func(args map[string]interface{}) (interface{}, error) {
		var startTime, endTime *time.Time

		if startDateStr, ok := args["start_date"].(string); ok && startDateStr != "" {
			t, err := time.Parse("2006-01-02", startDateStr)
			if err != nil {
				return nil, fmt.Errorf("无效的开始日期格式: %w", err)
			}
			startTime = &t
		}

		if endDateStr, ok := args["end_date"].(string); ok && endDateStr != "" {
			t, err := time.Parse("2006-01-02", endDateStr)
			if err != nil {
				return nil, fmt.Errorf("无效的结束日期格式: %w", err)
			}
			// 设置为当天的23:59:59
			t = time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 0, t.Location())
			endTime = &t
		}

		classes, err := bookingSvc.ListClasses(nil, startTime, endTime, 50, 0)
		if err != nil {
			return nil, fmt.Errorf("查询课程表失败: %w", err)
		}

		// 转换为JSON友好的格式
		result := make([]map[string]interface{}, len(classes))
		for i, class := range classes {
			result[i] = map[string]interface{}{
				"id":          class.ID.String(),
				"name":        class.Name,
				"description": class.Description,
				"instructor":  class.Instructor,
				"start_time":  class.StartTime.Format(time.RFC3339),
				"end_time":    class.EndTime.Format(time.RFC3339),
				"capacity":    class.Capacity,
				"booked_count": class.BookedCount,
				"available":   class.IsAvailable(),
			}
		}

		return result, nil
	})

	// 预订课程工具
	server.RegisterTool(mcp.Tool{
		Name:        "book_class",
		Description: "预订课程",
		Parameters: map[string]interface{}{
			"type": "object",
			"required": []string{"class_id", "user_id"},
			"properties": map[string]interface{}{
				"class_id": map[string]interface{}{
					"type":        "string",
					"description": "课程ID",
				},
				"user_id": map[string]interface{}{
					"type":        "string",
					"description": "用户ID（微信用户ID）",
				},
				"user_name": map[string]interface{}{
					"type":        "string",
					"description": "用户名称",
				},
			},
		},
	}, func(args map[string]interface{}) (interface{}, error) {
		classIDStr, ok := args["class_id"].(string)
		if !ok {
			return nil, fmt.Errorf("class_id 是必需的")
		}

		classID, err := uuid.Parse(classIDStr)
		if err != nil {
			return nil, fmt.Errorf("无效的课程ID: %w", err)
		}

		userID, ok := args["user_id"].(string)
		if !ok {
			return nil, fmt.Errorf("user_id 是必需的")
		}

		userName, _ := args["user_name"].(string)

		booking, err := bookingSvc.BookClass(nil, classID, userID, userName)
		if err != nil {
			return nil, fmt.Errorf("预订课程失败: %w", err)
		}

		return map[string]interface{}{
			"booking_id": booking.ID.String(),
			"class_id":   booking.ClassID.String(),
			"status":     booking.Status,
			"message":    "预订成功",
		}, nil
	})

	// 取消预订工具
	server.RegisterTool(mcp.Tool{
		Name:        "cancel_booking",
		Description: "取消预订",
		Parameters: map[string]interface{}{
			"type": "object",
			"required": []string{"booking_id"},
			"properties": map[string]interface{}{
				"booking_id": map[string]interface{}{
					"type":        "string",
					"description": "预订ID",
				},
			},
		},
	}, func(args map[string]interface{}) (interface{}, error) {
		bookingIDStr, ok := args["booking_id"].(string)
		if !ok {
			return nil, fmt.Errorf("booking_id 是必需的")
		}

		bookingID, err := uuid.Parse(bookingIDStr)
		if err != nil {
			return nil, fmt.Errorf("无效的预订ID: %w", err)
		}

		if err := bookingSvc.CancelBooking(nil, bookingID); err != nil {
			return nil, fmt.Errorf("取消预订失败: %w", err)
		}

		return map[string]interface{}{
			"booking_id": bookingIDStr,
			"message":    "取消预订成功",
		}, nil
	})

	// 查询用户预订
	server.RegisterTool(mcp.Tool{
		Name:        "query_user_bookings",
		Description: "查询用户的预订列表",
		Parameters: map[string]interface{}{
			"type": "object",
			"required": []string{"user_id"},
			"properties": map[string]interface{}{
				"user_id": map[string]interface{}{
					"type":        "string",
					"description": "用户ID（微信用户ID）",
				},
			},
		},
	}, func(args map[string]interface{}) (interface{}, error) {
		userID, ok := args["user_id"].(string)
		if !ok {
			return nil, fmt.Errorf("user_id 是必需的")
		}

		bookings, err := bookingSvc.ListUserBookings(nil, userID, 50, 0)
		if err != nil {
			return nil, fmt.Errorf("查询用户预订失败: %w", err)
		}

		result := make([]map[string]interface{}, len(bookings))
		for i, b := range bookings {
			result[i] = map[string]interface{}{
				"booking_id": b.ID.String(),
				"class_id":   b.ClassID.String(),
				"status":     b.Status,
				"created_at": b.CreatedAt.Format(time.RFC3339),
			}
		}

		return result, nil
	})
}

