package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"
)

var removeCmd = &cobra.Command{
	Use:   "remove [index|filename]",
	Short: "Remove a song from the playlist",
	Long:  "Remove a song from the playlist by specifying its index number or filename",
	Args:  cobra.ExactArgs(1),
	Run:   removeSong,
}

func removeSong(cmd *cobra.Command, args []string) {
	target := args[0]

	// Check if playlist exists
	if _, err := os.Stat("playlist.json"); os.IsNotExist(err) {
		fmt.Println("No playlist found. Use 'playit add' to add music files first.")
		return
	}

	// Read playlist
	data, err := os.ReadFile("playlist.json")
	if err != nil {
		fmt.Printf("Error reading playlist: %v\n", err)
		return
	}

	var playlist Playlist
	if err := json.Unmarshal(data, &playlist); err != nil {
		fmt.Printf("Error parsing playlist: %v\n", err)
		return
	}

	if len(playlist.Files) == 0 {
		fmt.Println("Playlist is empty. Use 'playit add' to add music files.")
		return
	}

	// Try to parse as index first
	if index, err := strconv.Atoi(target); err == nil {
		// Remove by index (1-based for user convenience)
		if index < 1 || index > len(playlist.Files) {
			fmt.Printf("Invalid index. Please use a number between 1 and %d.\n", len(playlist.Files))
			return
		}
		
		removedFile := playlist.Files[index-1]
		playlist.Files = append(playlist.Files[:index-1], playlist.Files[index:]...)
		
		fmt.Printf("Removed song %d: %s\n", index, removedFile)
	} else {
		// Remove by filename
		found := false
		for i, fileName := range playlist.Files {
			if fileName == target {
				playlist.Files = append(playlist.Files[:i], playlist.Files[i+1:]...)
				fmt.Printf("Removed song: %s\n", fileName)
				found = true
				break
			}
		}
		
		if !found {
			fmt.Printf("Song '%s' not found in playlist.\n", target)
			fmt.Println("Use 'playit playlist' to see available songs.")
			return
		}
	}

	// Save updated playlist
	data, err = json.MarshalIndent(playlist, "", "  ")
	if err != nil {
		fmt.Printf("Error saving playlist: %v\n", err)
		return
	}

	if err := os.WriteFile("playlist.json", data, 0644); err != nil {
		fmt.Printf("Error writing playlist: %v\n", err)
		return
	}

	fmt.Printf("Playlist updated. %d songs remaining.\n", len(playlist.Files))
}

func init() {
	rootCmd.AddCommand(removeCmd)
}
