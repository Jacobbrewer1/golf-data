package logging

import (
	"context"
	"log/slog"
)

// Ensure dedupeHandler implements the slog.Handler interface.
var _ slog.Handler = (*dedupeHandler)(nil)

// dedupeHandler is a custom slog.Handler that deduplicates attributes.
// It efficiently handles attribute merging when creating new handlers.
type dedupeHandler struct {
	// base is the underlying slog.Handler to which log records are delegated.
	base slog.Handler

	// attrs is a precomputed, immutable list of deduplicated attributes.
	attrs []slog.Attr
}

// NewDedupeHandler creates a new dedupeHandler instance.
// The base handler must not be nil.
func NewDedupeHandler(base slog.Handler) slog.Handler {
	if base == nil {
		panic("base handler cannot be nil")
	}
	return &dedupeHandler{
		base:  base,
		attrs: nil, // Use nil instead of empty slice for better memory efficiency
	}
}

// Enabled checks if the specified log level is enabled.
func (d *dedupeHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return d.base.Enabled(ctx, level)
}

// Handle processes a log record.
func (d *dedupeHandler) Handle(ctx context.Context, record slog.Record) error { // nolint:gocritic // Part of an interface
	// Clone the record to avoid modifying the original.
	record = record.Clone()
	// Add deduplicated attributes if any exist.
	if len(d.attrs) > 0 {
		record.AddAttrs(d.attrs...)
	}
	return d.base.Handle(ctx, record)
}

// WithAttrs returns a new handler with additional attributes.
// Duplicate keys are overridden by the new attributes.
func (d *dedupeHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	if len(attrs) == 0 {
		return d // No new attributes, return self
	}

	// Fast path: if no existing attributes, just check for duplicates in new attrs
	if len(d.attrs) == 0 {
		if len(attrs) == 1 {
			// Single attribute, no deduplication needed
			return &dedupeHandler{
				base:  d.base,
				attrs: []slog.Attr{attrs[0]},
			}
		}

		// Check if there are any duplicates in new attrs
		seen := make(map[string]struct{}, len(attrs))
		hasDuplicates := false
		for _, attr := range attrs {
			if _, exists := seen[attr.Key]; exists {
				hasDuplicates = true
				break
			}
			seen[attr.Key] = struct{}{}
		}

		if !hasDuplicates {
			// No duplicates, can reuse the slice directly
			newAttrs := make([]slog.Attr, len(attrs))
			copy(newAttrs, attrs)
			return &dedupeHandler{
				base:  d.base,
				attrs: newAttrs,
			}
		}
	}

	// Slow path: need to deduplicate
	combined := make(map[string]slog.Attr, len(d.attrs)+len(attrs))

	// Start with existing deduplicated attributes.
	for _, attr := range d.attrs {
		combined[attr.Key] = attr
	}

	// Add new attributes, overriding duplicates.
	for _, attr := range attrs {
		combined[attr.Key] = attr
	}

	// Convert back to slice (order is not guaranteed but doesn't matter for logging)
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
// The group is applied to the base handler, and existing attributes are preserved.
func (d *dedupeHandler) WithGroup(name string) slog.Handler {
	if name == "" {
		return d // Empty group name, return self
	}

	return &dedupeHandler{
		base:  d.base.WithGroup(name),
		attrs: d.attrs, // Preserve existing attributes
	}
}
