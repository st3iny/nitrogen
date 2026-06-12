package main

import (
	"log/slog"
	"os"

	"github.com/st3iny/nitrogen/internal/nitrogen"
)

func main() {
	err := nitrogen.Run()
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
}
