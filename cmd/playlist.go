package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var playlistCmd = &cobra.Command{
	Use:   "playlist",
	Short: "Show the current playlist",
	Long:  "Display all songs in the current playlist with file information",
	Run:   showPlaylist,
}

func showPlaylist(cmd *cobra.Command, args []string) {
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

	fmt.Printf("Playlist (%d songs):\n", len(playlist.Files))
	fmt.Println(strings.Repeat("-", 80))

	for i, fileName := range playlist.Files {
		filePath := filepath.Join("music", fileName)
		
		// Get file info
		info, err := os.Stat(filePath)
		if err != nil {
			fmt.Printf("%3d. %s [File not found]\n", i+1, fileName)
			continue
		}

		// Format file size
		size := formatBytes(info.Size())
		
		// Get file extension
		ext := strings.ToUpper(strings.TrimPrefix(filepath.Ext(fileName), "."))
		
		fmt.Printf("%3d. %s (%s, %s)\n", i+1, fileName, ext, size)
	}
	
	fmt.Println(strings.Repeat("-", 80))
}

func init() {
	rootCmd.AddCommand(playlistCmd)
}
