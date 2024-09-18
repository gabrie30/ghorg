package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	"github.com/gabrie30/ghorg/colorlog"
	"github.com/gabrie30/ghorg/utils"
	"github.com/spf13/cobra"
)

var lsCmd = &cobra.Command{
	Use:   "ls [dir]",
	Short: "List contents of your ghorg home or ghorg directories",
	Long:  `If no dir is specified it will list contents of GHORG_ABSOLUTE_PATH_TO_CLONE_TO`,
	Run:   lsFunc,
}

func lsFunc(cmd *cobra.Command, argz []string) {
	if len(argz) == 0 {
		listGhorgHome()
	}

	if len(argz) >= 1 {
		for _, arg := range argz {
			listGhorgDir(arg)
		}
	}

}

func listGhorgHome() {
	path := os.Getenv("GHORG_ABSOLUTE_PATH_TO_CLONE_TO")
	files, err := os.ReadDir(path)
	if err != nil {
		colorlog.PrintError("No clones found. Please clone some and try again.")
	}

	longFormat := false
	totalFormat := false
	for _, arg := range os.Args {
		if arg == "-l" || arg == "--long" {
			longFormat = true
		}
		if arg == "-t" || arg == "--total" {
			totalFormat = true
		}
	}

	if !longFormat && !totalFormat {
		for _, f := range files {
			if f.IsDir() {
				colorlog.PrintInfo(path + f.Name())
			}
		}
		return
	}

	spinningSpinner := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	spinningSpinner.Start()

	var totalDirs int
	var totalSizeMB float64
	var totalRepos int

	for _, f := range files {
		if f.IsDir() {
			totalDirs++
			dirPath := filepath.Join(path, f.Name())
			dirSizeMB, err := utils.CalculateDirSizeInMb(dirPath)
			if err != nil {
				colorlog.PrintError(fmt.Sprintf("Error calculating directory size for %s: %v", dirPath, err))
				continue
			}
			totalSizeMB += dirSizeMB

			// Count the number of directories with a depth of 1 inside
			subDirCount := 0
			subFiles, err := os.ReadDir(dirPath)
			if err != nil {
				colorlog.PrintError(fmt.Sprintf("Error reading directory contents for %s: %v", dirPath, err))
				continue
			}
			for _, subFile := range subFiles {
				if subFile.IsDir() {
					subDirCount++
				}
			}
			totalRepos += subDirCount
			if !totalFormat || longFormat {
				spinningSpinner.Stop()
				if longFormat {
					if dirSizeMB > 1000 {
						dirSizeGB := dirSizeMB / 1000
						colorlog.PrintInfo(fmt.Sprintf("%-60s %10.2f GB %10d repos", dirPath, dirSizeGB, subDirCount))
					} else {
						colorlog.PrintInfo(fmt.Sprintf("%-60s %10.2f MB %10d repos", dirPath, dirSizeMB, subDirCount))
					}
				} else {
					colorlog.PrintInfo(path + f.Name())
				}
			}
		}
	}

	spinningSpinner.Stop()
	if totalFormat {
		if totalSizeMB > 1000 {
			totalSizeGB := totalSizeMB / 1000
			colorlog.PrintInfo(fmt.Sprintf("Total: %d directories, %.2f GB, %d repos", totalDirs, totalSizeGB, totalRepos))
		} else {
			colorlog.PrintInfo(fmt.Sprintf("Total: %d directories, %.2f MB, %d repos", totalDirs, totalSizeMB, totalRepos))
		}
	}
}

func listGhorgDir(arg string) {

	path := os.Getenv("GHORG_ABSOLUTE_PATH_TO_CLONE_TO") + arg

	_, err := os.ReadDir(path)
	if err != nil {
		// ghorg natively uses underscores in folder names, but a user can specify an output dir with underscores
		// so first try what the user types if not then try replace
		arg = strings.ReplaceAll(arg, "-", "_")
		path = os.Getenv("GHORG_ABSOLUTE_PATH_TO_CLONE_TO") + arg
	}

	files, err := os.ReadDir(path)
	if err != nil {
		colorlog.PrintError("No clones found. Please clone some and try again.")
	}

	longFormat := false
	totalFormat := false
	for _, arg := range os.Args {
		if arg == "-l" || arg == "--long" {
			longFormat = true
		}
		if arg == "-t" || arg == "--total" {
			totalFormat = true
		}
	}

	if !longFormat && !totalFormat {
		for _, f := range files {
			if f.IsDir() {
				colorlog.PrintInfo(path + f.Name())
			}
		}
		return
	}

	spinningSpinner := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	spinningSpinner.Start()

	var totalDirs int
	var totalSizeMB float64

	for _, f := range files {
		if f.IsDir() {
			totalDirs++
			dirPath := filepath.Join(path, f.Name())
			dirSizeMB, err := utils.CalculateDirSizeInMb(dirPath)
			if err != nil {
				colorlog.PrintError(fmt.Sprintf("Error calculating directory size for %s: %v", dirPath, err))
				continue
			}
			totalSizeMB += dirSizeMB

			// Count the number of directories with a depth of 1 inside
			subDirCount := 0
			subFiles, err := os.ReadDir(dirPath)
			if err != nil {
				colorlog.PrintError(fmt.Sprintf("Error reading directory contents for %s: %v", dirPath, err))
				continue
			}
			for _, subFile := range subFiles {
				if subFile.IsDir() {
					subDirCount++
				}
			}
			if !totalFormat || longFormat {
				spinningSpinner.Stop()
				if longFormat {
					if dirSizeMB > 1000 {
						dirSizeGB := dirSizeMB / 1000
						colorlog.PrintInfo(fmt.Sprintf("%-80s %10.2f GB ", dirPath, dirSizeGB))
					} else {
						colorlog.PrintInfo(fmt.Sprintf("%-80s %10.2f MB", dirPath, dirSizeMB))
					}
				} else {
					colorlog.PrintInfo(path + f.Name())
				}
			}
		}
	}

	spinningSpinner.Stop()
	if totalFormat {
		if totalSizeMB > 1000 {
			totalSizeGB := totalSizeMB / 1000
			colorlog.PrintInfo(fmt.Sprintf("Total: %d repos, %.2f GB", totalDirs, totalSizeGB))
		} else {
			colorlog.PrintInfo(fmt.Sprintf("Total: %d repos, %.2f MB", totalDirs, totalSizeMB))
		}
	}
}
