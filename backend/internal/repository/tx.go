package repository

import "context"

// TxManager menjalankan fn dalam satu database transaction. ctx yang diteruskan
// ke fn membawa handle transaksi — repository lain yang menerima ctx yang sama
// (household, category, dst) otomatis ikut transaksi ini, bukan koneksi terpisah.
type TxManager interface {
	WithinTx(ctx context.Context, fn func(ctx context.Context) error) error
}
