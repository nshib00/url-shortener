package url

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

type URLSaver interface {
	SaveURL(urlToSave string, alias string) (int64, error)
}

type Request struct {
	URL   string `json:"url" validate:"required, url"`
	Alias string `json:"alias,omitempty"`
}

type Response struct {
	resp.Response
	Alias string `json:"alias,omitempty"`
}

func NewSaveHandler(log *slog.Logger, urlSaver URLSaver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const operation = "handlers.url.save.New"

		log = log.With(
			slog.String("op", operation),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)
		var req Request
		err := render.DecodeJSON(r.Body, &req)
		if err != nil {
			log.Error("handlers[save]: failed to decode response body", logger.Err(err))
			render.JSON(w, r, "handlers[save]: request body decoding failed")
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
		id, err := urlSaver.SaveURL(req.URL, alias)
		if err != nil {
			if errors.Is(err, storage.ErrURLAlreadyExists) {
				log.Info("handlers[save]: url already exists", slog.String("url", req.URL))
				render.JSON(w, r, resp.Error("url already exists"))
				return
			}
		}

		log.Info("handlers[save]: url added", slog.Int64("id", id))
		render.JSON(w, r, Response{
			Response: resp.OK(),
			Alias:    alias,
		})
	}
}
