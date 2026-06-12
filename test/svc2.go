package main

import (
	"log/slog"
	"time"
)

func main() {
	slog.Info("starting service 2")
	time.Sleep(100 * time.Second)
}
