package http

import (
	"homeapp/internal/delivery/http/handler"
	"homeapp/internal/delivery/http/middleware"
	"homeapp/internal/pkg/jwt"
	"homeapp/internal/pkg/storage"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/static"
)

// Handlers mengumpulkan semua HTTP handler yang di-wiring dari main.go.
type Handlers struct {
	Auth        *handler.AuthHandler
	Household   *handler.HouseholdHandler
	User        *handler.UserHandler
	Account     *handler.AccountHandler
	Category    *handler.CategoryHandler
	Transaction *handler.TransactionHandler
	Upload      *handler.UploadHandler
	Job         *handler.JobHandler
	Budget      *handler.BudgetHandler
	Goal        *handler.GoalHandler
	Bill        *handler.BillHandler
	Dashboard   *handler.DashboardHandler
	Report      *handler.ReportHandler
	BudgetPlan  *handler.BudgetPlanHandler
}

func RegisterRoutes(app *fiber.App, h *Handlers, jwtManager *jwt.Manager, isProduction bool, localStore *storage.LocalStore) {
	// Saat penyimpanan lokal aktif, sajikan file lampiran lewat route statis /uploads/*
	// (public — file_url dari upload harus bisa dibuka browser/lampiran tanpa autentikasi).
	if localStore != nil {
		if dir, err := localStore.DirAbs(); err == nil {
			// Endpoint PUT untuk menyimpan file ke penyimpanan LOKAL (driver=local).
			// Public (bukan JWT) supaya klien bisa upload langsung seperti presigned URL S3 —
			// keamanan terletak pada key acak (UUID) yang tidak bisa ditebak.
			// Didaftarkan di root app (bukan /api/v1) supaya lepas dari group protected,
			// dan SEBELUM static supaya method PUT tidak tertangkap static (yang hanya GET/HEAD).
			// Key memuat slash (mis. attachments/<uuid>-file.txt) → pakai wildcard `*`.
			app.Put("/uploads/*", h.Upload.UploadLocal)
			// Sajikan file lampiran lewat route statis /uploads/* (public).
			app.Use("/uploads", static.New(dir))
		}
	}

	api := app.Group("/api/v1")

	// --- Public routes ---
	auth := api.Group("/auth")
	auth.Post("/register", h.Auth.Register)
	auth.Post("/verify", h.Auth.VerifyEmail)
	auth.Post("/resend-verification", h.Auth.ResendVerification)
	auth.Post("/login", h.Auth.Login)

	// --- Protected routes (butuh JWT) ---
	protected := api.Group("", middleware.JWTProtected(jwtManager))

	users := protected.Group("/users")
	users.Get("/me", h.User.GetProfile)
	users.Patch("/me", h.User.UpdateProfile)
	users.Post("/me/change-password", h.User.ChangePassword)

	households := protected.Group("/households")
	households.Post("/", h.Household.Create)
	households.Post("/join", h.Household.Join)
	households.Get("/me", h.Household.GetDetail)
	households.Patch("/:id", h.Household.UpdateName)
	households.Get("/members", h.Household.GetMembers)
	households.Delete("/members", h.Household.RemoveMember)
	households.Post("/invitations", h.Household.CreateInvitation)
	households.Get("/invitations/active", h.Household.GetActiveInvitation)

	accounts := protected.Group("/accounts")
	accounts.Post("/", h.Account.Create)
	accounts.Get("/", h.Account.List)
	accounts.Get("/:id", h.Account.GetDetail)
	accounts.Patch("/:id", h.Account.Update)

	categories := protected.Group("/categories")
	categories.Post("/", h.Category.CreateCategory)
	categories.Get("/", h.Category.ListCategories)
	categories.Patch("/:id", h.Category.UpdateCategory)
	categories.Delete("/:id", h.Category.ArchiveCategory)
	categories.Post("/:id/unarchive", h.Category.UnarchiveCategory)

	transactions := protected.Group("/transactions")
	transactions.Get("/quick-select", h.Transaction.GetQuickSelect)
	transactions.Post("/", h.Transaction.CreateTransaction)
	transactions.Get("/", h.Transaction.ListTransactions)
	transactions.Get("/:id", h.Transaction.GetTransaction)
	transactions.Patch("/:id", h.Transaction.UpdateTransaction)
	transactions.Delete("/:id", h.Transaction.DeleteTransaction)

	uploads := protected.Group("/uploads")
	uploads.Post("/presign", h.Upload.PresignUpload)

	budgets := protected.Group("/budgets")
	budgets.Post("/", h.Budget.CreateBudget)
	budgets.Get("/", h.Budget.ListBudgets)
	budgets.Post("/copy-from-previous", h.Budget.CopyFromPrevious)
	budgets.Patch("/:id", h.Budget.UpdateBudget)
	budgets.Delete("/:id", h.Budget.DeleteBudget)

	protected.Get("/budget-plan", h.BudgetPlan.GetBudgetPlan)

	goals := protected.Group("/goals")
	goals.Post("/", h.Goal.CreateGoal)
	goals.Get("/", h.Goal.ListGoals)
	goals.Get("/:id", h.Goal.GetGoalDetail)
	goals.Patch("/:id", h.Goal.UpdateGoal)
	goals.Delete("/:id", h.Goal.DeleteGoal)

	bills := protected.Group("/bills")
	bills.Post("/", h.Bill.CreateBill)
	bills.Get("/", h.Bill.ListBills)
	bills.Get("/:id/periods", h.Bill.GetBillPeriods)
	bills.Patch("/:id", h.Bill.UpdateBill)
	bills.Delete("/:id", h.Bill.DeleteBill)
	bills.Post("/:id/stop", h.Bill.StopBill)

	billPeriods := protected.Group("/bill-periods")
	billPeriods.Post("/:id/pay", h.Bill.PayBillPeriod)

	protected.Get("/dashboard", h.Dashboard.GetDashboard)

	reports := protected.Group("/reports")
	reports.Get("/trend", h.Report.GetTrend)
	reports.Get("/category-breakdown", h.Report.GetCategoryBreakdown)
	reports.Get("/member-breakdown", h.Report.GetMemberBreakdown)
	reports.Get("/comparison", h.Report.GetComparison)
	reports.Get("/export", h.Report.ExportReport)

	// Trigger job manual untuk QA — sengaja tidak di-mount di production supaya
	// developer tidak bisa tidak sengaja kirim notifikasi dobel di prod.
	if !isProduction {
		internalJobs := protected.Group("/internal/jobs")
		internalJobs.Post("/:jobName/run", h.Job.RunJob)
	}

	// TODO: tambah modul lain di sini dengan pola yang sama:
	//   1. internal/entity           -> sudah ada semua (account, category, transaction, budget, goal, bill, bill_period)
	//   2. internal/repository       -> buat interface + implementasi postgres
	//   3. internal/usecase          -> buat business logic
	//   4. internal/delivery/http    -> buat dto + handler
	//   5. daftarkan route di sini
	//
	// Contoh urutan pengerjaan yang disarankan:
	//   accounts -> categories -> transactions -> budgets -> goals -> bills
}
