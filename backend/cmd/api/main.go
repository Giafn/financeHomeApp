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
	"strings"

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

	mailerClient := mailer.NewSMTPMailer(cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPUser, cfg.SMTPPassword, cfg.SMTPFrom)
	authUsecase := usecase.NewAuthUsecase(userRepo, jwtManager, mailerClient, cfg.AppURL)
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
	billUsecase := usecase.NewBillUsecase(billRepo, billPeriodRepo, categoryRepo, accountRepo, householdRepo, transactionUsecase, txManager)
	budgetPlanUsecase := usecase.NewBudgetPlanUsecase(budgetRepo, billPeriodRepo, transactionRepo, householdRepo, accountUsecase)
	dashboardUsecase := usecase.NewDashboardUsecase(accountUsecase, budgetUsecase, goalUsecase, transactionUsecase, budgetPlanUsecase, householdRepo, transactionRepo, billPeriodRepo)
	reportUsecase := usecase.NewReportUsecase(transactionRepo, householdRepo)

	// Penyimpanan lampiran: S3 (presigned, client upload langsung ke bucket) atau
	// lokal (file disimpan di disk server, disajikan lewat route statis /uploads).
	// Default: S3 kalau S3_BUCKET diisi, selain itu local. Bisa dipaksa lewat STORAGE_DRIVER.
	storageDriver := cfg.StorageDriver
	if storageDriver == "" {
		if cfg.S3Bucket != "" {
			storageDriver = "s3"
		} else {
			storageDriver = "local"
		}
	}

	var fileStore storage.Storage
	var localStore *storage.LocalStore
	switch storageDriver {
	case "s3":
		fileStore, err = storage.NewPresigner(context.Background(), cfg.S3Endpoint, cfg.S3Region, cfg.S3AccessKey, cfg.S3SecretKey, cfg.S3Bucket)
		if err != nil {
			log.Fatalf("gagal inisialisasi S3 presigner: %v", err)
		}
	case "local":
		localStore, err = storage.NewLocalStore(cfg.StorageLocalDir, cfg.StoragePublicBaseURL)
		if err != nil {
			log.Fatalf("gagal inisialisasi local storage: %v", err)
		}
		fileStore = localStore
	default:
		log.Fatalf("STORAGE_DRIVER tidak dikenal: %s (harus 'local' atau 's3')", storageDriver)
	}

	notificationLogRepo := postgresRepo.NewNotificationLogRepository(db)
	notificationGuard := notification.NewGuard(notificationLogRepo)

	jobRegistry := job.NewRegistry()
	job.RegisterJobs(jobRegistry, mailerClient, notificationGuard, cfg.SMTPUser, budgetRepo, householdRepo, billRepo, billPeriodRepo)

	handlers := &httpDelivery.Handlers{
		Auth:        handler.NewAuthHandler(authUsecase),
		User:        handler.NewUserHandler(userUsecase, fileStore),
		Household:   handler.NewHouseholdHandler(householdUsecase),
		Account:     handler.NewAccountHandler(accountUsecase),
		Category:    handler.NewCategoryHandler(categoryUsecase, householdRepo, validate),
		Transaction: handler.NewTransactionHandler(transactionUsecase, validate, fileStore),
		Upload:      handler.NewUploadHandler(fileStore, localStore, validate),
		Job:         handler.NewJobHandler(jobRegistry),
		Budget:      handler.NewBudgetHandler(budgetUsecase, validate),
		Goal:        handler.NewGoalHandler(goalUsecase, validate),
		Bill:        handler.NewBillHandler(billUsecase, validate),
		Dashboard:   handler.NewDashboardHandler(dashboardUsecase),
		Report:      handler.NewReportHandler(reportUsecase),
		BudgetPlan:  handler.NewBudgetPlanHandler(budgetPlanUsecase),
	}

	app := fiber.New(fiber.Config{
		AppName:   "Family Finance API",
		BodyLimit: cfg.MaxUploadMB << 20, // batas upload (MB -> bytes)
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
		corsConfig.AllowOrigins = strings.Split(cfg.CORSAllowOrigins, ",")
		corsConfig.AllowCredentials = true
	}

	app.Use(cors.New(corsConfig))

	app.Get("/health", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	// TODO: fiber-swagger (swaggo) hanya support fiber v2; belum ada handler v3-compatible.
	// docs.go/swagger.json tetap ter-generate oleh `swag init`, tapi UI-nya belum di-mount.

	httpDelivery.RegisterRoutes(app, handlers, jwtManager, cfg.AppEnv == "production", localStore)

	log.Printf("🚀 server berjalan di port %s (env: %s)", cfg.AppPort, cfg.AppEnv)
	if err := app.Listen(":" + cfg.AppPort); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
