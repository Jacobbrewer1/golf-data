package web

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

// CoreContext returns a context that is cancelled when an interrupt signal is received.
//
// Note: This is not a named function to allow for overriding by users of the package who may want to use a different
// core context. This is useful for testing or if the user wants to use a different signal or context.
var CoreContext = func() (context.Context, context.CancelFunc) {
	return signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
}
