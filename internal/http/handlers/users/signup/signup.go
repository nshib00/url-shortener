package signup

import (
	"errors"
	"go-url-shortener/internal/storage"
	resp "go-url-shortener/internal/utils/api/response"
	"go-url-shortener/internal/utils/logger"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
)

type UserCreator interface {
	CreateUser(username string, password string) (int64, error)
}

type Request struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required,min=8"`
}

type Response struct {
	resp.Response
	UserID int `json:"user_id"`
}

type SignupHandler struct {
	log         *slog.Logger
	userCreator UserCreator
}

func New(logger *slog.Logger, userCreator UserCreator) *SignupHandler {
	return &SignupHandler{
		log:         logger,
		userCreator: userCreator,
	}
}

func (h *SignupHandler) Handle(w http.ResponseWriter, r *http.Request) {
	const operation = "handlers.signup.Handle"

	log := h.log.With(
		slog.String("op", operation),
		slog.String("request_id", middleware.GetReqID(r.Context())),
	)
	var req Request
	if err := render.DecodeJSON(r.Body, &req); err != nil {
		log.Error("handlers[signup]: failed to decode request body", logger.Err(err))
		render.JSON(w, r, resp.Error("request body decoding failed"))
		return
	}
	log.Info("handlers[signup]: request body decoded", slog.Any("request", req))

	if err := validator.New().Struct(req); err != nil {
		validationErr := err.(validator.ValidationErrors)
		log.Error("handlers[signup]: invalid request", logger.Err(err))
		render.JSON(w, r, resp.ValidationError(validationErr))
		return
	}

	username := req.Username
	password := req.Password
	if username == "" || password == "" {
		log.Info("handlers[signup]: username or password is empty")
		render.JSON(w, r, "invalid request: username or password is empty")
		return
	}
	userID, err := h.userCreator.CreateUser(username, password)
	if err != nil {
		if errors.Is(err, storage.ErrUserAlreadyExists) {
			log.Info("handlers[signup]: user already exists", slog.Any("username", req))
			render.JSON(w, r, resp.Error("user already exists"))
			return
		}
		log.Info("handlers[signup]: failed to create user", slog.Any("username", req))
		render.JSON(w, r, resp.Error("failed to signup"))
		return
	}
	log.Info(
		"handlers[signup]: user created",
		slog.Int64("user_id", userID),
		slog.String("username", username),
	)
	render.JSON(w, r, Response{
		Response: resp.OK(),
		UserID:   int(userID),
	})
}
