package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func dmenu(options []string, args string) string {
	cmd := exec.Command("bemenu", "-i", args)
	cmd.Stdin = strings.NewReader(strings.Join(options, "\n"))
	out, err := cmd.Output()
	if err != nil {
		fmt.Println("Error running dmenu:", err)
		os.Exit(1)
	}

	if string(out) == "" {
		os.Exit(1)
	}

	return strings.TrimSpace(string(out))
}

func main() {
	xdgVideosDir := os.Getenv("XDG_VIDEOS_DIR")
	if xdgVideosDir == "" {
		fmt.Println("XDG_VIDEOS_DIR is not set.")
		os.Exit(1)
	}

	path := filepath.Join(xdgVideosDir, "mpvd")
	out := "-o " + path + "/%(title)s.%(ext)s"
	opt := "-ic --embed-chapters"
	err := os.MkdirAll(path, 0755)
	if err != nil {
		fmt.Printf("Failed to create directory: %v\n", err)
		os.Exit(1)
	}

	terminal := os.Getenv("TERMINAL")
	if terminal == "" {
		terminal = "xterm" // fallback terminal if $TERMINAL is not set
	}

	play := func() bool {
		choices := []string{"yes", "no"}
		choice := dmenu(choices, "-p play?")
		return choice == "yes"
	}

	args := os.Args[1:]
	if len(args) == 0 {
		files, err := os.ReadDir(path)
		if err != nil {
			fmt.Printf("Error reading directory: %v\n", err)
			os.Exit(1)
		}

		var videos = []string{"download"}
		for _, file := range files {
			videos = append(videos, file.Name())
		}

		video := dmenu(videos, "-p play")
		if video == "download" {
			if play() {
				cmd := exec.Command(terminal, "-e", "yt-dlp", out, opt, "clipboard-placeholder")
				cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
				cmd.Run()
			}
			os.Exit(0)
		} else {
			cmd := exec.Command("mpv", filepath.Join(path, video))
			cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
			cmd.Run()
		}
	} else {
		if play() {
			cmd := exec.Command(terminal, "-e", "yt-dlp", out, opt, strings.Join(args, " "))
			cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
			cmd.Run()
		}
	}
}
