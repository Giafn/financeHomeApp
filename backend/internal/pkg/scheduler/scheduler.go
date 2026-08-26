package scheduler

import (
	"log"

	"github.com/robfig/cron/v3"
)

// Scheduler membungkus robfig/cron supaya job-job (Phase 08/10) daftar pakai
// nama+jadwal cron string tanpa perlu tahu library cron yang dipakai di baliknya.
type Scheduler struct {
	cron *cron.Cron
}

func New() *Scheduler {
	return &Scheduler{cron: cron.New()}
}

// Register mendaftarkan job dengan jadwal cron standar 5-field ("0 8 * * *" = tiap jam 8 pagi).
// Panic dari job di-recover supaya satu job yang error tidak mematikan seluruh scheduler.
func (s *Scheduler) Register(name, cronExpr string, job func()) error {
	_, err := s.cron.AddFunc(cronExpr, func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("job %q panic: %v", name, r)
			}
		}()
		log.Printf("job %q mulai jalan", name)
		job()
		log.Printf("job %q selesai", name)
	})
	return err
}

func (s *Scheduler) Start() {
	s.cron.Start()
}

func (s *Scheduler) Stop() {
	<-s.cron.Stop().Done()
}
