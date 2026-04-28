package delete

import (
	"errors"
	"log/slog"
	"net/http"

	httpMiddleware "go-url-shortener/internal/http/middleware"
	"go-url-shortener/internal/storage"
	resp "go-url-shortener/internal/utils/api/response"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

type URLDeleter interface {
	DeleteURL(userID int, alias string) error
}

type DeleteHandler struct {
	log        *slog.Logger
	urlDeleter URLDeleter
}

func New(logger *slog.Logger, urlDeleter URLDeleter) *DeleteHandler {
	return &DeleteHandler{
		log:        logger,
		urlDeleter: urlDeleter,
	}
}

func (h *DeleteHandler) Handle(w http.ResponseWriter, r *http.Request) {
	operation := "handlers.delete.Handle"

	log := h.log.With(
		slog.String("op", operation),
		slog.String("request_id", middleware.GetReqID(r.Context())),
	)

	ctxVal := r.Context().Value(httpMiddleware.UserIDKey)
	userID, ok := ctxVal.(int)
	if !ok {
		log.Info("handlers[save]: unauthorized: request with wrong user ID", slog.Any("userID", ctxVal))
		render.JSON(w, r, resp.Error("unauthorized"))
		return
	}

	alias := chi.URLParam(r, "alias")
	if alias == "" {
		log.Info("handlers[delete]: empty alias passed")
		render.JSON(w, r, "invalid request: alias is empty")
		return
	}
	if err := h.urlDeleter.DeleteURL(userID, alias); err != nil {
		if errors.Is(err, storage.ErrURLNotFound) {
			log.Info("handlers[delete]: not found or forbidden", slog.String("alias", alias))
			render.JSON(w, r, resp.Error("not found or forbidden"))
			return
		}
		log.Error("handlers[delete]: failed to delete url by alias", slog.String("alias", alias))
		render.JSON(w, r, resp.Error("internal error"))
		return
	}
	log.Info("handlers[delete]: url successfully deleted")
	w.WriteHeader(http.StatusNoContent)
}
