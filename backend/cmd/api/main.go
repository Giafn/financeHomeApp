// @title Family Finance API
// @version 1.0
// @description API untuk manajemen keuangan keluarga
// @basePath /api/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Bearer token
package main

import (
	"log"

	"family-finance-api/internal/config"
	"family-finance-api/internal/database"
	httpDelivery "family-finance-api/internal/delivery/http"
	"family-finance-api/internal/delivery/http/handler"
	"family-finance-api/internal/pkg/jwt"
	postgresRepo "family-finance-api/internal/repository/postgres"
	"family-finance-api/internal/usecase"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	_ "family-finance-api/docs"
	fiberswagger "github.com/swaggo/fiber-swagger"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("gagal load konfigurasi: %v", err)
	}

	// PENTING: aplikasi TIDAK menjalankan migrasi otomatis di sini.
	// Pastikan migrasi sudah dijalankan manual sebelum start:
	//   make migrate-up   atau   go run ./cmd/migrate up
	db, err := database.NewPostgresConnection(cfg)
	if err != nil {
		log.Fatalf("gagal konek ke database: %v", err)
	}

	jwtManager := jwt.NewManager(cfg.JWTSecret, cfg.JWTExpiryHours)

	// --- Dependency wiring: repository -> usecase -> handler ---
	userRepo := postgresRepo.NewUserRepository(db)
	householdRepo := postgresRepo.NewHouseholdRepository(db)
	accountRepo := postgresRepo.NewAccountRepository(db)

	authUsecase := usecase.NewAuthUsecase(userRepo, jwtManager)
	userUsecase := usecase.NewUserUsecase(userRepo, householdRepo)
	householdUsecase := usecase.NewHouseholdUsecase(householdRepo)
	accountUsecase := usecase.NewAccountUsecase(householdRepo, accountRepo)

	handlers := &httpDelivery.Handlers{
		Auth:      handler.NewAuthHandler(authUsecase),
		User:      handler.NewUserHandler(userUsecase),
		Household: handler.NewHouseholdHandler(householdUsecase),
		Account:   handler.NewAccountHandler(accountUsecase),
	}

	app := fiber.New(fiber.Config{
		AppName: "Family Finance API",
	})

	app.Use(recover.New())
	app.Use(logger.New())
	corsConfig := cors.Config{
		AllowMethods: "GET,POST,PUT,PATCH,DELETE,OPTIONS",
		AllowHeaders: "Content-Type,Authorization",
	}

	if cfg.CORSAllowOrigins == "*" {
		corsConfig.AllowOrigins = "*"
		corsConfig.AllowCredentials = false
	} else {
		corsConfig.AllowOrigins = cfg.CORSAllowOrigins
		corsConfig.AllowCredentials = true
	}

	app.Use(cors.New(corsConfig))

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	app.Get("/swagger/*", fiberswagger.WrapHandler)

	httpDelivery.RegisterRoutes(app, handlers, jwtManager)

	log.Printf("🚀 server berjalan di port %s (env: %s)", cfg.AppPort, cfg.AppEnv)
	if err := app.Listen(":" + cfg.AppPort); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
