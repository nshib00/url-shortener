package routers

import (
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
	//deleteHandler *delete.DeleteHandler,
) *chi.Mux {
	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(httpMiddleware.LoggerMiddleware(log))
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)

	router.Post("/url", saveHandler.Handle)
	router.Post("/{alias}", redirectHandler.Handle)
	// router.Delete("/{alias}", deleteHandler.Handle)

	return router
}
