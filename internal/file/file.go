package file

import (
	"os"
	"path/filepath"
)

type File struct {
	homeDir string
	f       *os.File
}

func New() *File {
	return &File{}
}

// CreateDir creates a new directory if not exist.
func (f *File) CreateDir(dirs ...string) (string, error) {
	if f.homeDir == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		f.homeDir = homeDir
	}

	dirs = append([]string{f.homeDir}, dirs...)
	dir := filepath.Join(dirs...)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return "", err
	}
	return dir, nil
}

// OpenFile opens a file which is given as parameter.
func (f *File) OpenFile(path ...string) (*os.File, error) {
	filePath := filepath.Join(path...)
	logFile, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	f.f = logFile
	return logFile, nil
}

// Close closes the file.
func (f *File) Close() error {
	return f.f.Close()
}
