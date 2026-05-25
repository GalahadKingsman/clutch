package storage

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

type LocalStore struct {
	Root string
}

func NewLocalStore(root string) (*LocalStore, error) {
	if root == "" {
		root = "data/uploads"
	}
	if err := os.MkdirAll(root, 0o755); err != nil {
		return nil, err
	}
	return &LocalStore{Root: root}, nil
}

func (s *LocalStore) SaveProof(duelID, userID uuid.UUID, filename string, r io.Reader) (path string, hash string, err error) {
	ext := filepath.Ext(filename)
	if ext == "" {
		ext = ".bin"
	}
	dir := filepath.Join(s.Root, duelID.String(), userID.String())
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", "", err
	}
	id := uuid.New().String()
	rel := filepath.Join(duelID.String(), userID.String(), id+ext)
	abs := filepath.Join(s.Root, rel)
	f, err := os.Create(abs)
	if err != nil {
		return "", "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(io.MultiWriter(f, h), r); err != nil {
		return "", "", err
	}
	return rel, hex.EncodeToString(h.Sum(nil)), nil
}

func (s *LocalStore) Open(rel string) (*os.File, error) {
	rel = filepath.Clean(rel)
	if strings.Contains(rel, "..") {
		return nil, fmt.Errorf("invalid path")
	}
	return os.Open(filepath.Join(s.Root, rel))
}

func (s *LocalStore) PublicPath(rel string) string {
	return "/api/v1/files/" + strings.ReplaceAll(rel, string(os.PathSeparator), "/")
}
