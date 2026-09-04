package apperror

import "errors"

// Kumpulan error domain yang dikenali & dipetakan ke HTTP status code di layer handler.
var (
	ErrNotFound              = errors.New("data tidak ditemukan")
	ErrEmailAlreadyUsed      = errors.New("email sudah terdaftar")
	ErrInvalidCredential     = errors.New("email atau password salah")
	ErrEmailNotVerified      = errors.New("email belum diverifikasi, silakan cek email kamu")
	ErrEmailAlreadyVerified  = errors.New("email sudah diverifikasi")
	ErrInvalidVerificationToken = errors.New("token verifikasi tidak valid")
	ErrVerificationTokenExpired = errors.New("token verifikasi sudah kadaluarsa")
	ErrInvitationInvalid     = errors.New("kode undangan tidak valid, sudah dipakai, atau sudah kadaluarsa")
	ErrForbidden             = errors.New("anda tidak memiliki akses untuk aksi ini")
	ErrAlreadyInHousehold    = errors.New("anda sudah tergabung dalam rumah tangga")
	ErrCannotRemoveSoleOwner = errors.New("tidak bisa mengeluarkan owner tunggal dari rumah tangga")

	ErrCategoryTypeMismatch   = errors.New("tipe kategori tidak cocok dengan tipe transaksi")
	ErrTransferSameAccount    = errors.New("akun tujuan transfer tidak boleh sama dengan akun sumber")
	ErrDestinationRequired    = errors.New("akun tujuan wajib diisi untuk transfer")
	ErrCategoryRequired       = errors.New("kategori wajib diisi untuk transaksi income/expense")
	ErrAccountInactive        = errors.New("akun tidak aktif")
	ErrBillPeriodPaidConflict = errors.New("transaksi ini terhubung tagihan yang sudah dibayar, batalkan status bayar tagihan terlebih dahulu")

	ErrCategoryNotExpense  = errors.New("kategori harus bertipe expense")
	ErrBudgetAlreadyExists = errors.New("budget untuk kategori dan periode ini sudah ada, gunakan PATCH untuk mengubah")
	ErrInvalidPeriodFormat = errors.New("format period tidak valid, gunakan YYYY-MM")

	ErrGoalHasContributions = errors.New("goal masih punya transaksi kontribusi, tidak bisa dihapus")

	ErrBillPeriodAlreadyPaid = errors.New("periode tagihan ini sudah dibayar")

	ErrInvalidPeriodType   = errors.New("period_type harus 'month' atau 'year'")
	ErrInvalidExportFormat = errors.New("format export harus 'pdf' atau 'excel'")

	ErrCategoryNestingTooDeep    = errors.New("sub-kategori tidak boleh punya sub-kategori sendiri (maksimal 2 tingkat)")
	ErrCategoryParentTypeMismatch = errors.New("sub-kategori harus punya tipe yang sama dengan kategori induk")
	ErrCategoryHasChildren       = errors.New("kategori ini punya sub-kategori, budget/tagihan harus diarahkan ke sub-kategori")

	ErrPersonalAccountForbidden = errors.New("akun ini milik pribadi anggota lain, anda tidak punya akses transaksi ke akun ini")
)
