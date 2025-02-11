package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"github.com/yourusername/user-management-api/internal/config"
	"github.com/yourusername/user-management-api/internal/database/sqlite"
	"github.com/yourusername/user-management-api/internal/handlers"
	"github.com/yourusername/user-management-api/internal/middleware"
	"github.com/yourusername/user-management-api/internal/repository"
	"github.com/yourusername/user-management-api/pkg/logger"
)

func main() {
	// Set Gin to production mode
	gin.SetMode(gin.ReleaseMode)

	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}
	sqliteConfig := sqlite.SQLiteConfig{
		Path:     "./users.db",
		InMemory: true,
	}
	// Initialize configuration
	_, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize database connection
	db, err := sqlite.InitSQLite(sqliteConfig)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Run migrations
	if err := sqlite.RunSQLiteMigrations(); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// inject to handler
	userRepository := repository.NewUserRepository(db)
	userHandler := handlers.NewUserHandler(userRepository)
	authHandler := handlers.NewAuthHandler(userRepository)

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
	userGroup.Use(middleware.JWTAuthMiddleware())
	{
		userGroup.GET("/", userHandler.GetAllUsers)
		userGroup.GET("/:id", userHandler.GetUserByID)
		userGroup.PUT("/:id", userHandler.UpdateUser)
		userGroup.DELETE("/:id", userHandler.DeleteUser)
	}

	// Start server
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	logger.Info("Starting server on port", port)
	log.Printf("Starting server on port %s", port)

	if err := router.Run(":" + port); err != nil {
		logger.Error("Server failed to start:", err)
		log.Fatalf("Server failed to start: %v", err)
	}
}
