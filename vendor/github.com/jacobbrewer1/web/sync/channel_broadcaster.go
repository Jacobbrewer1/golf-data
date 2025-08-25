package sync

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"go.uber.org/multierr"

	"github.com/jacobbrewer1/web/logging"
)

// ChannelBroadcaster is a generic broadcaster that sends a single message of type T to multiple channels.
type ChannelBroadcaster[T any] struct {
	// mut is a mutex to protect the channels slice.
	mut sync.RWMutex

	// l is the logger for the broadcaster.
	l *slog.Logger

	// broadcastingChan is a channel used to send messages that are going to be broadcasted.
	broadcastingChan chan T

	// channels is a slice of channels to broadcast messages to.
	channels map[string]chan<- T
}

// NewChannelBroadcaster creates a new ChannelBroadcaster.
func NewChannelBroadcaster[T any](l *slog.Logger) *ChannelBroadcaster[T] {
	return &ChannelBroadcaster[T]{
		l:                l,
		broadcastingChan: make(chan T, 1), // Buffered channel to avoid blocking on send
		channels:         make(map[string]chan<- T),
	}
}

// Start starts the broadcaster, listening for messages on the broadcasting channel and sending them to all registered channels.
func (b *ChannelBroadcaster[T]) Start(ctx context.Context) {
	b.l.Info("starting channel broadcaster")

	go func() {
		for {
			select {
			case <-ctx.Done():
				// If the context is done, we stop processing messages
				return
			case msg, ok := <-b.broadcastingChan:
				if !ok {
					b.l.Info("broadcasting channel closed, stopping broadcaster")
					return
				}

				var broadcastErrs error
				b.mut.RLock()
				for name, ch := range b.channels {
					channelLogger := b.l.With(
						"channel_id", name,
					)

					channelLogger.Debug("broadcasting message to channel",
						"channel_id", name,
					)

					select {
					case ch <- msg: // Send message to each channel
					default:
						// Skip the channel if it is full
						err := fmt.Errorf("channel %s is full, skipping message", name)
						broadcastErrs = multierr.Append(broadcastErrs, err)
					}
				}
				b.mut.RUnlock()

				if broadcastErrs != nil {
					b.l.Error("errors occurred while broadcasting message",
						logging.KeyError, broadcastErrs,
					)
				} else {
					b.l.Debug("message broadcasted successfully")
				}
			}
		}
	}()
}

// GetBroadcastingChannel returns the channel used for broadcasting messages.
//
// Note: This channel is write only!
func (b *ChannelBroadcaster[T]) GetBroadcastingChannel() chan<- T {
	return b.broadcastingChan
}

// AddChannel adds a new channel to the broadcaster.
func (b *ChannelBroadcaster[T]) AddChannel(name string, ch chan<- T) {
	b.mut.Lock()
	defer b.mut.Unlock()

	if _, exists := b.channels[name]; exists {
		b.l.Warn("channel already exists, skipping addition",
			"channel_id", name,
		)
		return
	}

	b.channels[name] = ch // Add the new channel to the map
}

// RemoveChannel removes a channel from the broadcaster by its name.
// Note: This method does not close the channel, as the broadcaster doesn't own it.
func (b *ChannelBroadcaster[T]) RemoveChannel(name string) {
	b.mut.Lock()
	defer b.mut.Unlock()

	_, exists := b.channels[name]
	if !exists {
		b.l.Warn("channel does not exist, skipping removal",
			"channel_id", name,
		)
		return
	}

	delete(b.channels, name) // Remove the channel from the map
}

// Close stops the broadcaster and clears all registered channels.
// Note: This method does not close the registered channels, as the broadcaster doesn't own them.
// It only closes its own broadcasting channel to signal shutdown.
func (b *ChannelBroadcaster[T]) Close() {
	b.mut.Lock()
	defer b.mut.Unlock()

	// Close our own broadcasting channel to signal shutdown
	close(b.broadcastingChan)

	// Clear the channels map but don't close the channels since we don't own them
	b.channels = make(map[string]chan<- T)
}
