package sync

import (
	"fmt"
	"net/url"
	"os"
	"path"
	"path/filepath"

	"github.com/dustin/go-humanize"
	"github.com/fatih/color"
	"github.com/matthiasharzer/ilias-sync-webdav/config"
	"github.com/matthiasharzer/ilias-sync-webdav/webdav"
	"github.com/spf13/cobra"
)

type statItem[T webdav.Event] struct {
	Event  T
	Target config.Target
}

func newStatItem[T webdav.Event](event T, target config.Target) statItem[T] {
	return statItem[T]{
		Event:  event,
		Target: target,
	}
}

type syncStats struct {
	Downloaded []statItem[webdav.FileDownloadedEvent]
	Skipped    []statItem[webdav.FileDownloadSkippedEvent]
	Ignored    []statItem[webdav.FileIgnoredEvent]
	Errors     []statItem[webdav.ErrorEvent]
}

func (s *syncStats) Add(other syncStats) {
	s.Downloaded = append(s.Downloaded, other.Downloaded...)
	s.Skipped = append(s.Skipped, other.Skipped...)
	s.Ignored = append(s.Ignored, other.Ignored...)
	s.Errors = append(s.Errors, other.Errors...)
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

		syncer, err := webdav.NewSyncer(syncerConfig.Username, syncerConfig.Password, syncerConfig.AlreadyExistsStrategy)
		if err != nil {
			return fmt.Errorf("failed to create WebDAV syncer: %v", err)
		}

		globalStats := syncStats{}
		for idx, target := range syncerConfig.Targets {
			if idx > 0 {
				fmt.Println("-----")
			}
			targetDirectory := filepath.Join(syncerConfig.LocalDirectory, target.Name)
			fmt.Printf("Syncing course %s (%s) to local directory %s\n", color.BlueString(target.Name), color.BlueString(target.CourseID), color.GreenString(targetDirectory))
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
					size := humanize.Bytes(uint64(e.Stats.Size()))
					fmt.Printf("\tDownloaded %s -> %s (%s)\n", color.BlueString(e.RemoteFile), color.GreenString(e.LocalFile), color.YellowString(size))
					stats.Downloaded = append(stats.Downloaded, newStatItem(e, target))
				case webdav.FileIgnoredEvent:
					fmt.Printf("\tIgnored %s\n", color.YellowString(e.RemoteFile))
					if e.Reason != "" {
						fmt.Printf("\t\tReason: %s\n", color.YellowString(e.Reason))
					}
					stats.Ignored = append(stats.Ignored, newStatItem(e, target))
				case webdav.ErrorEvent:
					fmt.Printf("\tError downloading %s: %v\n", color.RedString(e.RemoteFile), e.Err)
					stats.Errors = append(stats.Errors, newStatItem(e, target))
				case webdav.FileDownloadSkippedEvent:
					if e.Reason != "" {
						fmt.Printf("\tSkipped %s (%s)\n", color.YellowString(e.RemoteFile), color.RGB(100, 100, 100).Sprint(e.Reason))
					} else {
						fmt.Printf("\tSkipped %s\n", color.YellowString(e.RemoteFile))
					}
					stats.Skipped = append(stats.Skipped, newStatItem(e, target))
				}
			}
			globalStats.Add(stats)
		}
		fmt.Printf("-----\nOverall Summary:\n")
		printStats(globalStats)

		if len(globalStats.Ignored) == 0 {
			return nil
		}
		fmt.Println("-----")
		fmt.Printf("List of ignored files:\n")
		for _, ignored := range globalStats.Ignored {
			fullPath := path.Join(ignored.Target.Name, ignored.Event.RemoteFile)
			fmt.Printf("\t%s\n", color.YellowString(fullPath))
		}
		return nil
	},
}

func printStats(stats syncStats) {
	totalSize := int64(0)
	for _, downloaded := range stats.Downloaded {
		totalSize += downloaded.Event.Stats.Size()
	}
	transferredSize := humanize.Bytes(uint64(totalSize))

	fmt.Printf("Downloaded files: %s (%s)\n", color.GreenString(fmt.Sprintf("%d", len(stats.Downloaded))), color.YellowString(transferredSize))
	fmt.Printf("Skipped files: %s\n", color.YellowString(fmt.Sprintf("%d", len(stats.Skipped))))
	fmt.Printf("Ignored files: %s\n", color.YellowString(fmt.Sprintf("%d", len(stats.Ignored))))
	fmt.Printf("Errors: %s\n", color.RedString(fmt.Sprintf("%d", len(stats.Errors))))
}
