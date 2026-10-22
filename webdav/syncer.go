package webdav

import (
	"fmt"
	"iter"
	"os"
	"path"
	"path/filepath"

	"github.com/matthiasharzer/ilias-sync-webdav/logging"
	"github.com/matthiasharzer/ilias-sync-webdav/queue"
	"github.com/studio-b12/gowebdav"
)

func iterFiles(client *gowebdav.Client) iter.Seq2[string, error] {
	return func(yield func(string, error) bool) {
		directories := queue.NewQueue("/")

		for !directories.IsEmpty() {
			webdavPath, exists := directories.Dequeue()
			if !exists {
				break
			}

			files, err := client.ReadDir(webdavPath)
			if err != nil {
				yield("", fmt.Errorf("failed to read webdav directory /: %v", err))
				return
			}

			for _, file := range files {
				filePath := path.Join(webdavPath, file.Name())
				if file.IsDir() {
					directories.Enqueue(filePath)
					continue
				}

				if !yield(filePath, nil) {
					return
				}
			}
		}
	}
}

func downloadFile(client *gowebdav.Client, remotePath, localDirectory string) error {
	stream, err := client.ReadStream(remotePath)
	if err != nil {
		return err
	}
	defer func() {
		err := stream.Close()
		if err != nil {
			logging.Error("Failed to close stream", "error", err)
		}
	}()

	localFilePath := filepath.Join(localDirectory, remotePath)
	err = os.MkdirAll(filepath.Dir(localFilePath), os.ModePerm)
	if err != nil {
		return err
	}

	localFile, err := os.Create(localFilePath)
	if err != nil {
		return err
	}
	defer func() {
		err := localFile.Close()
		if err != nil {
			logging.Error("Failed to close file", "error", err)
		}
	}()

	_, err = localFile.ReadFrom(stream)
	if err != nil {
		return err
	}

	logging.Info("Downloaded file", "remotePath", remotePath, "localPath", localFilePath)
	return nil
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

func (s *Syncer) Sync(url, localDirectory string) error {
	client, err := s.getClient(url)
	if err != nil {
		return err
	}

	for file, err := range iterFiles(client) {
		if err != nil {
			return fmt.Errorf("failed to iter files: %w", err)
		}

		err := downloadFile(client, file, localDirectory)
		if err != nil {
			return fmt.Errorf("failed to download file %s: %w", file, err)
		}
	}

	return nil
}
