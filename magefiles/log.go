//go:build mage

package main

import (
	"fmt"
	"log/slog"
)

func log(level slog.Level, msg string) {
	if level == slog.LevelDebug {
		if !isDebugMode() {
			return
		}
	}

	fmt.Printf("[%s] %s\n", level.String(), msg)
}
