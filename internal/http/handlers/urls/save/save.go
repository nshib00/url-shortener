package save

import (
	"errors"
	"go-url-shortener/internal/storage"
	resp "go-url-shortener/internal/utils/api/response"
	"go-url-shortener/internal/utils/logger"
	"go-url-shortener/internal/utils/random"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
)

const aliasLength = 8

//go:generate go run github.com/vektra/mockery/v2 --name=URLSaver
type URLSaver interface {
	SaveURL(urlToSave string, alias string) (int64, error)
}

type Request struct {
	URL   string `json:"url" validate:"required,url"`
	Alias string `json:"alias,omitempty"`
}

type Response struct {
	resp.Response
	Alias string `json:"alias,omitempty"`
}

type SaveHandler struct {
	log      *slog.Logger
	urlSaver URLSaver
}

func New(logger *slog.Logger, urlSaver URLSaver) *SaveHandler {
	return &SaveHandler{
		log:      logger,
		urlSaver: urlSaver,
	}
}

func (h *SaveHandler) Handle(w http.ResponseWriter, r *http.Request) {
	const operation = "handlers.save.Handle"

	log := h.log.With(
		slog.String("op", operation),
		slog.String("request_id", middleware.GetReqID(r.Context())),
	)
	var req Request
	err := render.DecodeJSON(r.Body, &req)
	if err != nil {
		log.Error("handlers[save]: failed to decode request body", logger.Err(err))
		render.JSON(w, r, resp.Error("request body decoding failed"))
		return
	}
	log.Info("handlers[save]: request body decoded", slog.Any("request", req))

	if err := validator.New().Struct(req); err != nil {
		validationErr := err.(validator.ValidationErrors)
		log.Error("handlers[save]: invalid request", logger.Err(err))
		render.JSON(w, r, resp.ValidationError(validationErr))
		return
	}

	alias := req.Alias
	if alias == "" {
		alias = random.NewRandomString(aliasLength)
	}
	id, err := h.urlSaver.SaveURL(req.URL, alias)
	if err != nil {
		if errors.Is(err, storage.ErrURLAlreadyExists) {
			log.Info("handlers[save]: url already exists", slog.String("url", req.URL))
			render.JSON(w, r, resp.Error("url already exists"))
			return
		}
		log.Error("handlers[save]: failed to save url", logger.Err(err))
		render.JSON(w, r, resp.Error("failed to add url"))
		return
	}

	log.Info("handlers[save]: url added", slog.Int64("id", id))
	render.JSON(w, r, Response{
		Response: resp.OK(),
		Alias:    alias,
	})
}
