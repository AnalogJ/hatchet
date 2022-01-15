package utils

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/mitchellh/go-homedir"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

func ExpandPath(filePath string) (string, error) {
	filePath, err := homedir.Expand(filePath)
	if err != nil {
		return "", err
	}

	filePath, err = filepath.Abs(filePath)
	if err != nil {
		return "", err
	}
	return filePath, nil
}

func FileWrite(filePath string, content string, perm os.FileMode, dryRun bool) error {
	filePath, err := ExpandPath(filePath)
	if err != nil {
		return err
	}

	if dryRun {
		fmt.Printf("%v %v %v:\n",
			color.GreenString("[DRYRUN]"),
			"Would have written content to",
			color.GreenString(filePath),
		)
		color.Green(content)
	} else {
		d1 := []byte(content)
		err = ioutil.WriteFile(filePath, d1, perm)
	}
	return err
}

func FileExists(filePath string) bool {
	filePath, err := ExpandPath(filePath)
	if err != nil {
		return false
	}

	if _, err := os.Stat(filePath); err != nil {
		return !os.IsNotExist(err)
	}
	return true
}

func FileDelete(filePath string) error {
	filePath, err := ExpandPath(filePath)
	if err != nil {
		return err
	}
	return os.Remove(filePath)
}

// CopyFile copies the contents of the file named src to the file named
// by dst. The file will be created if it does not already exist. If the
// destination file exists, all it's contents will be replaced by the contents
// of the source file. The file mode will be copied from the source and
// the copied data is synced/flushed to stable storage.
func CopyFile(src, dst string) (err error) {
	in, err := os.Open(src)
	if err != nil {
		return
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return
	}
	defer func() {
		if e := out.Close(); e != nil {
			err = e
		}
	}()
	_, err = io.Copy(out, in)
	if err != nil {
		return
	}
	err = out.Sync()
	if err != nil {
		return
	}
	si, err := os.Stat(src)
	if err != nil {
		return
	}
	err = os.Chmod(dst, si.Mode())
	if err != nil {
		return
	}
	return
}

// CopyDir recursively copies a directory tree, attempting to preserve permissions.
// Source directory must exist, destination directory must *not* exist.
// Symlinks are ignored and skipped.
func CopyDir(src string, dst string) (err error) {
	src = filepath.Clean(src)
	dst = filepath.Clean(dst)
	si, err := os.Stat(src)
	if err != nil {
		return err
	}
	if !si.IsDir() {
		return fmt.Errorf("source is not a directory")
	}
	_, err = os.Stat(dst)
	if err != nil && !os.IsNotExist(err) {
		return
	}
	if err == nil {
		return fmt.Errorf("destination already exists")
	}
	err = os.MkdirAll(dst, si.Mode())
	if err != nil {
		return
	}
	entries, err := ioutil.ReadDir(src)
	if err != nil {
		return
	}
	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())
		if entry.IsDir() {
			err = CopyDir(srcPath, dstPath)
			if err != nil {
				return
			}
		} else {
			// Skip symlinks.
			if entry.Mode()&os.ModeSymlink != 0 {
				continue
			}
			err = CopyFile(srcPath, dstPath)
			if err != nil {
				return
			}
		}
	}
	return
}
