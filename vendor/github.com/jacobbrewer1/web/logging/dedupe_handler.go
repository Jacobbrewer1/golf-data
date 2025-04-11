package logging

import (
	"context"
	"log/slog"
)

type dedupeHandler struct {
	base slog.Handler

	// precomputed deduplicated attributes — immutable
	attrs []slog.Attr
}

func (d *dedupeHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return d.base.Enabled(ctx, level)
}

func (d *dedupeHandler) Handle(ctx context.Context, record slog.Record) error { // nolint:gocritic // Part of an interface
	// Don't modify the original record — add deduplicated attrs temporarily
	record = record.Clone()
	record.AddAttrs(d.attrs...)
	return d.base.Handle(ctx, record)
}

func (d *dedupeHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	combined := make(map[string]slog.Attr)

	// Start with existing deduped attrs
	for _, attr := range d.attrs {
		combined[attr.Key] = attr
	}

	// Add new attrs, overriding duplicates
	for _, attr := range attrs {
		combined[attr.Key] = attr
	}

	// Rebuild slice
	newAttrs := make([]slog.Attr, 0, len(combined))
	for _, attr := range combined {
		newAttrs = append(newAttrs, attr)
	}

	return &dedupeHandler{
		base:  d.base,
		attrs: newAttrs,
	}
}

func (d *dedupeHandler) WithGroup(name string) slog.Handler {
	return &dedupeHandler{
		base:  d.base.WithGroup(name),
		attrs: d.attrs,
	}
}

func NewDedupeHandler(base slog.Handler) slog.Handler {
	return &dedupeHandler{
		base:  base,
		attrs: make([]slog.Attr, 0),
	}
}
