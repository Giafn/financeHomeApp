package apperror

import "errors"

// Kumpulan error domain yang dikenali & dipetakan ke HTTP status code di layer handler.
var (
	ErrNotFound              = errors.New("data tidak ditemukan")
	ErrEmailAlreadyUsed      = errors.New("email sudah terdaftar")
	ErrInvalidCredential     = errors.New("email atau password salah")
	ErrInvitationInvalid     = errors.New("kode undangan tidak valid, sudah dipakai, atau sudah kadaluarsa")
	ErrForbidden             = errors.New("anda tidak memiliki akses untuk aksi ini")
	ErrAlreadyInHousehold    = errors.New("anda sudah tergabung dalam rumah tangga")
	ErrCannotRemoveSoleOwner = errors.New("tidak bisa mengeluarkan owner tunggal dari rumah tangga")
)
