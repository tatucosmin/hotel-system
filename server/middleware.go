package server

import (
	"context"
	"log/slog"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/tatucosmin/hotel-system/store"
)

func NewLoggerMiddleware(logger *slog.Logger) func(h http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger.Info("incoming http request", "method", r.Method, "path", r.URL.Path)
			h.ServeHTTP(w, r)
		})
	}
}

type ContextUserKey struct{}

func WithUserContext(ctx context.Context, user *store.User) context.Context {
	return context.WithValue(ctx, ContextUserKey{}, user)
}

var admin_routes = []string{"/api/tickets"}

func NewPermissionsMiddleware() func(h http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.HasPrefix(r.URL.Path, "/api/auth") {
				h.ServeHTTP(w, r)
				return
			}

			user, err := GetUserFromContext(r.Context())
			if err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			for _, route := range admin_routes {
				if strings.HasPrefix(r.URL.Path, route) && !user.HasRole(store.RoleAdmin) {
					w.WriteHeader(http.StatusForbidden)
					return
				}
			}

			h.ServeHTTP(w, r)
		})
	}
}

func NewAuthMiddleware(jwtManager *JwtManager, userStore *store.UserStore) func(h http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.HasPrefix(r.URL.Path, "/api/auth") {
				h.ServeHTTP(w, r)
				return
			}

			authHeader := r.Header.Get("Authorization")
			var token string
			if splitted := strings.Split(authHeader, " "); len(splitted) == 2 {
				token = splitted[1]
			}

			if token == "" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			parsedToken, err := jwtManager.ParseToken(token)
			if err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				slog.Error("failed to parse token", "error", err)
				return
			}

			if !jwtManager.IsAccessToken(parsedToken) {
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte("providing refresh tokens is not permitted"))
				return
			}

			parsedUserId, err := parsedToken.Claims.GetSubject()
			if err != nil {
				slog.Error("failed to get subject from token", "error", err)
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			userId, err := uuid.Parse(parsedUserId)
			if err != nil {
				slog.Error("couldn't parse user id string into an uuid", "error", err)
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			user, err := userStore.ById(r.Context(), userId)
			if err != nil {
				slog.Error("couldn't get user from db with user uuid", "error", err)
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			h.ServeHTTP(w, r.WithContext(WithUserContext(r.Context(), user)))

		})
	}
}
