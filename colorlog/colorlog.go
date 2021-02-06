// Package colorlog has various Print functions that can be called to change the color of the text in standard out
package colorlog

import (
	"fmt"
	"github.com/fatih/color"
)

var UseColor = false

// PrintInfo prints yellow colored text to standard out
func PrintInfo(msg interface{}) {
	if UseColor {
		color.New(color.FgYellow).Println(msg)
	} else {
		fmt.Println(msg)
	}
}

// PrintSuccess prints green colored text to standard out
func PrintSuccess(msg interface{}) {
	if UseColor {
		color.New(color.FgGreen).Println(msg)
	} else {
		fmt.Println(msg)
	}
}

// PrintError prints red colored text to standard out
func PrintError(msg interface{}) {
	if UseColor {
		color.New(color.FgRed).Println(msg)
	} else {
		fmt.Println(msg)
	}
}

// PrintSubtleInfo prints magenta colored text to standard out
func PrintSubtleInfo(msg interface{}) {
	if UseColor {
		color.New(color.FgHiMagenta).Println(msg)
	} else {
		fmt.Println(msg)
	}
}
