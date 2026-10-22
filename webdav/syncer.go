package webdav

import (
	"fmt"
	"io/fs"
	"iter"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/matthiasharzer/ilias-sync-webdav/queue"
	"github.com/studio-b12/gowebdav"
)

const (
	webdavInfoFileName               = "#!_WEBDAV_INFORMATION.txt"
	webdavInfoNotDisplayedNotePrefix = "# Folgende Objekte konnten nicht angezeigt werden"
)

func isWebdavInfoFile(remoteFile fileInfo) bool {
	return remoteFile.info.Name() == webdavInfoFileName
}

func remoteToLocalPath(localDirectory, remotePath string) string {
	parts := strings.Split(remotePath, "/")
	for i, part := range parts {
		// Folders on ilias can have leading and trailing spaces, which are not allowed in local file paths. So we trim them.
		parts[i] = strings.TrimSpace(part)
	}
	cleanedUpRemotePath := path.Join(parts...)
	return filepath.Join(localDirectory, cleanedUpRemotePath)
}

type fileInfo struct {
	path string
	info fs.FileInfo
}

type urlSyncer struct {
	client         *gowebdav.Client
	localDirectory string
}

func (s *urlSyncer) iterFiles() iter.Seq2[fileInfo, error] {
	return func(yield func(fileInfo, error) bool) {
		directories := queue.NewQueue("/")

		for !directories.IsEmpty() {
			webdavPath, exists := directories.Dequeue()
			if !exists {
				break
			}

			files, err := s.client.ReadDir(webdavPath)
			if err != nil {
				yield(fileInfo{}, fmt.Errorf("failed to read webdav directory /: %v", err))
				return
			}

			for _, file := range files {
				filePath := path.Join(webdavPath, file.Name())
				if file.IsDir() {
					directories.Enqueue(filePath)
					continue
				}

				if !yield(fileInfo{
					path: filePath,
					info: file,
				}, nil) {
					return
				}
			}
		}
	}
}

func (s *urlSyncer) getWebdavIgnoredFiles(remoteFile fileInfo) (string, []string, error) {
	contents, err := s.client.Read(remoteFile.path)
	if err != nil {
		return "", nil, fmt.Errorf("failed to read webdav info file: %w", err)
	}

	lines := strings.Split(string(contents), "\n")

	if len(lines) == 0 {
		return "", nil, nil
	}

	if !strings.HasPrefix(lines[0], webdavInfoNotDisplayedNotePrefix) {
		return "", nil, nil
	}

	reason := strings.TrimPrefix(lines[0], "# ")

	var ignoredFiles []string
	for _, line := range lines {
		if strings.HasPrefix(line, "#") {
			continue
		}
		cleanedLine := strings.TrimSpace(line)
		if len(cleanedLine) > 0 {
			ignoredFiles = append(ignoredFiles, cleanedLine)
		}
	}
	return reason, ignoredFiles, nil
}

func (s *urlSyncer) downloadFile(remotePath string) (string, error) {
	stream, err := s.client.ReadStream(remotePath)
	if err != nil {
		return "", err
	}
	defer func() {
		_ = stream.Close()
	}()

	localFilePath := remoteToLocalPath(s.localDirectory, remotePath)
	err = os.MkdirAll(filepath.Dir(localFilePath), os.ModePerm)
	if err != nil {
		return "", err
	}

	localFile, err := os.Create(localFilePath)
	if err != nil {
		return "", err
	}
	defer func() {
		_ = localFile.Close()
	}()

	_, err = localFile.ReadFrom(stream)
	if err != nil {
		return "", err
	}

	return localFilePath, nil
}

func (s *urlSyncer) syncAll() iter.Seq[Event] {
	return func(yield func(Event) bool) {
		for file, err := range s.iterFiles() {
			if err != nil {
				if !yield(ErrorEvent{
					Err:        fmt.Errorf("failed to iter files: %w", err),
					RemoteFile: file.path,
				}) {
					return
				}
			}

			if isWebdavInfoFile(file) {
				if s.yieldWebdavIgnoredFiles(yield, file) {
					return
				}
				continue
			}

			localFilePath, err := s.downloadFile(file.path)
			if err != nil {
				if !yield(ErrorEvent{
					Err:        fmt.Errorf("failed to download file %s: %w", file, err),
					RemoteFile: file.path,
				}) {
					return
				}
				continue
			}
			if !yield(FileDownloadedEvent{
				LocalFile:  localFilePath,
				RemoteFile: file.path,
				Stats:      file.info,
			}) {
				return
			}
		}
	}

}

func (s *urlSyncer) yieldWebdavIgnoredFiles(yield func(Event) bool, file fileInfo) bool {
	reason, ignoredFiles, err := s.getWebdavIgnoredFiles(file)
	if err != nil {
		if !yield(ErrorEvent{
			Err:        fmt.Errorf("failed to download file %s: %w", file, err),
			RemoteFile: file.path,
		}) {
			return true
		}
	}
	fileParent := path.Dir(file.path)
	for _, ignoredFile := range ignoredFiles {
		if !yield(FileIgnoredEvent{
			Reason:     reason,
			RemoteFile: path.Join(fileParent, ignoredFile),
			Stats:      file.info,
		}) {
			return true
		}
	}
	return false
}

type Syncer struct {
	username string
	password string
}

func NewSyncer(username, password string) (*Syncer, error) {
	return &Syncer{
		username: username,
		password: password,
	}, nil
}

func (s *Syncer) getClient(url string) (*gowebdav.Client, error) {
	client := gowebdav.NewClient(url, s.username, s.password)
	err := client.Connect()
	if err != nil {
		return nil, err
	}

	return client, nil
}

func (s *Syncer) Sync(url, localDirectory string) (iter.Seq[Event], error) {
	client, err := s.getClient(url)
	if err != nil {
		return nil, err
	}

	syncer := urlSyncer{
		client:         client,
		localDirectory: localDirectory,
	}
	return syncer.syncAll(), nil
}
