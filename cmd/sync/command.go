package sync

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/matthiasharzer/ilias-sync-webdav/config"
	"github.com/matthiasharzer/ilias-sync-webdav/webdav"
	"github.com/spf13/cobra"
)

type syncStats struct {
	Downloaded int
	Ignored    int
	Errors     int
}

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
			fmt.Printf("Syncing course %s to local directory %s\n", color.BlueString(target.CourseID), color.GreenString(target.Name))
			targetPath := filepath.Join(syncerConfig.LocalDirectory, target.Name)
			err = os.MkdirAll(targetPath, os.ModePerm)
			if err != nil {
				return fmt.Errorf("failed to create local directory: %v", err)
			}

			courseURL := getILIASWebDavURL(syncerConfig.ILIASWebDAVBaseURL, target.CourseID)

			events, err := syncer.Sync(courseURL, targetPath)
			if err != nil {
				return fmt.Errorf("failed to sync course %s: %v", target.CourseID, err)
			}

			stats := syncStats{}
			for event := range events {
				switch e := event.(type) {
				case webdav.FileDownloadedEvent:
					fmt.Printf("Downloaded %s -> %s\n", color.BlueString(e.RemoteFile), color.GreenString(e.LocalFile))
					stats.Downloaded++
				case webdav.FileIgnoredEvent:
					if e.Reason != "" {
						fmt.Printf("Ignored %s: %s\n", color.YellowString(e.RemoteFile), e.Reason)
					} else {
						fmt.Printf("Ignored %s\n", color.YellowString(e.RemoteFile))
					}
					stats.Ignored++
				case webdav.ErrorEvent:
					fmt.Printf("Error downloading %s: %v\n", color.RedString(e.RemoteFile), e.Err)
					stats.Errors++
				}
			}

			fmt.Printf("-----\nSummary for course %s:\n", color.BlueString(target.CourseID))
			fmt.Printf("Downloaded: %s\n", color.GreenString(fmt.Sprintf("%d", stats.Downloaded)))
			fmt.Printf("Ignored: %s\n", color.YellowString(fmt.Sprintf("%d", stats.Ignored)))
			fmt.Printf("Errors: %s\n", color.RedString(fmt.Sprintf("%d", stats.Errors)))
		}
		return nil
	},
}
