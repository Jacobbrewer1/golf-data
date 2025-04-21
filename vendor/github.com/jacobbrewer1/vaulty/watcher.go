package vaulty

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	hashiVault "github.com/hashicorp/vault/api"
)

func monitorWatcher(ctx context.Context, l *slog.Logger, token string, watcher *hashiVault.LifetimeWatcher) (renewResult, error) {
	for {
		select {
		case <-ctx.Done():
			return exitRequested, nil

			// DoneCh will return if renewal fails, or if the remaining lease
			// duration is under a built-in threshold and either renewing is not
			// extending it or renewing is disabled.  In both cases, the caller
			// should attempt a re-read of the secret. Clients should check the
			// return value of the channel to see if renewal was successful.
		case err := <-watcher.DoneCh():
			// Leases created by a token get revoked when the token is revoked.
			return expiring, err

			// RenewCh is a channel that receives a message when a successful
			// renewal takes place and includes metadata about the renewal.
		case info := <-watcher.RenewCh():
			l.Debug("renewal successful",
				slog.String(loggingKeyRenewedAt, info.RenewedAt.String()),
				slog.String(loggingKeySecretName, token),
				slog.String(loggingKeyLeaseDuration, fmt.Sprintf("%ds", info.Secret.LeaseDuration)),
			)
		}
	}
}

func handleWatcherResult(l *slog.Logger, result renewResult, onExpire ...func()) error {
	switch {
	case result&exitRequested != 0:
		l.Debug("result is exitRequested", slog.Int(loggingKeyResult, int(result)))
		return nil
	case result&expiring != 0:
		l.Debug("result is expiring", slog.Int(loggingKeyResult, int(result)))
		if len(onExpire) == 0 {
			return errors.New("no onExpire functions provided")
		}
		for _, f := range onExpire {
			f()
		}
		return nil
	default:
		l.Debug("no action required", slog.Int(loggingKeyResult, int(result)))
		return nil
	}
}
