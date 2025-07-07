package files

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

// OwnerWritePerm provides 0755 permission.
const OwnerWritePerm = os.FileMode(0o755)

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
func FileCopy(src string, destination string, perms ...os.FileMode) error {
	perm := os.ModePerm
	if len(perms) != 0 {
		perm = perms[0]
	}

	input, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	return os.WriteFile(destination, input, perm)
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

	if _, err := fo.WriteString(value); err != nil {
		return err
	}

	return nil
}

// CreateAndOpenFile creates file and open it foe recording.
func CreateAndOpenFile(path string, fileName string, perms ...os.FileMode) (io.Writer, error) {
	perm := OwnerWritePerm
	if len(perms) != 0 {
		perm = perms[0]
	}

	filePath := path
	if filePath != "" {
		// Not dependent on the transmitted "/".
		filePath = strings.TrimRight(filePath, "/") + "/"

		if err := MkdirAll(path); err != nil {
			return nil, err
		}
	}

	filePath += fileName

	f, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, perm)
	if err != nil {
		return nil, err
	}

	return f, nil
}

// StatTimes gets file stats info.
func StatTimes(name string) (atime, mtime, ctime time.Time, err error) {
	info, err := os.Stat(name)
	if err != nil {
		return
	}

	mtime = info.ModTime()
	stat := info.Sys().(*syscall.Stat_t) //nolint
	atime = time.Unix(stat.Atim.Sec, stat.Atim.Nsec)
	ctime = time.Unix(stat.Ctim.Sec, stat.Ctim.Nsec)

	return
}

// MkdirAll creates a directory named path,
// along with any necessary parents, and returns nil,
// or else returns an error.
// The permission bits perm (before umask) are used for all
// directories that MkdirAll creates.
func MkdirAll(path string, perm ...os.FileMode) error {
	_, err := os.Stat(path)

	if os.IsNotExist(err) {
		p := os.ModePerm
		if len(perm) != 0 {
			p = perm[0]
		}

		return os.MkdirAll(path, p)
	}

	return err
}

// GetDirNamesInFolder returns slice with directory names in path.
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

// GetFileNamesInFolder returns slice with file names in path.
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

// GetAbsPath returns an absolute path based on the input path and a default path.
// It handles three cases for the input path:
//  1. If the path is already absolute (starts with "/"), home-relative (starts with "~/"),
//     or relative to current directory (starts with "./"), it returns the path as-is.
//  2. If the path is empty, it uses the defaultPath instead.
//  3. For all other cases, it treats the path as relative to the executable's directory.
//
// Parameters:
//   - path: The input path to process (can be empty, absolute, or relative)
//   - defaultPath: The default path to use if input path is empty
//
// Returns:
//   - string: The resulting absolute path
//   - error: Any error that occurred while getting the executable's directory
//
// Example usage:
//
//	absPath, err := GetAbsPath("config.json", "/etc/default/config.json")
//	// Returns "/path/to/executable/config.json" if no error
func GetAbsPath(path, defaultPath string) (string, error) {
	// Check if path is already in absolute, home-relative, or current-directory-relative form
	if path != "" && (string(path[0]) == "/" || strings.HasPrefix(path, "~/") || strings.HasPrefix(path, "./")) {
		return path, nil
	}

	// Get the absolute path of the directory containing the executable
	home, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return "", err
	}

	// Use default path if input path is empty
	if path == "" {
		path = defaultPath
	}

	// Combine executable directory with the relative path
	path = home + "/" + path

	return path, nil
}
