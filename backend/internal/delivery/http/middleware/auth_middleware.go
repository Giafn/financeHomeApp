package middleware

import (
	"strings"

	"family-finance-api/internal/pkg/jwt"
	"family-finance-api/internal/pkg/response"

	"github.com/gofiber/fiber/v2"
)

const LocalsUserID = "user_id"
const LocalsEmail = "email"

// JWTProtected memvalidasi header "Authorization: Bearer <token>" dan menyimpan
// user_id & email hasil parse ke c.Locals() supaya bisa dipakai handler berikutnya.
func JWTProtected(jwtManager *jwt.Manager) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return response.Error(c, fiber.StatusUnauthorized, "token otorisasi tidak ditemukan")
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			return response.Error(c, fiber.StatusUnauthorized, "format token tidak valid, gunakan: Bearer <token>")
		}

		claims, err := jwtManager.VerifyToken(parts[1])
		if err != nil {
			return response.Error(c, fiber.StatusUnauthorized, "token tidak valid atau sudah kadaluarsa")
		}

		c.Locals(LocalsUserID, claims.UserID)
		c.Locals(LocalsEmail, claims.Email)

		return c.Next()
	}
}
