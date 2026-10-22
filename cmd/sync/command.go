package sync

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"

	"github.com/matthiasharzer/ilias-sync-webdav/config"
	"github.com/matthiasharzer/ilias-sync-webdav/webdav"
	"github.com/spf13/cobra"
)

var configFile string

func init() {
	Command.Flags().StringVarP(&configFile, "config", "c", "", "Path to the configuration file")

	err := Command.MarkFlagRequired("config")
	if err != nil {
		panic(err)
	}
}

func getILIASWebDavURL(baseURL url.URL, courseID string) string {
	return fmt.Sprintf("%s/ref_%s/", baseURL.String(), courseID)
}

var Command = &cobra.Command{
	Use:  "sync",
	Long: "Sync files from ILIAS WebDAV to local directory",
	RunE: func(cmd *cobra.Command, args []string) error {
		syncerConfig, err := config.ParseFile(configFile)
		if err != nil {
			return fmt.Errorf("failed to parse config file: %v", err)
		}
		if len(syncerConfig.Targets) == 0 {
			return fmt.Errorf("no targets specified in the config file")
		}

		syncer, err := webdav.NewSyncer(syncerConfig.Username, syncerConfig.Password)
		if err != nil {
			return fmt.Errorf("failed to create WebDAV syncer: %v", err)
		}

		for _, target := range syncerConfig.Targets {
			targetPath := filepath.Join(syncerConfig.LocalDirectory, target.Name)
			err = os.MkdirAll(targetPath, os.ModePerm)
			if err != nil {
				return fmt.Errorf("failed to create local directory: %v", err)
			}

			courseURL := getILIASWebDavURL(syncerConfig.ILIASWebDAVBaseURL, target.CourseID)

			err := syncer.Sync(courseURL, targetPath)
			if err != nil {
				return fmt.Errorf("failed to sync course %s: %v", target.CourseID, err)
			}
		}

		return nil
	},
}
