package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

// LocalStore menyimpan file lampiran ke disk lokal dan menyajikannya lewat route
// statis backend. Dipakai saat S3 belum dikonfigurasi (STORAGE_DRIVER=local).
type LocalStore struct {
	// Dir adalah direktori tempat file disimpan di disk server.
	Dir string
	// PublicBaseURL adalah base URL publik untuk mengakses file yang disimpan,
	// misal http://localhost:8080 — tempat route statis /uploads dipasang.
	PublicBaseURL string
}

func NewLocalStore(dir, publicBaseURL string) (*LocalStore, error) {
	if err := os.MkdirAll(filepath.Join(dir, "attachments"), 0o755); err != nil {
		return nil, err
	}
	return &LocalStore{Dir: dir, PublicBaseURL: strings.TrimRight(publicBaseURL, "/")}, nil
}

// UploadURL mengembalikan endpoint PUT backend lokal + URL publik file.
// Client melakukan PUT raw body ke uploadURL (route PUT /uploads/:key di app root,
// di luar /api/v1 supaya lepas dari JWT — setara presigned URL S3), lalu backend
// menulis ke disk. uploadURL dan fileURL sama-sama menunjuk ke /uploads/:key.
func (s *LocalStore) UploadURL(_ context.Context, filename, _ string) (string, string, error) {
	key := fmt.Sprintf("attachments/%s-%s", uuid.NewString(), sanitizeFilename(filename))
	url := fmt.Sprintf("%s/uploads/%s", s.PublicBaseURL, key)
	return url, url, nil
}

// SaveKey menulis raw data ke disk untuk sebuah key yang dihasilkan UploadURL.
// key divalidasi ketat agar tidak bisa path-traversal / menulis di luar Dir.
func (s *LocalStore) SaveKey(key string, data []byte) (string, error) {
	if !validKey(key) {
		return "", fmt.Errorf("key tidak valid")
	}
	absDir, err := filepath.Abs(s.Dir)
	if err != nil {
		return "", err
	}
	dest := filepath.Join(absDir, filepath.FromSlash(key))

	// Pastikan dest tetap berada di dalam Dir (cegah traversal).
	if rel, err := filepath.Rel(absDir, dest); err != nil || strings.HasPrefix(rel, "..") {
		return "", fmt.Errorf("key di luar direktori penyimpanan")
	}

	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return "", err
	}
	return dest, os.WriteFile(dest, data, 0o644)
}

// OpenKey membaca file untuk key tertentu (dipakai route statis).
func (s *LocalStore) OpenKey(key string) (io.ReadCloser, error) {
	if !validKey(key) {
		return nil, fmt.Errorf("key tidak valid")
	}
	absDir, err := filepath.Abs(s.Dir)
	if err != nil {
		return nil, err
	}
	dest := filepath.Join(absDir, filepath.FromSlash(key))
	if rel, err := filepath.Rel(absDir, dest); err != nil || strings.HasPrefix(rel, "..") {
		return nil, fmt.Errorf("key di luar direktori penyimpanan")
	}
	return os.Open(dest)
}

// DirAbs mengembalikan absolute path direktori penyimpanan (untuk route statis).
func (s *LocalStore) DirAbs() (string, error) {
	return filepath.Abs(s.Dir)
}

// validKey hanya memperbolehkan path relatif sederhana: huruf, angka, '/' , '-', '_', '.'.
func validKey(key string) bool {
	if key == "" || strings.Contains(key, "..") || strings.ContainsAny(key, "\\") {
		return false
	}
	for _, r := range key {
		if !(r >= 'a' && r <= 'z') && !(r >= 'A' && r <= 'Z') && !(r >= '0' && r <= '9') &&
			r != '/' && r != '-' && r != '_' && r != '.' {
			return false
		}
	}
	return true
}

// sanitizeFilename menyingkirkan path separator & karakter berbahaya dari nama file
// asli user supaya tidak merusak struktur key.
func sanitizeFilename(name string) string {
	name = filepath.Base(name)
	name = strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9',
			r == '-', r == '_', r == '.', r == ' ':
			return r
		default:
			return '_'
		}
	}, name)
	return name
}
