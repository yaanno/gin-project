package main

import (
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog"

	"github.com/yourusername/user-management-api/internal/config"
	"github.com/yourusername/user-management-api/internal/database/sqlite"
	"github.com/yourusername/user-management-api/internal/handlers"
	"github.com/yourusername/user-management-api/internal/middleware"
	"github.com/yourusername/user-management-api/internal/repository"
	"github.com/yourusername/user-management-api/internal/services"
	"github.com/yourusername/user-management-api/pkg/logger"
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
		log.Fatal().Err(err).Msg("Error loading .env file")
	}

	sqliteConfig := sqlite.SQLiteConfig{
		Path:     "./data/users.db",
		InMemory: true,
	}

	// Initialize configuration
	_, err := config.LoadConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load configuration")
	}

	// Initialize database connection
	db, err := sqlite.NewSQLiteDatabase(sqliteConfig)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to database")
	}
	defer db.Close()

	// Run migrations
	if err := db.RunSQLiteMigrations(); err != nil {
		log.Fatal().Err(err).Msg("Failed to run migrations")
	}

	// inject to repository
	userRepository := repository.NewUserRepository(db, log)
	// inject to service
	userService := services.NewUserService(userRepository, log)
	authService := services.NewAuthService(userRepository, log)
	// inject to handler
	userHandler := handlers.NewUserHandler(userService, log)
	authHandler := handlers.NewAuthHandler(authService, log)

	// Setup Gin router
	router := gin.New()

	// Middleware
	router.Use(gin.Recovery())
	router.Use(middleware.ErrorHandler())

	// Not found handler
	router.NoRoute(middleware.HandleNotFound)

	// Authentication routes
	authGroup := router.Group("/auth")
	{
		authGroup.POST("/register", authHandler.RegisterUser)
		authGroup.POST("/login", authHandler.LoginUser)
		authGroup.POST("/refresh", authHandler.RefreshTokens)
	}

	// User routes (protected)
	userGroup := router.Group("/users")
	userGroup.Use(middleware.JWTAuthMiddleware(log))
	{
		userGroup.GET("/", userHandler.GetAllUsers)
		userGroup.GET("/:id", userHandler.GetUserByID)
		userGroup.PUT("/:id", userHandler.UpdateUser)
		userGroup.DELETE("/:id", userHandler.DeleteUser)
	}

	// Get port from environment or use default
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	log.Info().Str("port", port).Msg("Starting server")

	if err := router.Run(":" + port); err != nil {
		log.Fatal().Err(err).Msg("Server failed to start")
	}
}
