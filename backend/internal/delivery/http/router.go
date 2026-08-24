package http

import (
	"family-finance-api/internal/delivery/http/handler"
	"family-finance-api/internal/delivery/http/middleware"
	"family-finance-api/internal/pkg/jwt"

	"github.com/gofiber/fiber/v2"
)

// Handlers mengumpulkan semua HTTP handler yang di-wiring dari main.go.
type Handlers struct {
	Auth      *handler.AuthHandler
	Household *handler.HouseholdHandler
	User      *handler.UserHandler
	Account   *handler.AccountHandler
}

func RegisterRoutes(app *fiber.App, h *Handlers, jwtManager *jwt.Manager) {
	api := app.Group("/api/v1")

	// --- Public routes ---
	auth := api.Group("/auth")
	auth.Post("/register", h.Auth.Register)
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
