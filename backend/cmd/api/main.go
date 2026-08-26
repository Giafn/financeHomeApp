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
	"context"
	"log"

	"homeapp/internal/config"
	"homeapp/internal/database"
	httpDelivery "homeapp/internal/delivery/http"
	"homeapp/internal/delivery/http/handler"
	"homeapp/internal/job"
	"homeapp/internal/pkg/jwt"
	"homeapp/internal/pkg/mailer"
	"homeapp/internal/pkg/notification"
	"homeapp/internal/pkg/storage"
	postgresRepo "homeapp/internal/repository/postgres"
	"homeapp/internal/usecase"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/gofiber/fiber/v3/middleware/recover"
	_ "homeapp/docs"
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
	validate := validator.New()

	// --- Dependency wiring: repository -> usecase -> handler ---
	userRepo := postgresRepo.NewUserRepository(db)
	householdRepo := postgresRepo.NewHouseholdRepository(db)
	accountRepo := postgresRepo.NewAccountRepository(db)
	categoryRepo := postgresRepo.NewCategoryRepository(db)
	transactionRepo := postgresRepo.NewTransactionRepository(db)
	txManager := postgresRepo.NewTxManager(db)

	authUsecase := usecase.NewAuthUsecase(userRepo, jwtManager)
	userUsecase := usecase.NewUserUsecase(userRepo, householdRepo)
	householdUsecase := usecase.NewHouseholdUsecase(householdRepo, categoryRepo, txManager)
	accountUsecase := usecase.NewAccountUsecase(householdRepo, accountRepo)
	categoryUsecase := usecase.NewCategoryUsecase(categoryRepo)
	budgetRepo := postgresRepo.NewBudgetRepository(db)
	budgetUsecase := usecase.NewBudgetUsecase(budgetRepo, categoryRepo, householdRepo)
	goalRepo := postgresRepo.NewGoalRepository(db)
	goalUsecase := usecase.NewGoalUsecase(goalRepo, accountRepo, householdRepo, transactionRepo)
	transactionUsecase := usecase.NewTransactionUsecase(transactionRepo, accountRepo, categoryRepo, householdRepo, goalUsecase)
	billRepo := postgresRepo.NewBillRepository(db)
	billPeriodRepo := postgresRepo.NewBillPeriodRepository(db)
	billUsecase := usecase.NewBillUsecase(billRepo, billPeriodRepo, categoryRepo, accountRepo, householdRepo, transactionUsecase)
	dashboardUsecase := usecase.NewDashboardUsecase(accountUsecase, budgetUsecase, goalUsecase, transactionUsecase, householdRepo, transactionRepo, billPeriodRepo)
	reportUsecase := usecase.NewReportUsecase(transactionRepo, householdRepo)

	// Presigner opsional — kalau S3 belum dikonfigurasi, endpoint presign akan 503
	// alih-alih membuat server gagal start di lingkungan dev tanpa S3.
	var presigner *storage.Presigner
	if cfg.S3Bucket != "" {
		presigner, err = storage.NewPresigner(context.Background(), cfg.S3Endpoint, cfg.S3Region, cfg.S3AccessKey, cfg.S3SecretKey, cfg.S3Bucket)
		if err != nil {
			log.Fatalf("gagal inisialisasi S3 presigner: %v", err)
		}
	}

	notificationLogRepo := postgresRepo.NewNotificationLogRepository(db)
	notificationGuard := notification.NewGuard(notificationLogRepo)
	mailerClient := mailer.NewSMTPMailer(cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPUser, cfg.SMTPPassword, cfg.SMTPFrom)

	jobRegistry := job.NewRegistry()
	job.RegisterJobs(jobRegistry, mailerClient, notificationGuard, cfg.SMTPUser, budgetRepo, householdRepo, billRepo, billPeriodRepo)

	handlers := &httpDelivery.Handlers{
		Auth:        handler.NewAuthHandler(authUsecase),
		User:        handler.NewUserHandler(userUsecase),
		Household:   handler.NewHouseholdHandler(householdUsecase),
		Account:     handler.NewAccountHandler(accountUsecase),
		Category:    handler.NewCategoryHandler(categoryUsecase, householdRepo, validate),
		Transaction: handler.NewTransactionHandler(transactionUsecase, validate),
		Upload:      handler.NewUploadHandler(presigner, validate),
		Job:         handler.NewJobHandler(jobRegistry),
		Budget:      handler.NewBudgetHandler(budgetUsecase, validate),
		Goal:        handler.NewGoalHandler(goalUsecase, validate),
		Bill:        handler.NewBillHandler(billUsecase, validate),
		Dashboard:   handler.NewDashboardHandler(dashboardUsecase),
		Report:      handler.NewReportHandler(reportUsecase),
	}

	app := fiber.New(fiber.Config{
		AppName: "Family Finance API",
	})

	app.Use(recover.New())
	app.Use(logger.New())
	corsConfig := cors.Config{
		AllowMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders: []string{"Content-Type", "Authorization"},
	}

	if cfg.CORSAllowOrigins == "*" {
		corsConfig.AllowOrigins = []string{"*"}
		corsConfig.AllowCredentials = false
	} else {
		corsConfig.AllowOrigins = []string{cfg.CORSAllowOrigins}
		corsConfig.AllowCredentials = true
	}

	app.Use(cors.New(corsConfig))

	app.Get("/health", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	// TODO: fiber-swagger (swaggo) hanya support fiber v2; belum ada handler v3-compatible.
	// docs.go/swagger.json tetap ter-generate oleh `swag init`, tapi UI-nya belum di-mount.

	httpDelivery.RegisterRoutes(app, handlers, jwtManager, cfg.AppEnv == "production")

	log.Printf("🚀 server berjalan di port %s (env: %s)", cfg.AppPort, cfg.AppEnv)
	if err := app.Listen(":" + cfg.AppPort); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
