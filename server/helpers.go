package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/tatucosmin/hotel-system/store"
)

type ApiError struct {
	status int
	err    error
}

func (e *ApiError) Error() string {
	return e.err.Error()
}

func NewApiError(status int, err error) *ApiError {
	return &ApiError{status, err}
}

func handler(f func(w http.ResponseWriter, r *http.Request) error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := f(w, r); err != nil {
			status := http.StatusInternalServerError
			msg := http.StatusText(status)
			if e, ok := err.(*ApiError); ok {
				status = e.status
				msg = http.StatusText(e.status)
				if e.status == http.StatusBadRequest || e.status == http.StatusConflict || e.status == http.StatusUnauthorized {
					msg = e.err.Error()
				}
			}

			slog.Error("error while executing handler", "error", err, "status", status, "message", msg)
			w.WriteHeader(status)
			if err := json.NewEncoder(w).Encode(ApiResponse[struct{}]{
				Message: msg,
			}); err != nil {
				slog.Error("error encoding response", "error", err, "status", status, "message", msg)
			}
		}
	}
}

func encode[T any](w http.ResponseWriter, status int, v T) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		return fmt.Errorf("encode json: %w", err)
	}
	return nil
}

type Validator interface {
	Validate() error
}

func decode[T Validator](r *http.Request) (T, error) {
	var v T
	if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
		return v, fmt.Errorf("decode json: %w", err)
	}

	if err := v.Validate(); err != nil {
		return v, err
	}
	return v, nil
}

func GetUserFromContext(ctx context.Context) (*store.User, error) {
	user, ok := ctx.Value(ContextUserKey{}).(*store.User)
	if !ok {
		return nil, fmt.Errorf("failed to get user from context")
	}
	return user, nil
}
