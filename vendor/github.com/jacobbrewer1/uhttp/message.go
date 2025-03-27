package uhttp

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/jacobbrewer1/uhttp/common"
)

// NewMessage creates a new Message.
func NewMessage(message string) *common.Message {
	return &common.Message{
		Message: message,
	}
}

func MustSendMessageWithStatus(w http.ResponseWriter, status int, message string) {
	if err := SendMessageWithStatus(w, status, message); err != nil {
		slog.Error("Failed to send message", slog.String(loggingKeyError, err.Error()))
	}
}

func SendMessageWithStatus(w http.ResponseWriter, status int, message string) error {
	msg := NewMessage(message)
	if err := EncodeJSON(w, status, msg); err != nil {
		return fmt.Errorf("failed to encode message: %w", err)
	}
	return nil
}

func MustSendMessage(w http.ResponseWriter, message string) {
	if err := SendMessage(w, message); err != nil {
		slog.Error("Failed to send message", slog.String(loggingKeyError, err.Error()))
	}
}

func SendMessage(w http.ResponseWriter, message string) error {
	if err := SendMessageWithStatus(w, http.StatusOK, message); err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}
	return nil
}
