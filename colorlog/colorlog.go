package colorlog

import "github.com/fatih/color"

// PrintInfo prints yellow colored text to standard out
func PrintInfo(msg interface{}) {
	color.New(color.FgYellow).Println(msg)
}

// PrintSuccess prints green colored text to standard out
func PrintSuccess(msg interface{}) {
	color.New(color.FgGreen).Println(msg)
}

// PrintError prints red colored text to standard out
func PrintError(msg interface{}) {
	color.New(color.FgRed).Println(msg)
}

// PrintSubtleInfo prints magenta colored text to standard out
func PrintSubtleInfo(msg interface{}) {
	color.New(color.FgHiMagenta).Println(msg)
}
