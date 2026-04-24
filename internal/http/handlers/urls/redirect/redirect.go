package redirect

import (
	resp "go-url-shortener/internal/utils/api/response"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

type URLGetter interface {
	GetURL(alias string) (string, error)
}

type Response struct {
	resp.Response
	URL string `json:"url" validate:"required,url"`
}

type RedirectHandler struct {
	log       *slog.Logger
	urlGetter URLGetter
}

func New(logger *slog.Logger, urlGetter URLGetter) *RedirectHandler {
	return &RedirectHandler{
		log:       logger,
		urlGetter: urlGetter,
	}
}

func (h *RedirectHandler) Handle(w http.ResponseWriter, r *http.Request) {
	operation := "handlers.redirect.Handle"

	log := h.log.With(
		slog.String("op", operation),
		slog.String("request_id", middleware.GetReqID(r.Context())),
	)

	alias := chi.URLParam(r, "alias")
	if alias == "" {
		log.Info("handlers[redirect]: alias is empty")
		render.JSON(w, r, "invalid request: alias is empty")
		return
	}
	url, err := h.urlGetter.GetURL(alias)
	if err != nil {
		log.Error("handlers[redirect]: failed to find url by alias", slog.String("alias", alias))
		render.JSON(w, r, resp.Error("internal error"))
		return
	}

	log.Info(
		"handlers[redirect]: redirect successful",
		slog.String("alias", alias),
		slog.String("url", url),
	)
	http.Redirect(w, r, url, http.StatusFound)
}
