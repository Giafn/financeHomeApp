package handler

import (
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"homeapp/internal/delivery/http/dto"
	"homeapp/internal/pkg/response"
	"homeapp/internal/pkg/storage"
)

type UploadHandler struct {
	presigner *storage.Presigner
	validator *validator.Validate
}

func NewUploadHandler(presigner *storage.Presigner, validator *validator.Validate) *UploadHandler {
	return &UploadHandler{presigner: presigner, validator: validator}
}

// PresignUpload godoc
//
//	@Summary		Get a presigned S3 upload URL
//	@Description	Client uploads directly to S3 with the returned upload_url (PUT), then submits file_url as attachment_url
//	@Tags			Uploads
//	@Accept			json
//	@Produce		json
//	@Param			body	body		dto.PresignUploadRequest	true	"File info"
//	@Success		200		{object}	response.Envelope{data=dto.PresignUploadResponse}
//	@Failure		400		{object}	response.Envelope
//	@Router			/uploads/presign [post]
//	@Security		BearerAuth
func (h *UploadHandler) PresignUpload(c fiber.Ctx) error {
	if h.presigner == nil {
		return response.Error(c, fiber.StatusServiceUnavailable, "Upload lampiran belum dikonfigurasi di server ini")
	}

	var req dto.PresignUploadRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Format request tidak valid")
	}
	if err := h.validator.Struct(req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Validasi gagal: "+err.Error())
	}

	uploadURL, fileURL, err := h.presigner.PresignUpload(c.Context(), req.Filename, req.ContentType)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Gagal membuat presigned URL")
	}

	return response.Success(c, fiber.StatusOK, "ok", dto.PresignUploadResponse{
		UploadURL: uploadURL,
		FileURL:   fileURL,
	})
}
