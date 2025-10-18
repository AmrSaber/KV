package common

import (
	"log"
	"os"

	"github.com/fatih/color"
)

var (
	Stdout = log.New(os.Stdout, "", 0)
	Stderr = log.New(os.Stderr, "", 0)
)

var (
	yellow = color.New(color.FgYellow).SprintFunc()
	red    = color.New(color.FgRed).SprintFunc()
)

func Quiet(enable bool) {
	if enable {
		devNull, _ := os.Open(os.DevNull)
		Stdout.SetOutput(devNull)
		Stderr.SetOutput(devNull)
	} else {
		Stdout.SetOutput(os.Stdout)
		Stderr.SetOutput(os.Stderr)
	}
}

func Warn(msg string) {
	Stderr.Println(yellow(msg))
}

func Error(msg string, args ...any) {
	Stderr.Printf(red(msg)+"\n", args...)
}
