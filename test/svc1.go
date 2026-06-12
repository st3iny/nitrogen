package main

import (
	"log/slog"
	"os"
	"time"
)

func main() {
	slog.Info("starting service 1")
	time.Sleep(2 * time.Second)
	os.Exit(1)
}
