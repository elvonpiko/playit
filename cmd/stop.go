package cmd

import (
	"fmt"
	"os"
	"strconv"
	"syscall"

	"github.com/spf13/cobra"
)

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the running music player",
	Long:  "Stop the currently running music player process",
	Run:   stopPlayer,
}

func stopPlayer(cmd *cobra.Command, args []string) {
	pidFile := "playit.pid"
	
	if _, err := os.Stat(pidFile); os.IsNotExist(err) {
		fmt.Println("No music player is currently running.")
		return
	}

	// Read PID file
	pidData, err := os.ReadFile(pidFile)
	if err != nil {
		fmt.Printf("Error reading PID file: %v\n", err)
		return
	}

	pid, err := strconv.Atoi(string(pidData))
	if err != nil {
		fmt.Printf("Error parsing PID: %v\n", err)
		return
	}

	// Try to stop the process gracefully first
	process, err := os.FindProcess(pid)
	if err != nil {
		fmt.Printf("Error finding process: %v\n", err)
		return
	}

	// Send SIGTERM first (graceful shutdown)
	err = process.Signal(syscall.SIGTERM)
	if err != nil {
		fmt.Printf("Error sending SIGTERM: %v\n", err)
		return
	}

	fmt.Printf("Stopping music player (PID: %d)...\n", pid)
	fmt.Println("The player will stop after the current song finishes.")

	// Remove PID file
	if err := os.Remove(pidFile); err != nil {
		fmt.Printf("Warning: Could not remove PID file: %v\n", err)
	}

	fmt.Println("Music player stopped successfully.")
}

func init() {
	rootCmd.AddCommand(stopCmd)
}
