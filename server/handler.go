package server

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/tatucosmin/hotel-system/store"
)

type SignupRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type ApiResponse[T any] struct {
	Data    *T     `json:"data,omitempty"`
	Message string `json:"message,omitempty"`
}

func (req SignupRequest) Validate() error {
	if req.Email == "" {
		return errors.New("email is required to sign up")
	}

	if req.Password == "" {
		return errors.New("password is required to sign up")
	}

	return nil
}

func (s *Server) signUpHandler() http.HandlerFunc {
	return handler(func(w http.ResponseWriter, r *http.Request) error {

		req, err := decode[SignupRequest](r)

		if err != nil {
			return NewApiError(http.StatusBadRequest, err)
		}

		existingUser, err := s.store.User.ByEmail(r.Context(), req.Email)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return NewApiError(http.StatusInternalServerError, err)
		}

		if existingUser != nil {
			return NewApiError(http.StatusConflict, fmt.Errorf("email adress is already registered"))
		}

		_, err = s.store.User.CreateUser(r.Context(), req.Email, req.Password)
		if err != nil {
			return NewApiError(http.StatusInternalServerError, err)
		}

		if err := encode[ApiResponse[struct{}]](w, http.StatusCreated, ApiResponse[struct{}]{
			Message: "user has been signed up",
		}); err != nil {
			return NewApiError(http.StatusInternalServerError, err)
		}

		return nil
	})

}

type SigninRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type SigninResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func (req SigninRequest) Validate() error {
	if req.Email == "" {
		return errors.New("email is required to sign in")
	}

	if req.Password == "" {
		return errors.New("password is required to sign in")
	}

	return nil
}

