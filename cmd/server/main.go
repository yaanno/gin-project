package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog"

	"github.com/yourusername/user-management-api/internal/config"
	"github.com/yourusername/user-management-api/internal/database/sqlite-gorm"
	"github.com/yourusername/user-management-api/internal/handlers"
	"github.com/yourusername/user-management-api/internal/middleware"
	"github.com/yourusername/user-management-api/internal/repository"
	"github.com/yourusername/user-management-api/internal/services"
	"github.com/yourusername/user-management-api/pkg/authentication"
	"github.com/yourusername/user-management-api/pkg/logger"
	"github.com/yourusername/user-management-api/pkg/token"
)

func main() {
	// Set Gin to production mode
	gin.SetMode(gin.ReleaseMode)

	// Create logger
	log := logger.New(logger.Config{
		Level:       "info",
		FilePath:    "logs/app.log",
		MaxSize:     50,
		MaxBackups:  5,
		EnableFile:  true,
		Development: true,
	})

	// Set global logger
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Warn().Err(err).Msg("Error loading .env file")
	}

	databaseConfig := sqlite.DatabaseConfig{
		Path:            "./data/users.db",
		InMemory:        true,
		MaxOpenConns:    10,
		MaxIdleConns:    5,
		ConnMaxLifetime: 5 * time.Minute,
		ConnMaxIdleTime: 30 * time.Second,
	}

	// Initialize configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load configuration")
	}

	// Initialize database connection
	// db, err := sqlite.NewSQLiteDatabase(sqliteConfig)
	db, err := sqlite.InitializeDatabase(databaseConfig)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to database")
	}

	// inject to repository
	userRepository := repository.NewUserRepository(db, log)
	// inject to service
	userService := services.NewUserService(userRepository, log)
	// inject to auth service
	loginAttemptRepository := repository.NewLoginAttemptRepository(db, log)
	tokenManager := token.NewTokenManager(cfg.JWTSecret, cfg.JWTRefreshSecret)
	authManager := authentication.NewAuthenticationManager(userRepository, tokenManager, loginAttemptRepository, log)
	authService := services.NewAuthService(tokenManager, authManager, userRepository, log)
	// inject to handler
	userHandler := handlers.NewUserHandler(userService, log)
	authHandler := handlers.NewAuthHandler(authService, log)

	// Setup Gin router
	router := gin.New()

	// Middleware
	router.Use(gin.Recovery())

	// Add error middleware.
	// At this point the middleware will check if there are any errors in the context and respond accordingly
	// TODO: separate this from a proper error package
	router.Use(middleware.ErrorMiddleware(log))

	// Not found handler
	router.NoRoute(middleware.HandleNotFound(log))

	// Add sanitization middleware
	router.Use(middleware.SanitizationMiddleware())

	// Add ip based rate limit middleware
	router.Use(middleware.IPRateLimitMiddleware(cfg.RateLimitLimit, cfg.RateLimitBurst, cfg.RateLimitDuration, log))

	// Api routes
	v1Group := router.Group("/api/v1")
	{
		// Authentication routes
		authGroup := v1Group.Group("/auth")
		{
			authGroup.POST("/register", authHandler.RegisterUser)
			authGroup.POST("/login", authHandler.LoginUser)
			authGroup.POST("/refresh", authHandler.RefreshTokens)
		}
		// User routes (protected)
		userGroup := v1Group.Group("/users")
		// Add Auth middleware
		userGroup.Use(middleware.AuthMiddleware(authManager, log))
		{
			userGroup.GET("/", userHandler.GetAllUsers)
			userGroup.GET("/:id", userHandler.GetUserByID)
			userGroup.PUT("/:id", userHandler.UpdateUser)
			userGroup.DELETE("/:id", userHandler.DeleteUser)
		}
	}

	// Get port from environment or use default
	port := cfg.ServerPort
	if port == "" {
		port = "8080"
	}

	address := cfg.ServerAddress
	if address == "" {
		address = "localhost"
	}

	srv := http.Server{
		Addr:    address + ":" + port,
		Handler: router.Handler(),
	}

	go func() {
		log.Info().Str("address", address).Str("port", port).Msg("Starting server")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("Server failed to start")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error().Err(err).Msg("Failed to shutdown server")
	}

	select {
	case <-ctx.Done():
		log.Error().Err(ctx.Err()).Msg("Timeout shutting down server")
	default:
		log.Info().Msg("Server shut down")
	}
}
