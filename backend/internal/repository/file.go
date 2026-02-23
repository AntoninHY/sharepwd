package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jizo-hr/sharepwd/internal/model"
)

type FileRepository struct {
	db *pgxpool.Pool
}

func NewFileRepository(db *pgxpool.Pool) *FileRepository {
	return &FileRepository{db: db}
}

func (r *FileRepository) Create(ctx context.Context, f *model.File) error {
	query := `INSERT INTO files (
		id, secret_id, encrypted_name, file_size, original_size,
		storage_key, storage_backend, chunk_count
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err := r.db.Exec(ctx, query,
		f.ID, f.SecretID, f.EncryptedName, f.FileSize, f.OriginalSize,
		f.StorageKey, f.StorageBackend, f.ChunkCount,
	)
	return err
}

func (r *FileRepository) GetBySecretID(ctx context.Context, secretID uuid.UUID) (*model.File, error) {
	query := `SELECT id, secret_id, encrypted_name, file_size, original_size,
		storage_key, storage_backend, chunk_count, upload_complete, created_at
	FROM files WHERE secret_id = $1`

	var f model.File
	err := r.db.QueryRow(ctx, query, secretID).Scan(
		&f.ID, &f.SecretID, &f.EncryptedName, &f.FileSize, &f.OriginalSize,
		&f.StorageKey, &f.StorageBackend, &f.ChunkCount, &f.UploadComplete,
		&f.CreatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &f, nil
}

func (r *FileRepository) MarkUploadComplete(ctx context.Context, id uuid.UUID, totalSize int64) error {
	query := `UPDATE files SET upload_complete = true, file_size = $2 WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id, totalSize)
	return err
}

func (r *FileRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.File, error) {
	query := `SELECT id, secret_id, encrypted_name, file_size, original_size,
		storage_key, storage_backend, chunk_count, upload_complete, created_at
	FROM files WHERE id = $1`

	var f model.File
	err := r.db.QueryRow(ctx, query, id).Scan(
		&f.ID, &f.SecretID, &f.EncryptedName, &f.FileSize, &f.OriginalSize,
		&f.StorageKey, &f.StorageBackend, &f.ChunkCount, &f.UploadComplete,
		&f.CreatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &f, nil
}
