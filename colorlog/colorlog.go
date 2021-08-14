// Package colorlog has various Print functions that can be called to change the color of the text in standard out
package colorlog

import (
	"fmt"
	"os"

	"github.com/fatih/color"
)

// PrintInfo prints yellow colored text to standard out
func PrintInfo(msg interface{}) {
	switch os.Getenv("GHORG_COLOR") {
	case "enabled":
		color.New(color.FgYellow).Println(msg)
	default:
		fmt.Println(msg)
	}
}

// PrintSuccess prints green colored text to standard out
func PrintSuccess(msg interface{}) {
	switch os.Getenv("GHORG_COLOR") {
	case "enabled":
		color.New(color.FgGreen).Println(msg)
	default:
		fmt.Println(msg)
	}
}

// PrintError prints red colored text to standard out
func PrintError(msg interface{}) {
	switch os.Getenv("GHORG_COLOR") {
	case "enabled":
		color.New(color.FgRed).Println(msg)
	default:
		fmt.Println(msg)
	}
}

// PrintSubtleInfo prints magenta colored text to standard out
func PrintSubtleInfo(msg interface{}) {
	switch os.Getenv("GHORG_COLOR") {
	case "enabled":
		color.New(color.FgHiMagenta).Println(msg)
	default:
		fmt.Println(msg)
	}
}
