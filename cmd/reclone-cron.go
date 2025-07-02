package cmd

import (
	_ "embed"
	"os"
	"os/exec"
	"strconv"
	"sync"
	"time"

	"github.com/gabrie30/ghorg/colorlog"
	"github.com/spf13/cobra"
)

var (
	recloneRunning bool
	recloneMutex   sync.Mutex
)

var recloneCronCmd = &cobra.Command{
	Use:   "reclone-cron",
	Short: "Simple cron that will trigger your reclone command at a specified minute intervals indefinitely",
	Long:  `Read the documentation and examples in the Readme under Reclone Server heading`,
	Run: func(cmd *cobra.Command, args []string) {
		if cmd.Flags().Changed("minutes") {
			os.Setenv("GHORG_CRON_TIMER_MINUTES", cmd.Flag("minutes").Value.String())
		}

		startReCloneCron()
	},
}

func startReCloneCron() {
	cronTimer := os.Getenv("GHORG_CRON_TIMER_MINUTES")
	if cronTimer == "" {
		colorlog.PrintInfo("GHORG_CRON_TIMER_MINUTES is not set. Cron job will not start.")
		return
	}

	colorlog.PrintInfo("Cron activated and will first run after " + cronTimer + " minutes ")

	minutes, err := strconv.Atoi(cronTimer)
	if err != nil {
		colorlog.PrintError("Invalid GHORG_CRON_TIMER_MINUTES: " + cronTimer)
		return
	}

	ticker := time.NewTicker(time.Duration(minutes) * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		recloneMutex.Lock()
		if recloneRunning {
			recloneMutex.Unlock()
			continue
		}
		recloneRunning = true
		recloneMutex.Unlock()

		colorlog.PrintInfo("Starting reclone cron, time: " + time.Now().Format(time.RFC1123))
		cmd := exec.Command("ghorg", "reclone")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Start(); err != nil {
			colorlog.PrintError("Failed to start ghorg reclone: " + err.Error())
			recloneMutex.Lock()
			recloneRunning = false
			recloneMutex.Unlock()
			continue
		}

		go func() {
			if err := cmd.Wait(); err != nil {
				colorlog.PrintError("ghorg reclone command failed: " + err.Error())
			}
			recloneMutex.Lock()
			recloneRunning = false
			recloneMutex.Unlock()
		}()
	}
}
