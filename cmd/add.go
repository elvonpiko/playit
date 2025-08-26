package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add [file|folder|url]",
	Short: "Add music files or download from URL to the playlist",
	Long:  "Add single music files, entire folders, or download from URL to the music/ directory and update playlist.json",
	Args:  cobra.ExactArgs(1),
	Run:   addMusic,
}

func addMusic(cmd *cobra.Command, args []string) {
	path := args[0]

	// Check if input is a URL
	if isURL(path) {
		if err := addFromURL(path); err != nil {
			fmt.Printf("Error downloading from URL: %v\n", err)
			return
		}
		fmt.Printf("Successfully downloaded and added music file from URL to playlist\n")
		return
	}

	// Create music directory if it doesn't exist
	if err := os.MkdirAll("music", 0755); err != nil {
		fmt.Printf("Error creating music directory: %v\n", err)
		return
	}

	// Check if path exists
	info, err := os.Stat(path)
	if err != nil {
		fmt.Printf("Error accessing path: %v\n", err)
		return
	}

	var addedFiles []string

	if info.IsDir() {
		// Add entire folder
		err = filepath.Walk(path, func(filePath string, info fs.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() && isMusicFile(filePath) {
				if err := addSingleFile(filePath); err != nil {
					fmt.Printf("Error adding file %s: %v\n", filePath, err)
				} else {
					addedFiles = append(addedFiles, filepath.Base(filePath))
				}
			}
			return nil
		})
		if err != nil {
			fmt.Printf("Error walking directory: %v\n", err)
			return
		}
	} else {
		// Add single file
		if isMusicFile(path) {
			if err := addSingleFile(path); err != nil {
				fmt.Printf("Error adding file: %v\n", err)
				return
			}
			addedFiles = append(addedFiles, filepath.Base(path))
		} else {
			fmt.Printf("File %s is not a supported music format\n", path)
			return
		}
	}

	// Update playlist
	if err := updatePlaylist(addedFiles); err != nil {
		fmt.Printf("Error updating playlist: %v\n", err)
		return
	}

	fmt.Printf("Successfully added %d music file(s) to playlist\n", len(addedFiles))
}

func isMusicFile(filePath string) bool {
	ext := strings.ToLower(filepath.Ext(filePath))
	return ext == ".mp3" || ext == ".wav" || ext == ".flac"
}

func isURL(path string) bool {
	_, err := url.ParseRequestURI(path)
	return err == nil && (strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://"))
}

func addFromURL(urlStr string) error {
	// Parse URL
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return fmt.Errorf("invalid URL: %v", err)
	}

	// Extract filename from URL
	fileName := filepath.Base(parsedURL.Path)
	if fileName == "" || fileName == "." {
		fileName = "downloaded_music.mp3"
	}

	// Check if it's a supported music format
	if !isMusicFile(fileName) {
		return fmt.Errorf("URL does not point to a supported music format (.mp3, .wav, .flac)")
	}

	// Create music directory if it doesn't exist
	if err := os.MkdirAll("music", 0755); err != nil {
		return fmt.Errorf("failed to create music directory: %v", err)
	}

	// Download the file with progress tracking
	resp, err := http.Get(urlStr)
	if err != nil {
		return fmt.Errorf("failed to download file: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP error: %d", resp.StatusCode)
	}

	// Get content length for progress tracking
	contentLength := resp.ContentLength
	if contentLength > 0 {
		fmt.Printf("Downloading %s (%s)...\n", fileName, formatBytes(contentLength))
	} else {
		fmt.Printf("Downloading %s...\n", fileName)
	}

	// Create destination file
	destPath := filepath.Join("music", fileName)
	dest, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %v", err)
	}
	defer dest.Close()

	// Copy with progress tracking
	_, err = copyWithProgress(dest, resp.Body, contentLength, fileName)
	if err != nil {
		return fmt.Errorf("failed to save downloaded file: %v", err)
	}

	fmt.Printf("\nDownload complete: %s\n", fileName)

	// Add to playlist
	return updatePlaylist([]string{fileName})
}

func addSingleFile(filePath string) error {
	fileName := filepath.Base(filePath)
	destPath := filepath.Join("music", fileName)

	// Copy file to music directory
	src, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer src.Close()

	dest, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer dest.Close()

	// Stream copy to avoid loading entire file into memory
	_, err = io.Copy(dest, src)
	return err
}

func updatePlaylist(newFiles []string) error {
	playlistPath := "playlist.json"

	var pl Playlist
	if data, err := os.ReadFile(playlistPath); err == nil && len(data) > 0 {
		_ = json.Unmarshal(data, &pl)
	}

	// Append with simple dedupe
	existing := make(map[string]bool, len(pl.Files))
	for _, f := range pl.Files {
		existing[f] = true
	}
	for _, f := range newFiles {
		if !existing[f] {
			pl.Files = append(pl.Files, f)
			existing[f] = true
		}
	}

	data, err := json.MarshalIndent(pl, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(playlistPath, data, 0644)
}

// copyWithProgress copies data with a progress bar
func copyWithProgress(dst io.Writer, src io.Reader, total int64, filename string) (int64, error) {
	// 32KB buffer
	buf := make([]byte, 32*1024)
	var written int64
	var lastPercent int

	for {
		nr, er := src.Read(buf)
		if nr > 0 {
			nw, ew := dst.Write(buf[0:nr])
			if nw > 0 {
				written += int64(nw)
			}
			if ew != nil {
				return written, ew
			}
			if nr != nw {
				return written, io.ErrShortWrite
			}
		}
		if er != nil {
			if er == io.EOF {
				break
			}
			return written, er
		}

		// Update progress bar
		if total > 0 {
			percent := int((written * 100) / total)
			if percent > lastPercent {
				lastPercent = percent
				updateProgressBar(percent, filename)
			}
		}
	}

	return written, nil
}

// updateProgressBar displays a progress bar
func updateProgressBar(percent int, filename string) {
	const barWidth = 50
	filled := (barWidth * percent) / 100
	empty := barWidth - filled

	bar := "["
	for i := 0; i < filled; i++ {
		bar += "="
	}
	for i := 0; i < empty; i++ {
		bar += " "
	}
	bar += "]"

	fmt.Printf("\r%s %s %d%%", filename, bar, percent)
}

func init() {
	rootCmd.AddCommand(addCmd)
}
