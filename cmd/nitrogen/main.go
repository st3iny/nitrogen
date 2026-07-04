package main

import (
	"log/slog"
	"os"
	"runtime/debug"

	"github.com/st3iny/nitrogen/internal/nitrogen"
)

var (
	Version   = "(devel)"
	BuildDate = "(unknown)"
	GoVersion = goVersion()
)

func main() {
	slog.Info("starting nitrogen", "version", Version, "buildDate", BuildDate, "compiler", GoVersion)

	err := nitrogen.Run()
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
}

func goVersion() string {
	buildInfo, ok := debug.ReadBuildInfo()
	if !ok {
		return "(unknown)"
	}

	return buildInfo.GoVersion
}
