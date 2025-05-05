package logging

import (
	"context"
	"log/slog"
)

// Ensure dedupeHandler implements the slog.Handler interface.
//
// This declaration ensures that the dedupeHandler type satisfies the slog.Handler
// interface at compile time. If dedupeHandler does not implement all the methods
// required by the slog.Handler interface, a compile-time error will occur.
var _ slog.Handler = new(dedupeHandler)

// dedupeHandler is a custom slog.Handler that deduplicates attributes.
//
// This struct is used to wrap another slog.Handler and ensure that attributes
// with duplicate keys are deduplicated before being passed to the underlying handler.
type dedupeHandler struct {
	// base is the underlying slog.Handler to which log records are delegated.
	base slog.Handler

	// attrs is a precomputed, immutable list of deduplicated attributes.
	attrs []slog.Attr
}

// NewDedupeHandler creates a new dedupeHandler instance.
//
// This function initializes a dedupeHandler with the provided base slog.Handler
// and an empty list of attributes.
//
// Parameters:
//   - base (slog.Handler): The underlying handler to delegate log records to.
//
// Returns:
//   - slog.Handler: A new dedupeHandler instance.
func NewDedupeHandler(base slog.Handler) slog.Handler {
	return &dedupeHandler{
		base:  base,
		attrs: make([]slog.Attr, 0),
	}
}

// Enabled checks if the specified log level is enabled.
//
// This method delegates the check to the underlying handler.
//
// Parameters:
//   - ctx (context.Context): The context for the log operation.
//   - level (slog.Level): The log level to check.
//
// Returns:
//   - bool: True if the log level is enabled, false otherwise.
func (d *dedupeHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return d.base.Enabled(ctx, level)
}

// Handle processes a log record.
//
// This method clones the log record, adds deduplicated attributes, and then
// delegates the handling to the underlying handler.
//
// Parameters:
//   - ctx (context.Context): The context for the log operation.
//   - record (slog.Record): The log record to process.
//
// Returns:
//   - error: An error if the underlying handler fails to process the record.
func (d *dedupeHandler) Handle(ctx context.Context, record slog.Record) error { // nolint:gocritic // Part of an interface
	// Clone the record to avoid modifying the original.
	record = record.Clone()
	// Add deduplicated attributes temporarily.
	record.AddAttrs(d.attrs...)
	return d.base.Handle(ctx, record)
}

// WithAttrs returns a new handler with additional attributes.
//
// This method deduplicates the provided attributes by overriding existing ones
// with the same keys and returns a new dedupeHandler instance.
//
// Parameters:
//   - attrs ([]slog.Attr): The attributes to add.
//
// Returns:
//   - slog.Handler: A new dedupeHandler instance with the combined attributes.
func (d *dedupeHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	combined := make(map[string]slog.Attr)

	// Start with existing deduplicated attributes.
	for _, attr := range d.attrs {
		combined[attr.Key] = attr
	}

	// Add new attributes, overriding duplicates.
	for _, attr := range attrs {
		combined[attr.Key] = attr
	}

	// Rebuild the slice of attributes.
	newAttrs := make([]slog.Attr, 0, len(combined))
	for _, attr := range combined {
		newAttrs = append(newAttrs, attr)
	}

	return &dedupeHandler{
		base:  d.base,
		attrs: newAttrs,
	}
}

// WithGroup returns a new handler with a group name.
//
// This method creates a new dedupeHandler instance that wraps the underlying
// handler with the specified group name.
//
// Parameters:
//   - name (string): The name of the group.
//
// Returns:
//   - slog.Handler: A new dedupeHandler instance with the group name applied.
func (d *dedupeHandler) WithGroup(name string) slog.Handler {
	return &dedupeHandler{
		base:  d.base.WithGroup(name),
		attrs: d.attrs,
	}
}
