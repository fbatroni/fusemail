package utils 

import (
	"fmt"
	"io"
	"os"
	"path"
	"reflect"
	"syscall"
)

// MoveFile moves srcFilePath to dstFilePath; depending on src and dst
// partitions, MoveFile behaves as:
//  1. if src & dst are on the same partition, MoveFile renames src to
//     dst directly (overwrites existing file if any).
//  2. if src & dst are on different partitions, MoveFile copies src
//     to dst as a tmp file, rename it (overwrites existing file if any)
//     and deletes src.
func MoveFile(srcFilePath, dstFilePath string) error {
	// if src & dst share same device-id, just rename it
	srcDevID, err := GetDeviceIDFromPath(srcFilePath)
	if err == nil {
		dstDevID, err := GetDeviceIDFromPath(path.Dir(dstFilePath)) // dst-file might not exist, use its dir
		if err == nil && srcDevID == dstDevID {
			return os.Rename(srcFilePath, dstFilePath)
		}
	}

	// anything else, copy and rename
	dstTmpPath := fmt.Sprint(dstFilePath, ".tmp")
	if err = CopyFile(srcFilePath, dstTmpPath); err != nil {
		return err
	}
	err = os.Rename(dstTmpPath, dstFilePath)
	if err != nil {
		os.Remove(dstTmpPath)
		return err
	}

	// last, delete src-file
	return os.Remove(srcFilePath)
}

// CopyFile copies srcPath to dstPath preserving file mode.
// NOTE: srcPath and dstPath must represent 'files'.
func CopyFile(srcPath, dstPath string) error {
	// open src file
	srcFile, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer srcFile.Close()
	si, err := srcFile.Stat()
	if err != nil {
		return err
	}

	// create dst file
	dstFile, err := os.OpenFile(dstPath, os.O_WRONLY|os.O_CREATE, si.Mode())
	if err != nil {
		return err
	}
	defer dstFile.Close()

	// copy file content
	_, err = io.Copy(dstFile, srcFile)
	return err

}

// GetDeviceIDFromPath returns Device-ID of given srcPath.
// It is usually used to check if two paths are under the
// same partition (which share same Device-ID).
func GetDeviceIDFromPath(srcPath string) (uint64, error) {
	fi, err := os.Stat(srcPath)
	if err != nil {
		return 0, err
	}
	if s, ok := fi.Sys().(*syscall.Stat_t); ok {
		return s.Dev, nil
	}
	return 0, fmt.Errorf("unexpected type %T", reflect.TypeOf(fi.Sys()))
}
