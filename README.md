# PlayIt - CLI Music Player

🎵 A simple fun CLI music player built for the terminal.

## Features

- Add music files, folders, or download from URLs
- Organize songs into playlists
- Play music in background without blocking your terminal
- Support for MP3, WAV, FLAC formats
- Simple commands for managing your music

## Installation

For now, you have to build it from source :)
### Build from Source
```bash
git clone https://github.com/elvonpiko/playit.git
cd playit
go mod tidy
go build -o playit
chmod +x playit  # Linux/macOS only
```

## Quick Start

```bash
# Add music files
./playit add /path/to/music/folder
./playit add https://example.com/song.mp3

# View playlist
./playit playlist

# Start playing
./playit run

# Check status
./playit status

# Stop player
./playit stop
```

## Commands

- `add <file|folder|url>` - Add music files, folders, or download from URL
- `playlist` - View all songs in playlist
- `remove <index|filename>` - Remove songs from playlist
- `run [--shuffle]` - Start music player (add --shuffle for random order)
- `status` - Check if player is running
- `stop` - Stop the player

## Requirements

- Go 1.21+ (for building from source)
- Audio drivers for your OS (ALSA on Linux, Core Audio on macOS)

## License

MIT License
