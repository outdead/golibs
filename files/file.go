package files

import (
	"errors"
	"fmt"
	"os"
	"syscall"
	"time"
)

var ErrFolderAlreadyExist = errors.New("folder already exists")

// FileExists checks if a file exists and is not a directory before we
// try using it to prevent further errors.
func FileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}

	return !info.IsDir()
}

// FileCopy copies src file to destination path.
func FileCopy(src string, destination string, perm ...os.FileMode) error {
	p := os.ModePerm
	if len(perm) != 0 {
		p = perm[0]
	}

	input, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	return os.WriteFile(destination, input, p)
}

// ReadStringFile reads file as string.
func ReadStringFile(path string, name string) (string, error) {
	b, err := os.ReadFile(path + "/" + name)
	if err != nil {
		return "", err
	}

	return string(b), nil
}

// ReadBinFile reads file as slice of bytes.
func ReadBinFile(path string, name string) ([]byte, error) {
	b, err := os.ReadFile(path + "/" + name)
	if err != nil {
		return []byte{}, err
	}

	return b, nil
}

// WriteFileString writes string content to text file.
func WriteFileString(path string, name string, value string) error {
	fo, err := os.Create(path + "/" + name)
	if err != nil {
		return err
	}

	if _, err := fo.Write([]byte(value)); err != nil {
		return err
	}

	return nil
}

// StatTimes gets file stats info.
func StatTimes(name string) (atime, mtime, ctime time.Time, err error) {
	fi, err := os.Stat(name)
	if err != nil {
		return
	}

	mtime = fi.ModTime()
	stat := fi.Sys().(*syscall.Stat_t) //nolint
	atime = time.Unix(stat.Atim.Sec, stat.Atim.Nsec)
	ctime = time.Unix(stat.Ctim.Sec, stat.Ctim.Nsec)

	return
}

// MkdirAll creates a directory named path,
// along with any necessary parents, and returns nil,
// or else returns an error.
// The permission bits perm (before umask) are used for all
// directories that MkdirAll creates.
// If path is already a directory, MkdirAll returns error.
func MkdirAll(path string, perm ...os.FileMode) error {
	_, err := os.Stat(path)

	switch {
	case err == nil:
		return fmt.Errorf("%w: %s", ErrFolderAlreadyExist, path)
	case os.IsNotExist(err):
		p := os.ModePerm
		if len(perm) != 0 {
			p = perm[0]
		}

		return os.MkdirAll(path, p)
	}

	return err
}

func GetDirNamesInFolder(path string) ([]string, error) {
	items, err := os.ReadDir(path)
	if err != nil {
		return make([]string, 0), fmt.Errorf("scan dirrectory: %w", err)
	}

	names := make([]string, 0, len(items))

	for _, item := range items {
		if item.IsDir() {
			names = append(names, item.Name())
		}
	}

	return names, nil
}

func GetFileNamesInFolder(path string) ([]string, error) {
	items, err := os.ReadDir(path)
	if err != nil {
		return make([]string, 0), fmt.Errorf("scan dirrectory: %w", err)
	}

	names := make([]string, 0, len(items))

	for _, item := range items {
		if !item.IsDir() {
			names = append(names, item.Name())
		}
	}

	return names, nil
}
