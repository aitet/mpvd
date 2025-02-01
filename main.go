package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"syscall"
)

var path string

func run(bin string, args ...[]string) {
	var a []string
	for _, s := range args {
		a = append(a, s...)
	}
	cmd := exec.Command(bin, a...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	cmd.Dir = path

	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to run command '%s %v': %v\n", bin, args, err)
		os.Exit(1)
	}
}

func dmenu(options []string, args ...[]string) string {
	var a = []string{"-i"}
	for _, s := range args {
		a = append(a, s...)
	}
	cmd := exec.Command("bemenu", a...)
	cmd.Stdin = strings.NewReader(strings.Join(options, "\n"))
	out, err := cmd.Output()
	if err != nil || string(out) == "" {
		os.Exit(1)
	}

	return strings.TrimSpace(string(out))
}

func main() {
	var args = []string{"-ic", "--embed-chapters"}

	xdgVideosDir := os.Getenv("XDG_VIDEOS_DIR")
	if xdgVideosDir == "" {
		fmt.Println("XDG_VIDEOS_DIR is not set.")
		os.Exit(1)
	}

	path = filepath.Join(xdgVideosDir, "mpvd")
	err := os.MkdirAll(path, 0755)
	if err != nil {
		fmt.Printf("Failed to create directory: %v\n", err)
		os.Exit(1)
	}

	if len(os.Args[1:]) == 0 {
		files, err := os.ReadDir(path)
		if err != nil {
			fmt.Printf("Error reading directory: %v\n", err)
			os.Exit(1)
		}

		// Sort by birth time (newest first)
		sort.Slice(files, func(i, j int) bool {
			fi, _ := files[i].Info()
			fj, _ := files[j].Info()
			si := fi.Sys().(*syscall.Stat_t).Ctim
			sj := fj.Sys().(*syscall.Stat_t).Ctim
			// Compare seconds (and if needed, nanoseconds)
			if si.Sec == sj.Sec {
				return si.Nsec > sj.Nsec
			}
			return si.Sec > sj.Sec
		})

		var videos = []string{"download"}
		for _, file := range files {
			videos = append(videos, file.Name())
		}
		videos = append(videos, "delete all?")

		switch choice := dmenu(videos, []string{"-p", "play", "-W", "0.8"}); choice {
		case "download":
			if os.Getenv("BEMENU_BACKEND") == "curses" {
				if dmenu([]string{"yes", "no"}, []string{"-p", "play?"}) == "yes" {
					args = append(args, "--exec=mpv")
				}
			} else {
				args = append(args, "--exec=mpv")
			}
			cmd := exec.Command("clipboard", "-o")

			clipboard, err := cmd.Output()
			if err != nil {
				fmt.Println("Error getting clipboard: ", err)
				os.Exit(1)
			}

			args = append(args, string(clipboard))
			run("yt-dlp", args)
			os.Exit(0)
		case "delete all?":
			if dmenu([]string{"yes", "no"}, []string{"-p", "you sure?"}) == "yes" {
				os.RemoveAll(path)
				os.Exit(0)
			}
		default:
			run("mpv", []string{choice})
			os.Exit(0)
		}

	} else {
		if dmenu([]string{"yes", "no"}, []string{"-p", "play?"}) == "yes" {
			args = append(args, "--exec=mpv")
		}
		args = append(args, strings.Join(os.Args[1:], " "))
		run("yt-dlp", args)
		os.Exit(0)
	}
}
