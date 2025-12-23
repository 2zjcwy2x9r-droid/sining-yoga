package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/yoga/knowledge-base/internal/api/handler"
	"github.com/yoga/knowledge-base/internal/api/middleware"
	"github.com/yoga/knowledge-base/internal/config"
	mcpservice "github.com/yoga/knowledge-base/internal/mcp"
	"github.com/yoga/knowledge-base/internal/mcp/tools"
	"github.com/yoga/knowledge-base/internal/repository/postgres"
	aiservice "github.com/yoga/knowledge-base/internal/service/ai"
	"github.com/yoga/knowledge-base/internal/service/booking"
	"github.com/yoga/knowledge-base/internal/service/knowledge"
	"github.com/yoga/knowledge-base/pkg/embedding"
	mcppkg "github.com/yoga/knowledge-base/pkg/mcp"
	"github.com/yoga/knowledge-base/pkg/observability"
	"github.com/yoga/knowledge-base/pkg/openai"
	"github.com/yoga/knowledge-base/pkg/storage"
	"github.com/yoga/knowledge-base/pkg/vector"
	"go.uber.org/zap"
)

func main() {
	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("加载配置失败: %v\n", err)
		os.Exit(1)
	}

	// 初始化日志
	logger, err := config.NewLogger(cfg.Log)
	if err != nil {
		fmt.Printf("初始化日志失败: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	// 初始化追踪
	tp, err := observability.InitTracer("api-server", cfg.Jaeger.Endpoint)
	if err != nil {
		logger.Error("初始化追踪失败", zap.Error(err))
	} else {
		defer func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := tp.Shutdown(ctx); err != nil {
				logger.Error("关闭追踪器失败", zap.Error(err))
			}
		}()
	}

	// 初始化存储
	storageClient, err := storage.NewMinIOStorage(
		cfg.MinIO.Endpoint,
		cfg.MinIO.AccessKeyID,
		cfg.MinIO.SecretAccessKey,
		cfg.MinIO.UseSSL,
	)
	if err != nil {
		logger.Fatal("初始化存储失败", zap.Error(err))
	}

	// 确保存储桶存在
	ctx := context.Background()
	if err := storageClient.EnsureBucket(ctx, cfg.MinIO.BucketName); err != nil {
		logger.Fatal("创建存储桶失败", zap.Error(err))
	}

	// 初始化数据库连接（共享连接）
	db, err := postgres.GetDB(cfg.Database.DSN())
	if err != nil {
		logger.Fatal("初始化数据库失败", zap.Error(err))
	}

	// 初始化仓储
	kbRepo, err := postgres.NewKnowledgeRepository(cfg.Database.DSN())
	if err != nil {
		logger.Fatal("初始化知识库仓储失败", zap.Error(err))
	}
	bookingRepo := postgres.NewBookingRepository(db)

	// 初始化服务客户端
	embeddingClient := embedding.NewClient(cfg.Embedding.URL)
	vectorClient := vector.NewClient(cfg.Vector.URL)

	kbService := knowledge.NewService(kbRepo, storageClient, embeddingClient, vectorClient, cfg.MinIO.BucketName, logger)

	// 初始化定课服务
	bookingService := booking.NewService(bookingRepo, logger)

	// 初始化MCP服务器
	mcpServer := mcppkg.NewServer()
	tools.RegisterBookingTools(mcpServer, bookingService, logger)
	mcpService := mcpservice.NewService(mcpServer, logger)

	// 初始化AI服务
	openAIClient := openai.NewClient(cfg.OpenAI.APIKey, cfg.OpenAI.BaseURL, cfg.OpenAI.Model)
	openAIAdapter := openai.NewAdapter(openAIClient)
	aiRetriever := aiservice.NewRetriever(vectorClient, kbRepo, logger)
	aiService := aiservice.NewService(openAIAdapter, aiRetriever, mcpService, kbRepo, logger)

	// 初始化处理器
	kbHandler := handler.NewKnowledgeHandler(kbService, logger)
	aiHandler := handler.NewAIHandler(aiService, logger)
	bookingHandler := handler.NewBookingHandler(bookingService, logger)

	// 设置Gin
	if cfg.Log.Level != "debug" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.CORSMiddleware()) // CORS支持，允许小程序跨域请求
	router.Use(middleware.TracingMiddleware())
	router.Use(middleware.LoggingMiddleware(logger))

	// API路由
	api := router.Group("/api/v1")
	{
		// 知识库路由
		bases := api.Group("/knowledge-bases")
		{
			bases.POST("", kbHandler.CreateBase)
			bases.GET("", kbHandler.ListBases)
			bases.GET("/:base_id", kbHandler.GetBase)
			bases.PUT("/:base_id", kbHandler.UpdateBase)
			bases.DELETE("/:base_id", kbHandler.DeleteBase)

			// 知识项路由
			items := bases.Group("/:base_id/items")
			{
				items.POST("/text", kbHandler.CreateTextItem)
				items.POST("/file", kbHandler.CreateFileItem)
				items.GET("", kbHandler.ListItems)
				items.GET("/:id", kbHandler.GetItem)
				items.DELETE("/:id", kbHandler.DeleteItem)
			}
		}

		// AI问答路由
		ai := api.Group("/ai")
		{
			ai.POST("/chat", aiHandler.Chat)
		}

		// 课程和预订路由
		classes := api.Group("/classes")
		{
			classes.GET("", bookingHandler.ListClasses)
			classes.GET("/:id", bookingHandler.GetClass)
			classes.POST("/:id/book", bookingHandler.BookClass)
			classes.DELETE("/bookings/:id", bookingHandler.CancelBooking)
			classes.GET("/bookings", bookingHandler.ListUserBookings)
			// 评价路由
			classes.POST("/:id/reviews", bookingHandler.CreateReview)
			classes.GET("/:id/reviews", bookingHandler.ListClassReviews)
		}
	}

	// 健康检查
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// 启动服务器
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	// 优雅关闭
	go func() {
		logger.Info("服务器启动", zap.String("address", addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("服务器启动失败", zap.Error(err))
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("正在关闭服务器...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("服务器关闭失败", zap.Error(err))
	}

	logger.Info("服务器已关闭")
}
