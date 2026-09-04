package handler

import (
	"log"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"homeapp/internal/delivery/http/dto"
	"homeapp/internal/pkg/response"
	"homeapp/internal/pkg/storage"
)

// MaxUploadSizeBytes batas maksimum berkas yang boleh diunggah (5 MB).
const MaxUploadSizeBytes = 5 * 1024 * 1024

type UploadHandler struct {
	store      storage.Storage
	localStore *storage.LocalStore // non-nil hanya saat driver = local
	validator  *validator.Validate
}

func NewUploadHandler(store storage.Storage, localStore *storage.LocalStore, validator *validator.Validate) *UploadHandler {
	return &UploadHandler{store: store, localStore: localStore, validator: validator}
}

// PresignUpload godoc
//
//	@Summary		Get upload URL for a file
//	@Description	Returns an upload_url (client PUTs the file body to it) and the final file_url to store as attachment_url. Works for both S3 presigned URL and local storage.
//	@Tags			Uploads
//	@Accept			json
//	@Produce		json
//	@Param			body	body		dto.PresignUploadRequest	true	"File info"
//	@Success		200		{object}	response.Envelope{data=dto.PresignUploadResponse}
//	@Failure		400		{object}	response.Envelope
//	@Failure		503		{object}	response.Envelope
//	@Router			/uploads/presign [post]
//	@Security		BearerAuth
func (h *UploadHandler) PresignUpload(c fiber.Ctx) error {
	if h.store == nil {
		return response.Error(c, fiber.StatusServiceUnavailable, "Upload lampiran belum dikonfigurasi di server ini")
	}

	var req dto.PresignUploadRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Format request tidak valid")
	}
	if err := h.validator.Struct(req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Validasi gagal: "+err.Error())
	}

	// Hanya gambar yang boleh diunggah (avatar & lampiran ditampilkan via <img>).
	if !strings.HasPrefix(strings.ToLower(req.ContentType), "image/") {
		return response.Error(c, fiber.StatusBadRequest, "Hanya file gambar yang diperbolehkan")
	}

	uploadURL, fileURL, err := h.store.UploadURL(c.Context(), req.Filename, req.ContentType)
	if err != nil {
		log.Printf("gagal presign upload: filename=%s err=%v", req.Filename, err)
		return response.Error(c, fiber.StatusInternalServerError, "Gagal membuat URL upload")
	}

	return response.Success(c, fiber.StatusOK, "ok", dto.PresignUploadResponse{
		UploadURL: uploadURL,
		FileURL:   fileURL,
	})
}

// UploadLocal menangani PUT raw body ke penyimpanan lokal (HANYA saat driver = local).
// upload_url dari LocalStore.UploadURL menunjuk ke route ini: PUT /api/v1/uploads/:key
func (h *UploadHandler) UploadLocal(c fiber.Ctx) error {
	if h.localStore == nil {
		return response.Error(c, fiber.StatusServiceUnavailable, "Penyimpanan lokal tidak aktif")
	}

	key := c.Params("*")
	if key == "" {
		return response.Error(c, fiber.StatusBadRequest, "Key file tidak valid")
	}

	body := c.Body()
	if len(body) == 0 {
		return response.Error(c, fiber.StatusBadRequest, "Body file kosong atau tidak terbaca")
	}

	if _, err := h.localStore.SaveKey(key, body); err != nil {
		log.Printf("gagal menyimpan file upload: key=%s err=%v", key, err)
		return response.Error(c, fiber.StatusInternalServerError, "Gagal menyimpan file")
	}

	// fileURL sama dengan yang dihasilkan UploadURL, dibangun ulang dari key.
	fileURL := h.localStore.PublicBaseURL + "/uploads/" + key

	return response.Success(c, fiber.StatusOK, "ok", dto.PresignUploadResponse{
		UploadURL: "",
		FileURL:   fileURL,
	})
}
