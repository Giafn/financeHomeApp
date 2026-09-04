package storage

import "context"

// Storage adalah abstraksi penyimpanan file lampiran (attachment).
// Terdapat dua implementasi:
//   - S3/presigner : client upload langsung ke S3 lewat presigned URL.
//   - LocalStore    : backend menyimpan file ke disk lokal dan menyajikannya
//     lewat route statis (dipakai saat S3 belum dikonfigurasi).
//
// Contract berkas yang dihasilkan selalu: (uploadURL, fileURL).
// Client melakukan PUT body file ke uploadURL, lalu menyimpan fileURL sebagai
// attachment_url/avatar_url yang bisa diakses publik.
type Storage interface {
	// UploadURL mengembalikan (uploadURL, fileURL) untuk sebuah file.
	UploadURL(ctx context.Context, filename, contentType string) (uploadURL, fileURL string, err error)
	// ReadURL mengembalikan URL yang bisa diakses untuk membaca file yang sudah
	// disimpan dengan key tertentu. Untuk S3 ini presigned GET URL; untuk local
	// store ini URL publik yang sama dengan saat upload.
	ReadURL(ctx context.Context, key string) (string, error)
}
