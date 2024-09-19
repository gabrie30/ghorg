package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
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

var spinningSpinner *spinner.Spinner

func init() {
	spinningSpinner = spinner.New(spinner.CharSets[14], 100*time.Millisecond)
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

	spinningSpinner.Start()

	var totalDirs int
	var totalSizeMB float64
	var totalRepos int

	type result struct {
		dirPath     string
		dirSizeMB   float64
		subDirCount int
		err         error
	}

	results := make(chan result, len(files))
	var wg sync.WaitGroup

	for _, f := range files {
		if f.IsDir() {
			totalDirs++
			wg.Add(1)
			go func(f os.DirEntry) {
				defer wg.Done()
				dirPath := filepath.Join(path, f.Name())
				dirSizeMB, err := utils.CalculateDirSizeInMb(dirPath)
				if err != nil {
					results <- result{dirPath: dirPath, err: err}
					return
				}

				subDirCount := 0
				subFiles, err := os.ReadDir(dirPath)
				if err != nil {
					results <- result{dirPath: dirPath, err: err}
					return
				}
				for _, subFile := range subFiles {
					if subFile.IsDir() {
						subDirCount++
					}
				}
				results <- result{dirPath: dirPath, dirSizeMB: dirSizeMB, subDirCount: subDirCount}
			}(f)
		}
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	for res := range results {
		if res.err != nil {
			colorlog.PrintError(fmt.Sprintf("Error processing directory %s: %v", res.dirPath, res.err))
			continue
		}
		totalSizeMB += res.dirSizeMB
		totalRepos += res.subDirCount
		if !totalFormat || longFormat {
			spinningSpinner.Stop()
			if longFormat {
				if res.dirSizeMB > 1000 {
					dirSizeGB := res.dirSizeMB / 1000
					colorlog.PrintInfo(fmt.Sprintf("%-90s %10.2f GB %10d repos", res.dirPath, dirSizeGB, res.subDirCount))
				} else {
					colorlog.PrintInfo(fmt.Sprintf("%-90s %10.2f MB %10d repos", res.dirPath, res.dirSizeMB, res.subDirCount))
				}
			} else {
				colorlog.PrintInfo(res.dirPath)
			}
		}
	}

	spinningSpinner.Stop()
	if totalFormat {
		if totalSizeMB > 1000 {
			totalSizeGB := totalSizeMB / 1000
			colorlog.PrintSuccess(fmt.Sprintf("Total: %d directories, %.2f GB, %d repos", totalDirs, totalSizeGB, totalRepos))
		} else {
			colorlog.PrintSuccess(fmt.Sprintf("Total: %d directories, %.2f MB, %d repos", totalDirs, totalSizeMB, totalRepos))
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
				str := filepath.Join(path, f.Name())
				colorlog.PrintInfo(str)
			}
		}
		return
	}

	spinningSpinner.Start()

	var totalDirs int
	var totalSizeMB float64

	type result struct {
		dirPath     string
		dirSizeMB   float64
		subDirCount int
		err         error
	}

	results := make(chan result)
	var wg sync.WaitGroup

	for _, f := range files {
		if f.IsDir() {
			wg.Add(1)
			go func(f os.DirEntry) {
				defer wg.Done()
				dirPath := filepath.Join(path, f.Name())
				dirSizeMB, err := utils.CalculateDirSizeInMb(dirPath)
				if err != nil {
					results <- result{dirPath: dirPath, err: err}
					return
				}

				// Count the number of directories with a depth of 1 inside
				subDirCount := 0
				subFiles, err := os.ReadDir(dirPath)
				if err != nil {
					results <- result{dirPath: dirPath, err: err}
					return
				}
				for _, subFile := range subFiles {
					if subFile.IsDir() {
						subDirCount++
					}
				}
				results <- result{dirPath: dirPath, dirSizeMB: dirSizeMB, subDirCount: subDirCount}
			}(f)
		}
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	for res := range results {
		if res.err != nil {
			colorlog.PrintError(fmt.Sprintf("Error processing directory %s: %v", res.dirPath, res.err))
			continue
		}
		totalSizeMB += res.dirSizeMB
		totalDirs++
		if !totalFormat || longFormat {
			spinningSpinner.Stop()
			if longFormat {
				if res.dirSizeMB > 1000 {
					dirSizeGB := res.dirSizeMB / 1000
					colorlog.PrintInfo(fmt.Sprintf("%-90s %10.2f GB ", res.dirPath, dirSizeGB))
				} else {
					colorlog.PrintInfo(fmt.Sprintf("%-90s %10.2f MB", res.dirPath, res.dirSizeMB))
				}
			} else {
				colorlog.PrintInfo(res.dirPath)
			}
		}
	}

	spinningSpinner.Stop()
	if totalFormat {
		if totalSizeMB > 1000 {
			totalSizeGB := totalSizeMB / 1000
			colorlog.PrintSuccess(fmt.Sprintf("Total: %d repos, %.2f GB", totalDirs, totalSizeGB))
		} else {
			colorlog.PrintSuccess(fmt.Sprintf("Total: %d repos, %.2f MB", totalDirs, totalSizeMB))
		}
	}
}
