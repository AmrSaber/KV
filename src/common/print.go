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
	Yellow = color.New(color.FgYellow).SprintfFunc()
	Red    = color.New(color.FgRed).SprintfFunc()
	Blue   = color.New(color.FgBlue).SprintfFunc()
	Green  = color.New(color.FgGreen).SprintfFunc()
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

func Warn(msg string, args ...any) {
	Stderr.Println(Yellow(msg, args...))
}

func PrintCrossDBWarning() {
	Warn("Copying across DBs, transactional guarantees do not apply")
}
