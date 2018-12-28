package fileutil

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"path"
)

var (
	//ErrInvalidFolder happens when a non existing folder or a file path is passed as output folder parameter
	ErrInvalidFolder = errors.New("The given path is not a valid folder")
	//ErrReadWritePerm happens when user has no read/write permission over the output folder parameter
	ErrReadWritePerm = errors.New("User has no permission to write/read files in the given folder")
	// ErrInvalidFilePointer happens when a invalid os.File pointer is passed as parameter
	ErrInvalidFilePointer = errors.New("The given file pointer is invalid")
	// ErrInvalidFilePath happens when a invalid file path is passed as parameter
	ErrInvalidFilePath = errors.New("The given file path is invalid")
	//ErrInvalidReader happens when the ReadCloser passed as parameter is nil
	ErrInvalidReader = errors.New("The given ReadCloser is invalid")
)

// FileNaming is a interface to create a file name. It must create the file name with the extension
type FileNaming interface {
	GetNewFileName() string
}

// FileMeta struct holds the File Name, File Path and File Checksum
type FileMeta struct {
	FileName string
	FilePath string
	Checksum string
}

//GetFileFullPath return the full path of the file that owns this file meta
func (f *FileMeta) GetFileFullPath() string {
	return path.Join(f.FilePath, f.FileName)
}

/*
FileFormat holds an File Type (extension), FileEncode and A FileNaming implementation (returns file name)
These attributes and functions are used to create a new file.
*/
type FileFormat struct {
	FileEncode string
	FileNaming
}

/*
CreateFromReader creates a file in the given folder using data from a io.Reader
To use this method, an implementation of FileNaming must be given whithin the FileFormat passed as parameter
*/
func CreateFromReader(reader io.ReadCloser, folder string, fileFormat FileFormat) (FileMeta, error) {

	var fileMeta FileMeta

	//Check if reader is valid
	if reader == nil {
		return fileMeta, ErrInvalidReader
	}

	err := checkIsFolder(folder)

	if err != nil {
		return fileMeta, ErrInvalidFolder
	}

	fileName := fileFormat.GetNewFileName()
	fileFullPath := path.Join(folder, fileName)

	f, err := os.Create(fileFullPath)

	if err != nil {
		if os.IsPermission(err) {
			return fileMeta, ErrReadWritePerm
		}
		return fileMeta, err
	}

	//Copying the data to the new file
	_, err = io.Copy(f, reader)
	if err != nil {
		return fileMeta, err
	}

	err = f.Close()
	if err != nil {
		return fileMeta, err
	}

	f, err = os.Open(fileFullPath)
	if err != nil {
		return fileMeta, err
	}

	checksum, err := GetChecksum(f)
	if err != nil {
		return fileMeta, err
	}

	err = f.Close()

	fileMeta.Checksum = checksum
	fileMeta.FileName = fileName
	fileMeta.FilePath = folder

	return fileMeta, err
}

//ListFilesInFolder returns an slice of os.FileInfo containing a os.File for each file inside the given folder. Ignores sub-folders
func ListFilesInFolder(folderPath string) ([]os.FileInfo, error) {

	var files []os.FileInfo
	var err error

	fileInfos, err := ioutil.ReadDir(folderPath)
	if err != nil {
		return files, err
	}

	//Filter to exclude folders
	for _, finfo := range fileInfos {

		if !finfo.IsDir() {
			files = append(files, finfo)
		}

	}

	return files, err
}

// GetChecksum creates a checksum for a given file
func GetChecksum(file io.Reader) (string, error) {

	if file == nil {
		return "", ErrInvalidFilePointer
	}

	var checksum string
	hasher := sha256.New()

	//Copying file data to shasher to create the checksum
	_, err := io.Copy(hasher, file)
	if err != nil {
		return checksum, err
	}

	//Get the 16 bytes hash
	hashInBytes := hasher.Sum(nil)[:16]

	//Convert the bytes to a string
	checksum = hex.EncodeToString(hashInBytes)

	return checksum, nil
}

// GetChecksumByName gets the Checksum for the given file
func GetChecksumByName(fileName string) (string, error) {

	f, err := os.Open(fileName)
	defer func() {
		err = f.Close()
	}()

	if err != nil {
		if os.IsNotExist(err) {
			return "", ErrInvalidFilePath
		}
		return "", err
	}

	return GetChecksum(f)
}

//CompareFiles compares two files using it checksum
func CompareFiles(fileA *os.File, fileB *os.File) (bool, error) {

	if fileA == nil || fileB == nil {
		return false, ErrInvalidFilePointer
	}

	checksumA, err := GetChecksum(fileA)
	if err != nil {
		return false, err
	}
	checksumB, err := GetChecksum(fileB)

	ok := (checksumA == checksumB)

	return ok, err
}

//CompareFilesByName compares two files based on the checksum of each, using the files' name to fetch their data
func CompareFilesByName(fileAName string, fileBName string) (bool, error) {

	checksumA, err := GetChecksumByName(fileAName)
	if err != nil {
		return false, err
	}
	checksumB, err := GetChecksumByName(fileBName)

	return checksumA == checksumB, err
}

func checkIsFolder(folder string) error {

	info, err := os.Stat(folder)

	if err != nil && !os.IsNotExist(err) {
		return err
	} else if os.IsNotExist(err) {
		return ErrInvalidFolder
	}

	if !info.IsDir() {
		return ErrInvalidFolder
	}

	return nil
}
