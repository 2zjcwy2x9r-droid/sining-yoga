package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/yoga/knowledge-base/internal/domain/ai"
	aiservice "github.com/yoga/knowledge-base/internal/service/ai"
	"go.uber.org/zap"
)

// contains 检查字符串是否包含子串（不区分大小写）
func contains(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}

// AIHandler AI处理器
type AIHandler struct {
	service *aiservice.Service
	logger  *zap.Logger
}

// NewAIHandler 创建AI处理器
func NewAIHandler(service *aiservice.Service, logger *zap.Logger) *AIHandler {
	return &AIHandler{
		service: service,
		logger:  logger,
	}
}

// Chat 处理聊天请求
func (h *AIHandler) Chat(c *gin.Context) {
	var req ai.ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := h.service.Chat(c.Request.Context(), req)
	if err != nil {
		h.logger.Error("AI聊天失败", zap.Error(err))
		
		// 根据错误类型返回不同的状态码和错误信息
		errorMsg := err.Error()
		statusCode := http.StatusInternalServerError
		
		// 检查是否是余额不足或API密钥问题
		if contains(errorMsg, "余额不足") || contains(errorMsg, "Insufficient Balance") {
			statusCode = http.StatusPaymentRequired // 402
			errorMsg = "AI服务账户余额不足，请联系管理员充值"
		} else if contains(errorMsg, "API密钥无效") || contains(errorMsg, "401") {
			statusCode = http.StatusUnauthorized // 401
			errorMsg = "AI服务配置错误，请联系管理员"
		} else if contains(errorMsg, "请求频率过高") || contains(errorMsg, "429") {
			statusCode = http.StatusTooManyRequests // 429
			errorMsg = "请求过于频繁，请稍后重试"
		}
		
		c.JSON(statusCode, gin.H{
			"error": errorMsg,
			"detail": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

