package app

import (
	"fmt"
	"log/slog"
	"os"
	"ticketprocessing/internal/api"
	"ticketprocessing/internal/auth"
	"ticketprocessing/internal/config"
	"ticketprocessing/internal/db"
	"ticketprocessing/internal/messaging"
	"ticketprocessing/internal/models"
	"ticketprocessing/internal/repository"
	"ticketprocessing/internal/service"
	"time"

	"github.com/go-redis/redis"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func NewApp() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	cfg, err := config.LoadConfig()
	if err != nil {
		slog.Warn("failed to load config", slog.String("error", err.Error()))
		return
	}

	log.Debug("load config", slog.Any("cfg", cfg))

	// Initialize Redis
	cli := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
		Password: "",
		DB:       cfg.Redis.DB,
	})
	status, err := cli.Ping().Result()
	if err != nil {
		slog.Warn("failed to ping redis", slog.String("error", err.Error()))
		return
	}
	log.Debug("redis started", slog.Any("status", status))

	// Initialize PostgreSQL
	db, err := db.InitPostgres(&cfg.Postgres)
	if err != nil {
		slog.Warn("failed to init postgres", slog.String("error", err.Error()))
		return
	}

	if err := db.AutoMigrate(&models.User{}, &models.RentalRequest{}, &models.RequestStatusLog{}); err != nil {
		slog.Warn("failed to migrate postgres", slog.String("error", err.Error()))
		return
	}

	// Initialize repositories
	authRepo := repository.NewAuthRepository(db)
	rentalRequestRepo := repository.NewRentalRequestRepository(db)
	requestStatusLogRepo := repository.NewRequestStatusLogRepository(db)
	equipmentRepo := repository.NewEquipmentRepository(db)

	// Initialize RabbitMQ
	rabbitMQ, err := messaging.NewRabbitMQPublisher(&cfg.RabbitMQ)
	if err != nil {
		slog.Warn("failed to initialize RabbitMQ", slog.String("error", err.Error()))
		return
	}
	defer rabbitMQ.Close()

	// Initialize auth components
	jwtManager := auth.NewJWTManager(cfg.JWT.Secret, cfg.JWT.TTLMinutes)
	redisAddr := fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port)
	redisStore := auth.NewRedisTokenStore(redisAddr, cfg.Redis.DB, time.Duration(cfg.JWT.TTLMinutes)*time.Minute)

	// Initialize services
	authService := service.NewAuthService(authRepo, jwtManager, redisStore)
	rentalRequestService := service.NewRentalRequestService(rentalRequestRepo, requestStatusLogRepo, equipmentRepo, rabbitMQ)

	// Initialize Echo
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	// Initialize handlers
	userHandler := api.NewUserHandler(authService)
	rentalRequestHandler := api.NewRentalRequestHandler(rentalRequestService)

	// Public routes
	e.POST("/register", userHandler.Register)
	e.POST("/login", userHandler.Login)

	// Protected routes
	me := e.Group("/me")
	me.Use(api.AuthMiddleware(jwtManager, redisStore))
	me.GET("", userHandler.Me)

	// Rental request routes
	rental := e.Group("/rental_request")
	rental.Use(api.AuthMiddleware(jwtManager, redisStore))
	rental.POST("", rentalRequestHandler.CreateRentalRequest)
	rental.GET("/:id/status", rentalRequestHandler.GetRequestStatus)
	rental.GET("/:id/status_at", rentalRequestHandler.GetRequestStatusAt)

	// Start server
	port := cfg.App.Port
	if port == 0 {
		port = 8080
	}
	addr := fmt.Sprintf(":%d", port)
	log.Info("Server started at", slog.String("addr", addr))
	if err := e.Start(addr); err != nil {
		log.Error("failed to start server", slog.String("error", err.Error()))
	}
}
