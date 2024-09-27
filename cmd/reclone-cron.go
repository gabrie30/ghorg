package cmd

import (
	_ "embed"
	"log"
	"os"
	"os/exec"
	"strconv"
	"time"

	"github.com/gabrie30/ghorg/colorlog"
	"github.com/spf13/cobra"
)

var recloneCronCmd = &cobra.Command{
	Use:   "reclone-cron",
	Short: "Simple cron that will trigger your reclone command at a specified minute intervals indefinitely",
	Long:  `Read the documentation and examples in the readme`,
	Run: func(cmd *cobra.Command, args []string) {
		if cmd.Flags().Changed("minutes") {
			os.Setenv("GHORG_CRON_TIMER_MINUTES", cmd.Flag("minutes").Value.String())
		}

		startReCloneCron()
	},
}

func startReCloneCron() {
	if os.Getenv("GHORG_CRON_TIMER_MINUTES") == "" {
		return
	}
	colorlog.PrintInfo("Cron activated and will first run after " + os.Getenv("GHORG_CRON_TIMER_MINUTES") + " minutes ")

	minutes, err := strconv.Atoi(os.Getenv("GHORG_CRON_TIMER_MINUTES"))
	if err != nil {
		log.Fatalf("Invalid GHORG_CRON_TIMER_MINUTES: %v", err)
	}

	ticker := time.NewTicker(time.Duration(minutes) * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		colorlog.PrintInfo("starting reclone cron, time: " + time.Now().Format(time.RFC1123))
		cmd := exec.Command("ghorg", "reclone")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			log.Printf("Failed to run ghorg reclone: %v", err)
		}
	}
}
