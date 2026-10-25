// SharePwd.io — Burn After Reading
// Zero-knowledge secret sharing. Client-side AES-256-GCM encryption, secrets self-destruct after viewing.
// Copyright (c) 2025-2026 Antonin HILY — CTO, Jizo AI

package service

import (
	"context"
	"errors"
	"fmt"
	"io/fs"

	"github.com/AntoninHY/sharepwd/internal/storage"
)

// chunkKeys returns the object keys holding the chunks of a stored file, in the
// same layout the upload path writes them (see handler.FileHandler.UploadChunk).
// A stored file always has at least one chunk, so a non-positive count coming
// from the database is read as one rather than silently leaving the blob behind.
func chunkKeys(storageKey string, chunkCount int) []string {
	if chunkCount < 1 {
		chunkCount = 1
	}

	keys := make([]string, 0, chunkCount)
	for i := 0; i < chunkCount; i++ {
		keys = append(keys, fmt.Sprintf("%s/chunk_%d", storageKey, i))
	}
	return keys
}

// deleteFileBlobs removes every chunk of a stored file from object storage.
//
// A chunk that is already gone counts as removed, so a retry after a partial
// failure converges instead of failing on the chunks the previous attempt
// deleted. Every chunk is attempted even when one of them fails, and the
// returned error reports the chunk indexes that could not be removed.
//
// Keys embed the secret access token, so they are never included in errors or
// logs; only chunk indexes are.
func deleteFileBlobs(ctx context.Context, store storage.Storage, storageKey string, chunkCount int) error {
	var failed []int

	for i, key := range chunkKeys(storageKey, chunkCount) {
		if err := store.Delete(ctx, key); err != nil && !isAlreadyGone(err) {
			failed = append(failed, i)
		}
	}

	if len(failed) > 0 {
		return fmt.Errorf("delete file blobs: %d of %d chunks remain, indexes %v",
			len(failed), len(chunkKeys(storageKey, chunkCount)), failed)
	}
	return nil
}

// isAlreadyGone reports whether an error means the object is no longer there,
// which makes the deletion a success. The local backend wraps os.Remove and
// surfaces fs.ErrNotExist; the S3 backend treats DeleteObject as idempotent and
// returns no error at all for a missing key.
func isAlreadyGone(err error) bool {
	return errors.Is(err, fs.ErrNotExist)
}
