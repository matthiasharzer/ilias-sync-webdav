package webdav

import "io/fs"

type Event interface {
	discriminator()
}

type FileDownloadedEvent struct {
	LocalFile  string
	RemoteFile string
	Stats      fs.FileInfo
}

func (FileDownloadedEvent) discriminator() {}

type FileDownloadSkippedEvent struct {
	Reason     string
	RemoteFile string
	Stats      fs.FileInfo
}

func (FileDownloadSkippedEvent) discriminator() {}

type FileIgnoredEvent struct {
	Reason     string
	RemoteFile string
	Stats      fs.FileInfo
}

func (FileIgnoredEvent) discriminator() {}

type ErrorEvent struct {
	RemoteFile string
	Err        error
}

func (ErrorEvent) discriminator() {}
