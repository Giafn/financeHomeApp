package handler

import (
	"errors"

	"github.com/gofiber/fiber/v3"
	"homeapp/internal/job"
	"homeapp/internal/pkg/response"
)

// JobHandler adalah pintu masuk manual untuk QA memicu job tanpa menunggu jadwal cron.
// Hanya di-mount di router kalau APP_ENV != production (lihat router.go).
type JobHandler struct {
	registry *job.Registry
}

func NewJobHandler(registry *job.Registry) *JobHandler {
	return &JobHandler{registry: registry}
}

// RunJob godoc
//
//	@Summary		Manually trigger a background job (non-production only)
//	@Tags			Internal
//	@Produce		json
//	@Param			jobName	path	string	true	"Job name"
//	@Success		200	{object}	response.Envelope
//	@Failure		404	{object}	response.Envelope
//	@Router			/internal/jobs/{jobName}/run [post]
//	@Security		BearerAuth
func (h *JobHandler) RunJob(c fiber.Ctx) error {
	jobName := c.Params("jobName")

	result, err := h.registry.Run(c.Context(), jobName)
	if err != nil {
		if errors.Is(err, job.ErrJobNotFound) {
			return response.Error(c, fiber.StatusNotFound, "Job tidak ditemukan: "+jobName)
		}
		return response.Error(c, fiber.StatusInternalServerError, "Job gagal: "+err.Error())
	}

	return response.Success(c, fiber.StatusOK, "job dijalankan", result)
}
