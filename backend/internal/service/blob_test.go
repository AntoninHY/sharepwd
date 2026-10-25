// SharePwd.io — Burn After Reading
// Zero-knowledge secret sharing. Client-side AES-256-GCM encryption, secrets self-destruct after viewing.
// Copyright (c) 2025-2026 Antonin HILY — CTO, Jizo AI

package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"reflect"
	"strings"
	"testing"
)

// fakeStorage records the keys it was asked to delete and fails the ones listed
// in failOn with the configured error.
type fakeStorage struct {
	deleted []string
	failOn  map[string]error
}

func newFakeStorage() *fakeStorage {
	return &fakeStorage{failOn: map[string]error{}}
}

func (f *fakeStorage) Put(context.Context, string, io.Reader, int64) error { return nil }
func (f *fakeStorage) Get(context.Context, string) (io.ReadCloser, error)  { return nil, nil }
func (f *fakeStorage) Exists(context.Context, string) (bool, error)        { return true, nil }

func (f *fakeStorage) Delete(_ context.Context, key string) error {
	f.deleted = append(f.deleted, key)
	return f.failOn[key]
}

func TestChunkKeysCoversEveryChunk(t *testing.T) {
	got := chunkKeys("files/token/file-id", 3)
	want := []string{
		"files/token/file-id/chunk_0",
		"files/token/file-id/chunk_1",
		"files/token/file-id/chunk_2",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("chunkKeys = %v, want %v", got, want)
	}
}

func TestChunkKeysTreatsNonPositiveCountAsOneChunk(t *testing.T) {
	for _, count := range []int{0, -1} {
		got := chunkKeys("files/token/file-id", count)
		want := []string{"files/token/file-id/chunk_0"}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("chunkKeys with count %d = %v, want %v", count, got, want)
		}
	}
}

func TestDeleteFileBlobsRemovesEveryChunk(t *testing.T) {
	store := newFakeStorage()

	if err := deleteFileBlobs(context.Background(), store, "files/token/file-id", 3); err != nil {
		t.Fatalf("deleteFileBlobs returned %v, want nil", err)
	}

	want := []string{
		"files/token/file-id/chunk_0",
		"files/token/file-id/chunk_1",
		"files/token/file-id/chunk_2",
	}
	if !reflect.DeepEqual(store.deleted, want) {
		t.Fatalf("deleted %v, want %v", store.deleted, want)
	}
}

// A retry after a partial failure must converge: the chunks the first attempt
// removed are gone, and that has to count as success.
func TestDeleteFileBlobsTreatsMissingObjectAsRemoved(t *testing.T) {
	store := newFakeStorage()
	store.failOn["files/token/file-id/chunk_0"] = fs.ErrNotExist
	store.failOn["files/token/file-id/chunk_1"] = fmt.Errorf("wrapped: %w", fs.ErrNotExist)

	if err := deleteFileBlobs(context.Background(), store, "files/token/file-id", 2); err != nil {
		t.Fatalf("deleteFileBlobs returned %v, want nil", err)
	}
}

func TestDeleteFileBlobsAttemptsEveryChunkDespiteFailure(t *testing.T) {
	store := newFakeStorage()
	store.failOn["files/token/file-id/chunk_0"] = errors.New("backend unreachable")

	err := deleteFileBlobs(context.Background(), store, "files/token/file-id", 3)
	if err == nil {
		t.Fatal("deleteFileBlobs returned nil, want an error")
	}

	if len(store.deleted) != 3 {
		t.Fatalf("attempted %d chunks, want 3: one failure must not stop the others", len(store.deleted))
	}
}

// Storage keys embed the secret access token, so they must never reach an error
// string that ends up in a log line.
func TestDeleteFileBlobsErrorDoesNotLeakTheKey(t *testing.T) {
	store := newFakeStorage()
	store.failOn["files/secret-access-token/file-id/chunk_0"] = errors.New("backend unreachable")

	err := deleteFileBlobs(context.Background(), store, "files/secret-access-token/file-id", 1)
	if err == nil {
		t.Fatal("deleteFileBlobs returned nil, want an error")
	}
	if strings.Contains(err.Error(), "secret-access-token") {
		t.Fatalf("error leaks the storage key: %q", err)
	}
}
