package utils

import (
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
)

// MkdirAllIsExist creates subDirs recursively (using os.Mkdir) based on baseDir.
// If baseDir does not exist, return error; if baseDir/subDirs exists, return
// the item-exist-error (e.g. os.ErrExist) returned by package os.
// NOTE: on error return, there might exist one ore more already created subDir; it is
//       caller's responsibility for cleaning up (so this method can be lock-free).
func MkdirAllIsExist(mode os.FileMode, baseDir string, subDirs ...string) (string, error) {
	if _, err := os.Stat(baseDir); err != nil {
		return "", err
	}

	var err error
	fullFolderPath, err := filepath.Abs(baseDir)
	if err != nil {
		return "", err
	}

	for _, subDir := range subDirs {
		fullFolderPath = path.Join(fullFolderPath, subDir)
		if err = os.Mkdir(fullFolderPath, mode); err != nil {
			if os.IsExist(err) {
				continue
			}
			return "", err
		}
	}

	return fullFolderPath, err
}

// MkdirAll is similar to MkdirAllIsExist() but does not return item-exist-error if
// baseDir/subDirs exists; instead, it returns nil followed by the full-path
// of baseDir/subDirs.
func MkdirAll(mode os.FileMode, baseDir string, subDirs ...string) (string, error) {
	fullPath, err := MkdirAllIsExist(mode, baseDir, subDirs...)
	if err != nil && os.IsExist(err) {
		return fullPath, nil
	}
	return fullPath, err
}

// IsFolderEmpty returns true if input folder is empty, else return false
// (err might or might not be empty).
func IsFolderEmpty(name string) (bool, error) {
	f, err := os.Open(name)
	if err != nil {
		return false, err
	}
	defer f.Close()

	_, err = f.Readdirnames(1) // Or f.Readdir(1)
	if err == io.EOF {
		return true, nil
	}
	return false, err // Either not empty or error, suits both cases
}

type walkFunc func(dirPath string) bool

// RmdirIfEmtpy traverses all input baseDir's sub-folders and delete
// any empty sub-folders. If deleteBase is TRUE and baseDir is empty
// (or becomes empty after traversed/rm its empty sub-folders), then
// baseDir will also be deleted. Returns sub folder paths being deleted
// or corresponding error during the traversion/deletion.
func RmdirIfEmtpy(baseDir string, deleteBase bool) ([]string, error) {

	rmDirList := []string{}

	absBaseDir, err := filepath.Abs(baseDir)
	if err != nil {
		return rmDirList, err
	}
	rmDirList, err = WalkDirRecurse(absBaseDir, func(dirPath string) bool {
		isEmpty, _ := IsFolderEmpty(dirPath)
		if isEmpty && os.Remove(dirPath) == nil {
			return true
		}
		return false
	})
	if err == nil && deleteBase {
		isEmpty, errSub := IsFolderEmpty(absBaseDir)
		if isEmpty {
			if errRm := os.Remove(absBaseDir); errRm == nil {
				rmDirList = append(rmDirList, absBaseDir)
				return rmDirList, nil
			}
		}
		err = errSub
	}
	return rmDirList, err
}

// WalkDirRecurse recursively traverses input searchDir and calls input
// walkFn for each traversed directory from the deepest sub directoy.
func WalkDirRecurse(searchDir string, walkFn walkFunc) ([]string, error) {
	walkedDirs := []string{}
	toBeWalked, err := filepath.Glob(path.Join(searchDir, "*"))
	if err != nil {
		return walkedDirs, err
	}

	var errs []error
	// call walkFn for every sub folder traversed (deepest sub-folder first)
	for _, walkPath := range toBeWalked {
		fi, err := os.Stat(walkPath)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		if !fi.IsDir() {
			continue
		}
		successSubs, err2 := WalkDirRecurse(walkPath, walkFn)
		walkedDirs = append(walkedDirs, successSubs...)
		if err2 != nil {
			return walkedDirs, err2
		}
		if walkFn(walkPath) {
			walkedDirs = append(walkedDirs, walkPath)
		}
	}
	if len(errs) > 0 {
		err = fmt.Errorf("err stat: %v", errs)
	}

	return walkedDirs, err
}
