package routers

import (
	hDelete "go-url-shortener/internal/http/handlers/delete"
	"go-url-shortener/internal/http/handlers/redirect"
	"go-url-shortener/internal/http/handlers/save"
	httpMiddleware "go-url-shortener/internal/http/middleware"
	"log/slog"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func New(
	log *slog.Logger,
	saveHandler *save.SaveHandler,
	redirectHandler *redirect.RedirectHandler,
	deleteHandler *hDelete.DeleteHandler,
) *chi.Mux {
	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(httpMiddleware.LoggerMiddleware(log))
	router.Use(middleware.Recoverer)

	router.Post("/url", saveHandler.Handle)
	router.Get("/{alias}", redirectHandler.Handle)
	router.Delete("/{alias}", deleteHandler.Handle)

	return router
}
