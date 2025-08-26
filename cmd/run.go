package cmd

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
	"github.com/faiface/beep/wav"
	"github.com/spf13/cobra"
)

var shuffleMode bool

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Play songs from the playlist",
	Long:  "Play songs in playlist order using faiface/beep. Use --shuffle for random order. Runs as background process.",
	Run:   runMusic,
}

var runMusicCmd = &cobra.Command{
	Use:    "run-music",
	Short:  "Internal command to run music player",
	Hidden: true,
	Run:    runMusicPlayer,
}

func init() {
	rootCmd.AddCommand(runCmd)
	runCmd.Flags().BoolVarP(&shuffleMode, "shuffle", "s", false, "Shuffle playlist before playing")

	// Add the internal run-music command
	rootCmd.AddCommand(runMusicCmd)
	runMusicCmd.Flags().BoolVarP(&shuffleMode, "shuffle", "s", false, "Shuffle playlist before playing")
}

func runMusic(cmd *cobra.Command, args []string) {
	// Check if already running
	pidFile := "playit.pid"
	if _, err := os.Stat(pidFile); err == nil {
		fmt.Println("Music player is already running!")
		fmt.Println("Use 'playit status' to check status or 'playit stop' to stop it.")
		return
	}

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

	// Start background process
	fmt.Println("Starting music player in background...")
	fmt.Println("Use 'playit status' to check status or 'playit stop' to stop it.")

	// Fork a new process to run the music player
	cmdArgs := []string{"run-music"}
	if shuffleMode {
		cmdArgs = append(cmdArgs, "--shuffle")
	}

	// Create the command
	runCmd := exec.Command(os.Args[0], cmdArgs...)

	// Set up the process to run in background
	runCmd.Stdout = nil
	runCmd.Stderr = nil
	runCmd.Stdin = nil

	// Start the detached process
	if err := runCmd.Start(); err != nil {
		fmt.Printf("Error starting music player: %v\n", err)
		return
	}

	// Write PID file with the child process PID
	childPID := runCmd.Process.Pid
	if err := os.WriteFile(pidFile, []byte(fmt.Sprintf("%d", childPID)), 0644); err != nil {
		fmt.Printf("Warning: Could not write PID file: %v\n", err)
	}

	fmt.Printf("Music player started with PID: %d\n", childPID)
	fmt.Println("Use 'playit status' to check status or 'playit stop' to stop it.")
}

func runMusicPlayer(cmd *cobra.Command, args []string) {
	// This is the internal command that actually runs the music player
	// It's called by the main run command after forking

	// Check if playlist exists
	if _, err := os.Stat("playlist.json"); os.IsNotExist(err) {
		fmt.Println("No playlist found.")
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
		fmt.Println("Playlist is empty.")
		return
	}

	// Set up signal handling for graceful shutdown
	pidFile := "playit.pid"

	// Clean up PID file when done
	defer func() {
		os.Remove(pidFile)
		fmt.Println("Music player stopped.")
	}()

	// Play the playlist
	playPlaylist(playlist, shuffleMode)
}

func playPlaylist(playlist Playlist, shuffle bool) {
	// Shuffle playlist if shuffle mode is enabled
	if shuffle {
		rand.Shuffle(len(playlist.Files), func(i, j int) {
			playlist.Files[i], playlist.Files[j] = playlist.Files[j], playlist.Files[i]
		})
	}

	// Initialize speaker
	sr := beep.SampleRate(44100)
	speaker.Init(sr, sr.N(time.Second/10))

	// Play each song
	for _, fileName := range playlist.Files {
		filePath := filepath.Join("music", fileName)

		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			continue
		}

		if err := playFile(filePath); err != nil {
			continue
		}
	}
}

func playFile(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	var streamer beep.StreamSeekCloser
	var format beep.Format

	ext := filepath.Ext(filePath)
	switch ext {
	case ".mp3":
		streamer, format, err = mp3.Decode(file)
	case ".wav":
		streamer, format, err = wav.Decode(file)
	case ".flac":
		// Note: flac support would require additional library
		return fmt.Errorf("FLAC format not yet supported")
	default:
		return fmt.Errorf("unsupported format: %s", ext)
	}

	if err != nil {
		return err
	}
	defer streamer.Close()

	// Resample if necessary
	if format.SampleRate != beep.SampleRate(44100) {
		resampled := beep.Resample(4, format.SampleRate, beep.SampleRate(44100), streamer)
		// Play the audio
		done := make(chan bool)
		speaker.Play(beep.Seq(resampled, beep.Callback(func() {
			done <- true
		})))
		// Wait for the song to finish
		<-done
	} else {
		// Play the audio
		done := make(chan bool)
		speaker.Play(beep.Seq(streamer, beep.Callback(func() {
			done <- true
		})))
		// Wait for the song to finish
		<-done
	}

	return nil
}
