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

func Warn(msg string) {
	Stderr.Println(yellow(msg))
}

func Error(msg string) {
	Stderr.Println(red(msg))
}
