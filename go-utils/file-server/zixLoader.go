package main

import (
	"fmt"

	"gustavo.org/fusemail/go-utils/file-server/downloader"
)

type ZixDownloader struct {
	FileName string
}

func (d ZixDownloader) GetFile() downloader.FileInfo {
	f := downloader.FileInfo{
		FileName: d.FileName,
	}
	return f
}

func int() {
	fmt.Println("-------------------------------------Initing Zix Plugin---------------------------------")
	zixDownloader := ZixDownloader{FileName: "zix.csv"}
	downloader.RegisterDownloaderImpl(zixDownloader)
}
