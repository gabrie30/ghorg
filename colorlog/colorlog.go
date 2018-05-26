package colorlog

import "github.com/fatih/color"

func PrintInfo(msg interface{}) {
	color.New(color.FgYellow).Println(msg)
}

func PrintSuccess(msg interface{}) {
	color.New(color.FgGreen).Println(msg)
}

func PrintError(msg interface{}) {
	color.New(color.FgRed).Println(msg)
}

func PrintSubtleInfo(msg interface{}) {
	color.New(color.FgHiMagenta).Println(msg)
}
