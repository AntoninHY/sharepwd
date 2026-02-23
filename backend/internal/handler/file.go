package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jizo-hr/sharepwd/internal/config"
	"github.com/jizo-hr/sharepwd/internal/model"
	"github.com/jizo-hr/sharepwd/internal/repository"
	"github.com/jizo-hr/sharepwd/internal/service"
	"github.com/jizo-hr/sharepwd/internal/storage"
)

type FileHandler struct {
	secretService *service.SecretService
	fileRepo      *repository.FileRepository
	storage       storage.Storage
	config        *config.Config
}

func NewFileHandler(
	secretSvc *service.SecretService,
	fileRepo *repository.FileRepository,
	store storage.Storage,
	cfg *config.Config,
) *FileHandler {
	return &FileHandler{
		secretService: secretSvc,
		fileRepo:      fileRepo,
		storage:       store,
		config:        cfg,
	}
}

func (h *FileHandler) InitUpload(w http.ResponseWriter, r *http.Request) {
	var req model.InitFileUploadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.EncryptedName == "" || req.IV == "" || req.ChunkCount < 1 {
		writeError(w, http.StatusBadRequest, "encrypted_name, iv, and chunk_count are required")
		return
	}

	if req.OriginalSize > h.config.MaxFileSize {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("file too large, max %d bytes", h.config.MaxFileSize))
		return
	}

	ip := extractIP(r)
	ua := r.UserAgent()

	createReq := &model.CreateSecretRequest{
		EncryptedData: "file_placeholder",
		IV:            req.IV,
		Salt:          req.Salt,
		MaxViews:      req.MaxViews,
		ExpiresIn:     req.ExpiresIn,
		BurnAfterRead: req.BurnAfterRead,
		ContentType:   "file",
	}

	secretResp, err := h.secretService.Create(r.Context(), createReq, ip, ua)
	if err != nil {
		slog.Error("failed to create secret for file", "error", err)
		writeError(w, http.StatusInternalServerError, "failed to create secret")
		return
	}

	fileID := uuid.New()
	storageKey := fmt.Sprintf("files/%s/%s", secretResp.AccessToken, fileID.String())

	file := &model.File{
		ID:             fileID,
		EncryptedName:  req.EncryptedName,
		FileSize:       0,
		OriginalSize:   req.OriginalSize,
		StorageKey:     storageKey,
		StorageBackend: h.config.StorageBackend,
		ChunkCount:     req.ChunkCount,
	}

	if err := h.fileRepo.Create(r.Context(), file); err != nil {
		slog.Error("failed to create file record", "error", err)
		writeError(w, http.StatusInternalServerError, "failed to create file record")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(model.InitFileUploadResponse{
		SecretAccessToken: secretResp.AccessToken,
		FileID:            fileID.String(),
		CreatorToken:      secretResp.CreatorToken,
	})
}

func (h *FileHandler) UploadChunk(w http.ResponseWriter, r *http.Request) {
	fileIDStr := chi.URLParam(r, "id")
	chunkStr := chi.URLParam(r, "n")

	fileID, err := uuid.Parse(fileIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid file id")
		return
	}

	chunkN, err := strconv.Atoi(chunkStr)
	if err != nil || chunkN < 0 {
		writeError(w, http.StatusBadRequest, "invalid chunk number")
		return
	}

	file, err := h.fileRepo.GetByID(r.Context(), fileID)
	if err != nil || file == nil {
		writeError(w, http.StatusNotFound, "file not found")
		return
	}

	if file.UploadComplete {
		writeError(w, http.StatusConflict, "upload already complete")
		return
	}

	key := fmt.Sprintf("%s/chunk_%d", file.StorageKey, chunkN)

	body := http.MaxBytesReader(w, r.Body, 5*1024*1024)
	defer body.Close()

	data, err := io.ReadAll(body)
	if err != nil {
		writeError(w, http.StatusBadRequest, "failed to read chunk data")
		return
	}

	if err := h.storage.Put(r.Context(), key, bytes.NewReader(data), int64(len(data))); err != nil {
		slog.Error("failed to store chunk", "error", err, "chunk", chunkN)
		writeError(w, http.StatusInternalServerError, "failed to store chunk")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *FileHandler) CompleteUpload(w http.ResponseWriter, r *http.Request) {
	fileIDStr := chi.URLParam(r, "id")
	fileID, err := uuid.Parse(fileIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid file id")
		return
	}

	file, err := h.fileRepo.GetByID(r.Context(), fileID)
	if err != nil || file == nil {
		writeError(w, http.StatusNotFound, "file not found")
		return
	}

	for i := 0; i < file.ChunkCount; i++ {
		key := fmt.Sprintf("%s/chunk_%d", file.StorageKey, i)
		exists, err := h.storage.Exists(r.Context(), key)
		if err != nil || !exists {
			writeError(w, http.StatusBadRequest, fmt.Sprintf("chunk %d not uploaded", i))
			return
		}
	}

	if err := h.fileRepo.MarkUploadComplete(r.Context(), fileID, file.OriginalSize); err != nil {
		slog.Error("failed to mark upload complete", "error", err)
		writeError(w, http.StatusInternalServerError, "failed to complete upload")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *FileHandler) DownloadChunk(w http.ResponseWriter, r *http.Request) {
	fileIDStr := chi.URLParam(r, "id")
	chunkStr := chi.URLParam(r, "n")

	fileID, err := uuid.Parse(fileIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid file id")
		return
	}

	chunkN, err := strconv.Atoi(chunkStr)
	if err != nil || chunkN < 0 {
		writeError(w, http.StatusBadRequest, "invalid chunk number")
		return
	}

	file, err := h.fileRepo.GetByID(r.Context(), fileID)
	if err != nil || file == nil || !file.UploadComplete {
		writeError(w, http.StatusNotFound, "file not found")
		return
	}

	key := fmt.Sprintf("%s/chunk_%d", file.StorageKey, chunkN)
	reader, err := h.storage.Get(r.Context(), key)
	if err != nil {
		writeError(w, http.StatusNotFound, "chunk not found")
		return
	}
	defer reader.Close()

	w.Header().Set("Content-Type", "application/octet-stream")
	io.Copy(w, reader)
}
