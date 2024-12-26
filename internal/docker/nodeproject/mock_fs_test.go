package nodeproject

import (
	"os"
	"path/filepath"
	"sync"
	"time"
)

// MockFS implements a mock file system for testing
type MockFS struct {
	files map[string][]byte
	mu    sync.RWMutex
}

func NewMockFS() *MockFS {
	return &MockFS{
		files: make(map[string][]byte),
	}
}

func (m *MockFS) WriteFile(path string, data []byte, perm os.FileMode) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.files[path] = data
	return nil
}

func (m *MockFS) ReadFile(path string) ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if data, ok := m.files[path]; ok {
		return data, nil
	}
	return nil, os.ErrNotExist
}

func (m *MockFS) Stat(path string) (os.FileInfo, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if _, ok := m.files[path]; ok {
		return &mockFileInfo{path: path}, nil
	}
	return nil, os.ErrNotExist
}

func (m *MockFS) MkdirAll(path string, perm os.FileMode) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.files[path] = nil // Directory is represented as nil content
	return nil
}

// mockFileInfo implements os.FileInfo interface
type mockFileInfo struct {
	path string
}

func (m *mockFileInfo) Name() string       { return filepath.Base(m.path) }
func (m *mockFileInfo) Size() int64        { return 0 }
func (m *mockFileInfo) Mode() os.FileMode  { return 0644 }
func (m *mockFileInfo) ModTime() time.Time { return time.Now() }
func (m *mockFileInfo) IsDir() bool        { return false }
func (m *mockFileInfo) Sys() interface{}   { return nil }
