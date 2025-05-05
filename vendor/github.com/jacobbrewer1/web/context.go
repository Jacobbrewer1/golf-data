package web

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

// CoreContext provides a context that is canceled when an interrupt signal is received.
//
// This function creates a context that listens for operating system signals such as
// `os.Interrupt` and `syscall.SIGTERM`. It is useful for gracefully shutting down
// applications when these signals are received.
//
// Returns:
//   - context.Context: The base context that is canceled on receiving the specified signals.
//   - context.CancelFunc: A function to manually cancel the context.
//
// Note:
//   - This is defined as a variable to allow users of the package to override it if needed.
//     For example, users may want to use a different signal or context for testing or customization.
var CoreContext = func() (context.Context, context.CancelFunc) {
	return signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
}
