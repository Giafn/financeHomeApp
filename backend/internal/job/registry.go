package job

import (
	"context"
	"errors"
)

// Func adalah bentuk generik sebuah background job. Result dikembalikan sebagai ringkasan
// (misal jumlah email terkirim) supaya endpoint manual trigger bisa menampilkannya ke QA.
type Func func(ctx context.Context) (map[string]interface{}, error)

var ErrJobNotFound = errors.New("job tidak ditemukan")

// Registry memetakan nama job ke implementasinya. Dipakai baik oleh cmd/worker
// (dijalankan lewat Scheduler sesuai jadwal cron) maupun cmd/api (dipicu manual
// lewat endpoint /internal/jobs/{jobName}/run untuk QA) — registrasi job yang sama
// dipakai di kedua tempat lewat RegisterJobs, supaya tidak dobel definisi.
type Registry struct {
	jobs map[string]Func
}

func NewRegistry() *Registry {
	return &Registry{jobs: make(map[string]Func)}
}

func (r *Registry) Add(name string, fn Func) {
	r.jobs[name] = fn
}

func (r *Registry) Get(name string) (Func, bool) {
	fn, ok := r.jobs[name]
	return fn, ok
}

func (r *Registry) Run(ctx context.Context, name string) (map[string]interface{}, error) {
	fn, ok := r.jobs[name]
	if !ok {
		return nil, ErrJobNotFound
	}
	return fn(ctx)
}
