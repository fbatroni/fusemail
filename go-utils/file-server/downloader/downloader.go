package downloader

import (
	"errors"
	"fmt"
)

type FileInfo struct {
	FileName     string
	FilePath     string
	FileChecksum string
}

type Downloader struct {
	SomeConfig     string
	DownloaderStep DownloaderStep
}

type DownloaderStep interface {
	GetFile() FileInfo
}

var downloaderStep DownloaderStep

func New(someConfig string) *Downloader {
	i := &Downloader{
		SomeConfig:     someConfig,
		DownloaderStep: downloaderStep,
	}
	return i
}

func (Downloader) StartDownload() (string, error) {

	if downloaderStep == nil {
		return "", errors.New("DownloaderStep not implemented")
	}

	return downloaderStep.GetFile().FileName, nil
}

func RegisterDownloaderImpl(ds DownloaderStep) error {

	fmt.Println("-----------------------------------Registring %v", ds)

	downloaderStep = ds

	return nil
}

func init() {
	fmt.Println("------------INITING DOWNLOADER----------")
}
