package cmd

import (
	_ "embed"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"sync"

	"github.com/gabrie30/ghorg/colorlog"
	"github.com/spf13/cobra"
)

var recloneServerCmd = &cobra.Command{
	Use:   "reclone-server",
	Short: "Server allowing you to trigger adhoc or cron based reclone commands",
	Long:  `Read the documentation and examples in the readme`,
	Run: func(cmd *cobra.Command, args []string) {
		if cmd.Flags().Changed("port") {
			os.Setenv("GHORG_RECLONE_SERVER_PORT", cmd.Flag("port").Value.String())
		}

		startReCloneServer()
	},
}

func startReCloneServer() {
	var mu sync.Mutex
	serverPort := os.Getenv("GHORG_RECLONE_SERVER_PORT")
	if serverPort != "" && serverPort[0] != ':' {
		serverPort = ":" + serverPort
	}

	http.HandleFunc("/trigger/reclone", func(w http.ResponseWriter, r *http.Request) {
		userCmd := r.URL.Query().Get("cmd")

		if !mu.TryLock() {
			http.Error(w, "Server is busy, please try again later", http.StatusTooManyRequests)
			return
		}

		// Signal channel to notify when the command has started
		started := make(chan struct{})

		go func() {
			defer mu.Unlock()
			var cmd *exec.Cmd
			if userCmd == "" {
				cmd = exec.Command("ghorg", "reclone")
			} else {
				cmd = exec.Command("ghorg", "reclone", userCmd)
			}
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr

			// Notify that the command has started
			close(started)

			if err := cmd.Run(); err != nil {
				fmt.Printf("Error running command: %s\n", err)
			}
		}()

		// Wait for the command to start before responding
		<-started
		w.WriteHeader(http.StatusOK)
	})

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	colorlog.PrintInfo("Starting reclone server on " + serverPort)
	if err := http.ListenAndServe(serverPort, nil); err != nil {
		fmt.Printf("Error starting server: %s\n", err)
	}
}