func (s *Server) signInHandler() http.HandlerFunc {
	return handler(func(w http.ResponseWriter, r *http.Request) error {
		req, err := decode[SigninRequest](r)
		if err != nil {
			return NewApiError(http.StatusBadRequest, err)
		}

		user, err := s.store.User.ByEmail(r.Context(), req.Email)
		if err != nil {
			return NewApiError(http.StatusInternalServerError, err)
		}

		if err := user.ComparePasswordHash(req.Password); err != nil {
			return NewApiError(http.StatusUnauthorized, err)
		}

		tokens, err := s.jwtManager.GenerateTokens(user.Id)

		if err != nil {
			return NewApiError(http.StatusUnauthorized, err)
		}

		err = s.store.RefreshToken.RevokeAllFromUser(r.Context(), user.Id)

		if err != nil {
			return NewApiError(http.StatusInternalServerError, err)
		}

		_, err = s.store.RefreshToken.Create(r.Context(), tokens.RefreshToken, user.Id)
		if err != nil {
			return NewApiError(http.StatusInternalServerError, err)
		}

		if err := user.ComparePasswordHash(req.Password); err != nil {
			return NewApiError(http.StatusInternalServerError, err)
		}

		if err := encode[ApiResponse[SigninResponse]](w, http.StatusOK, ApiResponse[SigninResponse]{
			Data: &SigninResponse{
				AccessToken:  tokens.AccessToken.Raw,
				RefreshToken: tokens.RefreshToken.Raw,
			},
		}); err != nil {
			return NewApiError(http.StatusInternalServerError, err)
		}

		return nil

	})
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

func (req RefreshRequest) Validate() error {
	if req.RefreshToken == "" {
		return errors.New("refresh_token is required")
	}

	return nil
}

type RefreshResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func (s *Server) refreshTokenHandler() http.HandlerFunc {
	return handler(func(w http.ResponseWriter, r *http.Request) error {
		req, err := decode[RefreshRequest](r)
		if err != nil {
			return NewApiError(http.StatusBadRequest, err)
		}

		nowRefreshToken, err := s.jwtManager.ParseToken(req.RefreshToken)
		if err != nil {
			return NewApiError(http.StatusUnauthorized, err)
		}

		parsedUserId, err := nowRefreshToken.Claims.GetSubject()
		if err != nil {
			return NewApiError(http.StatusUnauthorized, err)
		}

		userId, err := uuid.Parse(parsedUserId)
		if err != nil {
			return NewApiError(http.StatusUnauthorized, err)
		}

		nowRefreshTokenRow, err := s.store.RefreshToken.ByPK(r.Context(), nowRefreshToken, userId)
		if err != nil {
			status := http.StatusInternalServerError
			if errors.Is(err, sql.ErrNoRows) {
				status = http.StatusUnauthorized
			}

			return NewApiError(status, err)
		}

		if nowRefreshTokenRow.ExpiresAt.Before(time.Now()) {
			return NewApiError(http.StatusUnauthorized, fmt.Errorf("refresh token has expired"))
		}

		tokens, err := s.jwtManager.GenerateTokens(userId)
		if err != nil {
			return NewApiError(http.StatusInternalServerError, err)
		}

		if err = s.store.RefreshToken.RevokeAllFromUser(r.Context(), userId); err != nil {
			return NewApiError(http.StatusInternalServerError, err)
		}

		if _, err := s.store.RefreshToken.Create(r.Context(), tokens.RefreshToken, userId); err != nil {
			return NewApiError(http.StatusInternalServerError, err)
		}

		if err = encode[ApiResponse[RefreshResponse]](w, http.StatusOK, ApiResponse[RefreshResponse]{
			Data: &RefreshResponse{
				AccessToken:  tokens.AccessToken.Raw,
				RefreshToken: tokens.RefreshToken.Raw,
			},
		}); err != nil {
			return NewApiError(http.StatusInternalServerError, err)
		}

		return nil

	})
}

// * Realistically, this should never fail because the user is already authenticated
// * and shouldn't be called in public routes such as /api/auth/*

func (s *Server) getUserFromContext(ctx context.Context) *store.User {
	user, ok := ctx.Value(ContextUserKey{}).(*store.User)
	if !ok {
		return nil
	}
	return user
}

type TicketRequest struct {
	Id uuid.UUID `json:"id"`
}

func (req TicketRequest) Validate() error {
	if req.Id == uuid.Nil {
		return errors.New("id is required")
	}

	return nil
}

func (s *Server) getTicketHandler() http.HandlerFunc {
	return handler(func(w http.ResponseWriter, r *http.Request) error {

		user := s.getUserFromContext(r.Context())

		req, err := decode[TicketRequest](r)
		if err != nil {
			return NewApiError(http.StatusBadRequest, err)
		}

		ticket, err := s.store.Ticket.ById(r.Context(), req.Id)
		if err != nil {
			status := http.StatusInternalServerError
			if errors.Is(err, sql.ErrNoRows) {
				status = http.StatusNotFound
			}
			return NewApiError(status, err)
		}

		if ticket.Creator != user.Id {
			return NewApiError(http.StatusForbidden, fmt.Errorf("you are not allowed to access this ticket"))
		}

		if err := encode[ApiResponse[store.Ticket]](w, http.StatusOK, ApiResponse[store.Ticket]{
			Data: ticket,
		}); err != nil {
			return NewApiError(http.StatusInternalServerError, err)
		}

		return nil
	})
}

type GetAllTicketsResponse struct {
	Tickets []store.Ticket `json:"tickets"`
}

func (s *Server) getAllTicketsHandler() http.HandlerFunc {
	return handler(func(w http.ResponseWriter, r *http.Request) error {
		tickets, err := s.store.Ticket.All(r.Context())
		if err != nil {
			return NewApiError(http.StatusInternalServerError, err)
		}

		if err := encode[ApiResponse[GetAllTicketsResponse]](w, http.StatusOK, ApiResponse[GetAllTicketsResponse]{
			Data: &GetAllTicketsResponse{
				Tickets: tickets,
			},
		}); err != nil {
			return NewApiError(http.StatusInternalServerError, err)
		}
		return nil
	})
}

type CreateTicketRequest struct {
	Title       string               `json:"title"`
	Description string               `json:"description"`
	Priority    store.TicketPriority `json:"priority"`
}

func (req CreateTicketRequest) Validate() error {
	if req.Title == "" {
		return errors.New("title is required")
	}

	if req.Description == "" {
		return errors.New("description is required")
	}

	if !req.Priority.WithinBounds() {
		return errors.New("priority is required")
	}

	return nil
}

func (s *Server) createTicketHandler() http.HandlerFunc {
	return handler(func(w http.ResponseWriter, r *http.Request) error {

		req, err := decode[CreateTicketRequest](r)

		if err != nil {
			return NewApiError(http.StatusBadRequest, err)
		}

		user := s.getUserFromContext(r.Context())

		ticket, err := s.store.Ticket.Create(r.Context(), req.Title, req.Description, user.Id, req.Priority)

		if err != nil {
			return NewApiError(http.StatusInternalServerError, err)
		}

		if err := encode[ApiResponse[store.Ticket]](w, http.StatusCreated, ApiResponse[store.Ticket]{
			Data: ticket,
		}); err != nil {
			return NewApiError(http.StatusInternalServerError, err)
		}

		return nil
	})
}

type UpdateTicketRequest struct {
	Id       uuid.UUID            `json:"id"`
	Priority store.TicketPriority `json:"priority"`
	Status   store.TicketStatus   `json:"status"`
}

func (req UpdateTicketRequest) Validate() error {
	if req.Id == uuid.Nil {
		return errors.New("id is required")
	}

	if !req.Priority.WithinBounds() {
		return errors.New("priority is required")
	}

	if !req.Status.WithinBounds() {
		return errors.New("status is required")
	}

	return nil
}

func (s *Server) updateTicketHandler() http.HandlerFunc {
	return handler(func(w http.ResponseWriter, r *http.Request) error {

		req, err := decode[UpdateTicketRequest](r)

		if err != nil {
			return NewApiError(http.StatusBadRequest, err)
		}

		err = s.store.Ticket.Update(r.Context(), req.Id, req.Priority, req.Status)

		if err != nil {
			return NewApiError(http.StatusInternalServerError, err)
		}

		if err := encode[ApiResponse[struct{}]](w, http.StatusOK, ApiResponse[struct{}]{
			Message: "ticket has been updated",
		}); err != nil {
			return NewApiError(http.StatusInternalServerError, err)
		}

		return nil
	})
}
