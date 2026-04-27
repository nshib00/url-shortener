package login

import (
	"errors"
	"go-url-shortener/internal/auth"
	"go-url-shortener/internal/hashpwd"
	"go-url-shortener/internal/http/handlers/users"
	"go-url-shortener/internal/storage"
	resp "go-url-shortener/internal/utils/api/response"
	"go-url-shortener/internal/utils/logger"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
)

type Request struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type Response struct {
	resp.Response
	Token string `json:"token"`
}

type UserGetter interface {
	GetUserByName(username string) (users.User, error)
}

type LoginHandler struct {
	log        *slog.Logger
	userGetter UserGetter
	secretKey  string
}

func New(logger *slog.Logger, userGetter UserGetter, secretKey string) *LoginHandler {
	return &LoginHandler{
		log:        logger,
		userGetter: userGetter,
		secretKey:  secretKey,
	}
}

func (h *LoginHandler) Handle(w http.ResponseWriter, r *http.Request) {
	const operation = "handlers.login.Handle"

	log := h.log.With(
		slog.String("op", operation),
		slog.String("request_id", middleware.GetReqID(r.Context())),
	)

	var req Request
	if err := render.DecodeJSON(r.Body, &req); err != nil {
		log.Error("handlers[login]: failed to decode request body", logger.Err(err))
		render.JSON(w, r, resp.Error("invalid request body"))
		return
	}

	if err := validator.New().Struct(req); err != nil {
		validationErr := err.(validator.ValidationErrors)
		log.Error("handlers[login]: invalid request", logger.Err(err))
		render.JSON(w, r, resp.ValidationError(validationErr))
		return
	}

	user, err := h.userGetter.GetUserByName(req.Username)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			log.Info("handlers[login]: user not found", logger.Err(err))
			render.JSON(w, r, resp.Error("invalid credentials"))
			return
		}
		log.Error("handlers[login]: failed to get user by name", logger.Err(err))
		render.JSON(w, r, resp.Error("internal error"))
		return
	}

	if ok := hashpwd.VerifyPassword(req.Password, user.PasswordHash); !ok {
		log.Info("handlers[login]: failed to verify password")
		render.JSON(w, r, resp.Error("invalid credentials"))
		return
	}

	token, err := auth.GenerateToken(user.ID, h.secretKey)
	if err != nil {
		log.Error("handlers[login]: failed to generate token", logger.Err(err))
		render.JSON(w, r, resp.Error("internal error"))
		return
	}

	render.JSON(w, r, Response{Response: resp.OK(), Token: token})
}
