// Package colorlog has various Print functions that can be called to change the color of the text in standard out
package colorlog

import (
	"fmt"
	"os"

	"github.com/fatih/color"
)

// PrintInfo prints yellow colored text to standard out
func PrintInfo(msg any) {
	if os.Getenv("GHORG_QUIET") == "true" {
		return
	}

	switch os.Getenv("GHORG_COLOR") {
	case "enabled":
		color.New(color.FgYellow).Println(msg)
	default:
		fmt.Println(msg)
	}
}

// PrintSuccess prints green colored text to standard out
func PrintSuccess(msg any) {
	switch os.Getenv("GHORG_COLOR") {
	case "enabled":
		color.New(color.FgGreen).Println(msg)
	default:
		fmt.Println(msg)
	}
}

// PrintError prints red colored text to standard out
func PrintError(msg any) {
	switch os.Getenv("GHORG_COLOR") {
	case "enabled":
		color.New(color.FgRed).Println(msg)
	default:
		fmt.Println(msg)
	}
}

// PrintErrorAndExit prints red colored text to standard out then exits 1
func PrintErrorAndExit(msg any) {
	switch os.Getenv("GHORG_COLOR") {
	case "enabled":
		color.New(color.FgRed).Println(msg)
	default:
		fmt.Println(msg)
	}

	os.Exit(1)
}

// PrintSubtleInfo prints magenta colored text to standard out
func PrintSubtleInfo(msg any) {
	if os.Getenv("GHORG_QUIET") == "true" {
		return
	}

	switch os.Getenv("GHORG_COLOR") {
	case "enabled":
		color.New(color.FgHiMagenta).Println(msg)
	default:
		fmt.Println(msg)
	}
}
