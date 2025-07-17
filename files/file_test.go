package files

import (
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const TestFilesDir = "./.tmp"

func setupTest(t *testing.T) func(t *testing.T) {
	os.RemoveAll(TestFilesDir)
	os.Mkdir(TestFilesDir, 0o777)

	return func(t *testing.T) {
		os.RemoveAll(TestFilesDir)
	}
}

func TestFileExists(t *testing.T) {
	down := setupTest(t)
	defer down(t)

	t.Run("existing file", func(t *testing.T) {
		tmpfile, err := os.CreateTemp(TestFilesDir, "testfile")
		require.NoError(t, err)
		defer os.Remove(tmpfile.Name())

		assert.True(t, FileExists(tmpfile.Name()))
	})

	t.Run("non-existent file", func(t *testing.T) {
		assert.False(t, FileExists("/nonexistent/file"))
	})

	t.Run("directory", func(t *testing.T) {
		dir, err := os.MkdirTemp(TestFilesDir, "testdir")
		require.NoError(t, err)
		defer os.RemoveAll(dir)

		assert.False(t, FileExists(dir))
	})
}

func TestFileCopy(t *testing.T) {
	down := setupTest(t)
	defer down(t)

	t.Run("successful copy", func(t *testing.T) {
		src, err := os.CreateTemp(TestFilesDir, "src")
		require.NoError(t, err)
		defer os.Remove(src.Name())
		src.WriteString("test content")

		dst := filepath.Join(TestFilesDir, "dstfile")
		defer os.Remove(dst)

		err = FileCopy(src.Name(), dst)
		require.NoError(t, err)

		content, err := os.ReadFile(dst)
		require.NoError(t, err)
		assert.Equal(t, "test content", string(content))
	})

	t.Run("nonexistent source", func(t *testing.T) {
		err := FileCopy("/nonexistent/src", "/tmp/dst")
		assert.Error(t, err)
	})

	t.Run("custom permissions", func(t *testing.T) {
		src, err := os.CreateTemp(TestFilesDir, "src")
		require.NoError(t, err)
		defer os.Remove(src.Name())

		dst := filepath.Join(TestFilesDir, "dstfile")
		defer os.Remove(dst)

		err = FileCopy(src.Name(), dst, 0644)
		require.NoError(t, err)

		info, err := os.Stat(dst)
		require.NoError(t, err)
		assert.Equal(t, os.FileMode(0644), info.Mode().Perm())
	})
}

func TestReadStringFile(t *testing.T) {
	down := setupTest(t)
	defer down(t)

	t.Run("successful read", func(t *testing.T) {
		dir, err := os.MkdirTemp(TestFilesDir, "testdir")
		require.NoError(t, err)
		defer os.RemoveAll(dir)

		err = os.WriteFile(filepath.Join(dir, "test.txt"), []byte("hello world"), 0644)
		require.NoError(t, err)

		content, err := ReadStringFile(dir, "test.txt")
		require.NoError(t, err)
		assert.Equal(t, "hello world", content)
	})

	t.Run("nonexistent file", func(t *testing.T) {
		_, err := ReadStringFile("/nonexistent", "file.txt")
		assert.Error(t, err)
	})
}

func TestReadBinFile(t *testing.T) {
	down := setupTest(t)
	defer down(t)

	t.Run("successful read", func(t *testing.T) {
		dir, err := os.MkdirTemp(TestFilesDir, "testdir")
		require.NoError(t, err)
		defer os.RemoveAll(dir)

		testData := []byte{0x01, 0x02, 0x03}
		err = os.WriteFile(filepath.Join(dir, "test.bin"), testData, 0644)
		require.NoError(t, err)

		content, err := ReadBinFile(dir, "test.bin")
		require.NoError(t, err)
		assert.Equal(t, testData, content)
	})
}

func TestWriteFileString(t *testing.T) {
	down := setupTest(t)
	defer down(t)

	t.Run("successful write", func(t *testing.T) {
		dir, err := os.MkdirTemp(TestFilesDir, "testdir")
		require.NoError(t, err)
		defer os.RemoveAll(dir)

		err = WriteFileString(dir, "test.txt", "test content")
		require.NoError(t, err)

		content, err := os.ReadFile(filepath.Join(dir, "test.txt"))
		require.NoError(t, err)
		assert.Equal(t, "test content", string(content))
	})

	t.Run("invalid path", func(t *testing.T) {
		err := WriteFileString("/nonexistent/path", "test.txt", "content")
		assert.Error(t, err)
	})
}

func TestCreateAndOpenFile(t *testing.T) {
	down := setupTest(t)
	defer down(t)

	t.Run("successful create", func(t *testing.T) {
		dir, err := os.MkdirTemp(TestFilesDir, "testdir")
		require.NoError(t, err)
		defer os.RemoveAll(dir)

		writer, err := CreateAndOpenFile(dir, "test.txt")
		require.NoError(t, err)
		defer writer.(io.Closer).Close()

		_, err = writer.Write([]byte("test"))
		require.NoError(t, err)

		content, err := os.ReadFile(filepath.Join(dir, "test.txt"))
		require.NoError(t, err)
		assert.Equal(t, "test", string(content))
	})

	t.Run("custom permissions", func(t *testing.T) {
		down := setupTest(t)
		defer down(t)

		dir, err := os.MkdirTemp(TestFilesDir, "testdir")
		require.NoError(t, err)
		defer os.RemoveAll(dir)

		writer, err := CreateAndOpenFile(dir, "test.txt", 0600)
		require.NoError(t, err)
		writer.(io.Closer).Close()

		info, err := os.Stat(filepath.Join(dir, "test.txt"))
		require.NoError(t, err)
		assert.Equal(t, os.FileMode(0600), info.Mode().Perm())
	})

	t.Run("nonexistent directory", func(t *testing.T) {
		_, err := CreateAndOpenFile("/nonexistent/path", "test.txt")
		assert.Error(t, err)
	})
}

func TestStatTimes(t *testing.T) {
	down := setupTest(t)
	defer down(t)

	t.Run("successful stat", func(t *testing.T) {
		tmpfile, err := os.CreateTemp(TestFilesDir, "testfile")
		require.NoError(t, err)
		defer os.Remove(tmpfile.Name())

		atime, mtime, ctime, err := StatTimes(tmpfile.Name())
		require.NoError(t, err)

		assert.False(t, atime.IsZero())
		assert.False(t, mtime.IsZero())
		assert.False(t, ctime.IsZero())
	})

	t.Run("nonexistent file", func(t *testing.T) {
		_, _, _, err := StatTimes("/nonexistent/file")
		assert.Error(t, err)
	})
}

func TestMkdirAll(t *testing.T) {
	down := setupTest(t)
	defer down(t)

	t.Run("successful create", func(t *testing.T) {
		dir := filepath.Join(TestFilesDir, "testdir")
		defer os.RemoveAll(dir)

		err := MkdirAll(dir)
		require.NoError(t, err)

		assert.DirExists(t, dir)
	})

	t.Run("custom permissions", func(t *testing.T) {
		dir := filepath.Join(TestFilesDir, "testdir")
		defer os.RemoveAll(dir)

		err := MkdirAll(dir, 0750)
		require.NoError(t, err)

		info, err := os.Stat(dir)
		require.NoError(t, err)
		assert.Equal(t, os.FileMode(0750), info.Mode().Perm())
	})

	t.Run("existing directory", func(t *testing.T) {
		dir, err := os.MkdirTemp(TestFilesDir, "testdir")
		require.NoError(t, err)
		defer os.RemoveAll(dir)

		err = MkdirAll(dir)
		assert.NoError(t, err)
	})
}

func TestGetDirNamesInFolder(t *testing.T) {
	down := setupTest(t)
	defer down(t)

	t.Run("successful read", func(t *testing.T) {
		dir, err := os.MkdirTemp(TestFilesDir, "testdir")
		require.NoError(t, err)
		defer os.RemoveAll(dir)

		err = os.Mkdir(filepath.Join(dir, "dir1"), 0755)
		require.NoError(t, err)

		err = os.Mkdir(filepath.Join(dir, "dir2"), 0755)
		require.NoError(t, err)

		err = os.WriteFile(filepath.Join(dir, "file.txt"), []byte("test"), 0644)
		require.NoError(t, err)

		names, err := GetDirNamesInFolder(dir)
		require.NoError(t, err)

		assert.ElementsMatch(t, []string{"dir1", "dir2"}, names)
	})

	t.Run("nonexistent directory", func(t *testing.T) {
		_, err := GetDirNamesInFolder("/nonexistent")
		assert.Error(t, err)
	})
}

func TestGetFileNamesInFolder(t *testing.T) {
	down := setupTest(t)
	defer down(t)

	t.Run("successful read", func(t *testing.T) {
		dir, err := os.MkdirTemp(TestFilesDir, "testdir")
		require.NoError(t, err)
		defer os.RemoveAll(dir)

		err = os.WriteFile(filepath.Join(dir, "file1.txt"), []byte("test"), 0644)
		require.NoError(t, err)

		err = os.WriteFile(filepath.Join(dir, "file2.txt"), []byte("test"), 0644)
		require.NoError(t, err)

		err = os.Mkdir(filepath.Join(dir, "subdir"), 0755)
		require.NoError(t, err)

		names, err := GetFileNamesInFolder(dir)
		require.NoError(t, err)

		assert.ElementsMatch(t, []string{"file1.txt", "file2.txt"}, names)
	})
}

func TestGetAbsPath(t *testing.T) {
	down := setupTest(t)
	defer down(t)

	t.Run("absolute path", func(t *testing.T) {
		path := "/absolute/path"
		result, err := GetAbsPath(path, "")
		require.NoError(t, err)
		assert.Equal(t, path, result)
	})

	t.Run("home-relative path", func(t *testing.T) {
		path := "~/relative/path"
		result, err := GetAbsPath(path, "")
		require.NoError(t, err)
		assert.Equal(t, path, result)
	})

	t.Run("current-dir-relative path", func(t *testing.T) {
		path := "./relative/path"
		result, err := GetAbsPath(path, "")
		require.NoError(t, err)
		assert.Equal(t, path, result)
	})

	t.Run("empty path with default", func(t *testing.T) {
		defaultPath := "/default/path"
		result, err := GetAbsPath("", defaultPath)
		require.NoError(t, err)
		assert.Equal(t, defaultPath, result)
	})

	t.Run("relative path", func(t *testing.T) {
		exeDir, err := filepath.Abs(filepath.Dir(os.Args[0]))
		require.NoError(t, err)

		result, err := GetAbsPath("relative/path", "")
		require.NoError(t, err)
		assert.Equal(t, filepath.Join(exeDir, "relative/path"), result)
	})
}

func TestClearDir(t *testing.T) {
	down := setupTest(t)
	defer down(t)

	createTestFile := func(dir, name string) string {
		path := filepath.Join(dir, name)
		err := os.WriteFile(path, []byte("test"), 0644)
		require.NoError(t, err)
		return path
	}

	t.Run("should clear directory with files", func(t *testing.T) {
		dir, err := os.MkdirTemp(TestFilesDir, "testdir")
		require.NoError(t, err)
		defer os.RemoveAll(dir)

		createTestFile(dir, "file1.txt")
		createTestFile(dir, "file2.txt")

		err = ClearDir(dir)
		require.NoError(t, err)

		entries, err := os.ReadDir(dir)
		require.NoError(t, err)
		assert.Empty(t, entries, "Directory should be empty")
	})

	t.Run("should clear directory with subdirectories", func(t *testing.T) {
		dir, err := os.MkdirTemp(TestFilesDir, "testdir")
		require.NoError(t, err)
		defer os.RemoveAll(dir)

		err = os.Mkdir(filepath.Join(dir, "subdir"), 0755)
		require.NoError(t, err)
		createTestFile(filepath.Join(dir, "subdir"), "nested.txt")

		err = ClearDir(dir)
		require.NoError(t, err)

		entries, err := os.ReadDir(dir)
		require.NoError(t, err)
		assert.Empty(t, entries, "Directory should be empty")
	})

	t.Run("should preserve empty directory", func(t *testing.T) {
		dir, err := os.MkdirTemp(TestFilesDir, "testdir")
		require.NoError(t, err)
		defer os.RemoveAll(dir)

		err = ClearDir(dir)
		require.NoError(t, err)

		_, err = os.Stat(dir)
		assert.NoError(t, err, "Directory should still exist")
	})

	t.Run("should handle symbolic links", func(t *testing.T) {
		dir, err := os.MkdirTemp(TestFilesDir, "testdir")
		require.NoError(t, err)
		defer os.RemoveAll(dir)

		targetFile := createTestFile(dir, "target.txt")
		linkPath := filepath.Join(dir, "link.txt")
		err = os.Symlink(targetFile, linkPath)
		require.NoError(t, err)

		err = ClearDir(dir)
		require.NoError(t, err)

		entries, err := os.ReadDir(dir)
		require.NoError(t, err)
		assert.Empty(t, entries, "Directory should be empty")
	})
}
