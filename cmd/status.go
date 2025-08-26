package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show the status of the music player",
	Long:  "Check if the music player is currently running or stopped",
	Run:   showStatus,
}

func showStatus(cmd *cobra.Command, args []string) {
	pidFile := "playit.pid"

	if _, err := os.Stat(pidFile); os.IsNotExist(err) {
		fmt.Println("Status: STOPPED")
		fmt.Println("Playit is not currently running.")
		fmt.Println("Use 'playit run' to start playing music.")
		return
	}

	// Read PID file
	pidData, err := os.ReadFile(pidFile)
	if err != nil {
		fmt.Printf("Error reading PID file: %v\n", err)
		return
	}

	pid := string(pidData)
	fmt.Printf("Status: RUNNING (PID: %s)\n", pid)
	fmt.Println("Playit is currently active.")
	fmt.Println("Use 'playit stop' to stop the player.")
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
