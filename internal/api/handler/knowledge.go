package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/yoga/knowledge-base/internal/service/knowledge"
	"go.uber.org/zap"
)

// KnowledgeHandler 知识库处理器
type KnowledgeHandler struct {
	service *knowledge.Service
	logger  *zap.Logger
}

// NewKnowledgeHandler 创建知识库处理器
func NewKnowledgeHandler(service *knowledge.Service, logger *zap.Logger) *KnowledgeHandler {
	return &KnowledgeHandler{
		service: service,
		logger:  logger,
	}
}

// CreateBase 创建知识库
func (h *KnowledgeHandler) CreateBase(c *gin.Context) {
	var req struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
		Type        string `json:"type" binding:"required,oneof=text image video mixed"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	base, err := h.service.CreateBase(c.Request.Context(), req.Name, req.Description, req.Type)
	if err != nil {
		h.logger.Error("创建知识库失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建知识库失败"})
		return
	}

	c.JSON(http.StatusCreated, base)
}

// GetBase 获取知识库
func (h *KnowledgeHandler) GetBase(c *gin.Context) {
	id, err := uuid.Parse(c.Param("base_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	base, err := h.service.GetBase(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "知识库不存在"})
		return
	}

	c.JSON(http.StatusOK, base)
}

// ListBases 列出知识库
func (h *KnowledgeHandler) ListBases(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	if limit <= 0 || limit > 100 {
		limit = 20
	}

	bases, err := h.service.ListBases(c.Request.Context(), limit, offset)
	if err != nil {
		h.logger.Error("查询知识库列表失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"items": bases, "limit": limit, "offset": offset})
}

// UpdateBase 更新知识库
func (h *KnowledgeHandler) UpdateBase(c *gin.Context) {
	id, err := uuid.Parse(c.Param("base_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	var req struct {
		Name        *string `json:"name"`
		Description *string `json:"description"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	base, err := h.service.UpdateBase(c.Request.Context(), id, req.Name, req.Description)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新失败"})
		return
	}

	c.JSON(http.StatusOK, base)
}

// DeleteBase 删除知识库
func (h *KnowledgeHandler) DeleteBase(c *gin.Context) {
	id, err := uuid.Parse(c.Param("base_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	if err := h.service.DeleteBase(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

// CreateTextItem 创建文本知识项
func (h *KnowledgeHandler) CreateTextItem(c *gin.Context) {
	baseID, err := uuid.Parse(c.Param("base_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的知识库ID"})
		return
	}

	var req struct {
		Title   string `json:"title" binding:"required"`
		Content string `json:"content" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	item, err := h.service.CreateTextItem(c.Request.Context(), baseID, req.Title, req.Content)
	if err != nil {
		h.logger.Error("创建知识项失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建失败"})
		return
	}

	c.JSON(http.StatusCreated, item)
}

// CreateFileItem 创建文件知识项
func (h *KnowledgeHandler) CreateFileItem(c *gin.Context) {
	baseID, err := uuid.Parse(c.Param("base_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的知识库ID"})
		return
	}

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "文件上传失败"})
		return
	}
	defer file.Close()

	title := c.PostForm("title")
	if title == "" {
		title = header.Filename
	}

	item, err := h.service.CreateFileItem(
		c.Request.Context(),
		baseID,
		title,
		file,
		header.Size,
		header.Header.Get("Content-Type"),
		header.Filename,
	)
	if err != nil {
		h.logger.Error("创建文件知识项失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建失败"})
		return
	}

	c.JSON(http.StatusCreated, item)
}

// ListItems 列出知识项
func (h *KnowledgeHandler) ListItems(c *gin.Context) {
	baseID, err := uuid.Parse(c.Param("base_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的知识库ID"})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	if limit <= 0 || limit > 100 {
		limit = 20
	}

	items, err := h.service.ListItems(c.Request.Context(), baseID, limit, offset)
	if err != nil {
		h.logger.Error("查询知识项列表失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"items": items, "limit": limit, "offset": offset})
}

// GetItem 获取知识项
func (h *KnowledgeHandler) GetItem(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	item, err := h.service.GetItem(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "知识项不存在"})
		return
	}

	c.JSON(http.StatusOK, item)
}

// DeleteItem 删除知识项
func (h *KnowledgeHandler) DeleteItem(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	if err := h.service.DeleteItem(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

