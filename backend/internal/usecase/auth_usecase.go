package usecase

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"homeapp/internal/entity"
	"homeapp/internal/pkg/apperror"
	"homeapp/internal/pkg/hash"
	"homeapp/internal/pkg/jwt"
	"homeapp/internal/pkg/mailer"
	"homeapp/internal/repository"
)

type AuthUsecase struct {
	userRepo   repository.UserRepository
	jwtManager *jwt.Manager
	mailer     mailer.Mailer
	appURL     string
}

func NewAuthUsecase(userRepo repository.UserRepository, jwtManager *jwt.Manager, mailer mailer.Mailer, appURL string) *AuthUsecase {
	return &AuthUsecase{userRepo: userRepo, jwtManager: jwtManager, mailer: mailer, appURL: appURL}
}

type RegisterInput struct {
	Name     string
	Email    string
	Password string
}

type LoginInput struct {
	Email    string
	Password string
}

type AuthResult struct {
	Token string
	User  *entity.User
}

// verificationTTL menentukan berapa lama token verifikasi email berlaku.
const verificationTTL = 24 * time.Hour

func (u *AuthUsecase) Register(ctx context.Context, in RegisterInput) (*entity.User, error) {
	existing, err := u.userRepo.FindByEmail(ctx, in.Email)
	if err != nil && !errors.Is(err, apperror.ErrNotFound) {
		return nil, err
	}
	if existing != nil {
		return nil, apperror.ErrEmailAlreadyUsed
	}

	hashed, err := hash.HashPassword(in.Password)
	if err != nil {
		return nil, err
	}

	token := randomToken()

	now := time.Now()
	user := &entity.User{
		Name:                 in.Name,
		Email:                in.Email,
		PasswordHash:         &hashed,
		VerificationToken:    &token,
		VerificationTokenExp: &now,
	}
	*(user.VerificationTokenExp) = now.Add(verificationTTL)

	if err := u.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	if err := u.sendVerificationEmail(ctx, user.Name, user.Email, token); err != nil {
		// Email gagal terkirim -> gagalkan pendaftaran supaya user tahu akunnya
		// belum bisa dipakai sampai verifikasi berhasil.
		_ = u.userRepo.Delete(ctx, user.ID)
		return nil, err
	}

	return user, nil
}

func (u *AuthUsecase) Login(ctx context.Context, in LoginInput) (*AuthResult, error) {
	user, err := u.userRepo.FindByEmail(ctx, in.Email)
	if err != nil {
		if errors.Is(err, apperror.ErrNotFound) {
			return nil, apperror.ErrInvalidCredential
		}
		return nil, err
	}

	if user.PasswordHash == nil || !hash.CheckPassword(in.Password, *user.PasswordHash) {
		return nil, apperror.ErrInvalidCredential
	}

	if !user.Verified() {
		return nil, apperror.ErrEmailNotVerified
	}

	token, err := u.jwtManager.GenerateToken(user.ID, user.Email)
	if err != nil {
		return nil, err
	}

	return &AuthResult{Token: token, User: user}, nil
}

// VerifyEmail memvalidasi token verifikasi email dan menandai user sebagai verified.
func (u *AuthUsecase) VerifyEmail(ctx context.Context, token string) error {
	user, err := u.userRepo.FindByVerificationToken(ctx, token)
	if err != nil {
		if errors.Is(err, apperror.ErrNotFound) {
			return apperror.ErrInvalidVerificationToken
		}
		return err
	}

	if user.Verified() {
		return nil
	}

	if user.VerificationTokenExp == nil || time.Now().After(*user.VerificationTokenExp) {
		return apperror.ErrVerificationTokenExpired
	}

	now := time.Now()
	user.EmailVerifiedAt = &now
	user.VerificationToken = nil
	user.VerificationTokenExp = nil

	return u.userRepo.Update(ctx, user)
}

// ResendVerification membuat token baru dan mengirim ulang email verifikasi.
func (u *AuthUsecase) ResendVerification(ctx context.Context, in LoginInput) error {
	user, err := u.userRepo.FindByEmail(ctx, in.Email)
	if err != nil {
		return apperror.ErrInvalidCredential
	}
	if user.PasswordHash == nil || !hash.CheckPassword(in.Password, *user.PasswordHash) {
		return apperror.ErrInvalidCredential
	}
	if user.Verified() {
		return apperror.ErrEmailAlreadyVerified
	}

	token := randomToken()
	exp := time.Now().Add(verificationTTL)
	user.VerificationToken = &token
	user.VerificationTokenExp = &exp
	if err := u.userRepo.Update(ctx, user); err != nil {
		return err
	}

	return u.sendVerificationEmail(ctx, user.Name, user.Email, token)
}

func (u *AuthUsecase) sendVerificationEmail(ctx context.Context, name, email, token string) error {
	verifyURL := fmt.Sprintf("%s/verify?token=%s", u.appURL, token)
	subject := "Verifikasi email Family Finance"
	body := fmt.Sprintf(`<div style="font-family:Arial,Helvetica,sans-serif;max-width:560px;margin:auto;padding:24px;border:1px solid #e5e7eb;border-radius:12px">
  <h2 style="margin:0 0 16px;color:#111827">Halo %s,</h2>
  <p style="color:#374151;line-height:1.6">Terima kasih sudah mendaftar di <strong>Family Finance</strong>.
  Untuk mengaktifkan akunmu, silakan verifikasi alamat email dengan klik tombol di bawah ini:</p>
  <p style="text-align:center;margin:28px 0">
    <a href="%s" style="background:#4f46e5;color:#ffffff;text-decoration:none;padding:12px 28px;border-radius:8px;font-weight:600;display:inline-block">Verifikasi Email</a>
  </p>
  <p style="color:#6b7280;font-size:13px;line-height:1.6">Jika tombol tidak berfungsi, salin dan buka link berikut di browser:<br>
  <a href="%s" style="color:#4f46e5">%s</a></p>
  <p style="color:#9ca3af;font-size:12px;margin-top:24px">Link ini berlaku selama 24 jam. Jika kamu tidak mendaftar, abaikan email ini.</p>
</div>`, name, verifyURL, verifyURL, verifyURL)

	return u.mailer.Send([]string{email}, subject, body)
}

func randomToken() string {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
