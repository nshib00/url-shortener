package routers

import (
	hDelete "go-url-shortener/internal/http/handlers/urls/delete"
	"go-url-shortener/internal/http/handlers/urls/redirect"
	"go-url-shortener/internal/http/handlers/urls/save"
	"go-url-shortener/internal/http/handlers/users/signup"
	httpMiddleware "go-url-shortener/internal/http/middleware"
	"log/slog"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func New(
	log *slog.Logger,
	secretKey string,
	saveHandler *save.SaveHandler,
	redirectHandler *redirect.RedirectHandler,
	deleteHandler *hDelete.DeleteHandler,
	signupHandler *signup.SignupHandler,
	//loginHandler *login.LoginHandler,
) *chi.Mux {
	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(httpMiddleware.LoggerMiddleware(log))
	router.Use(middleware.Recoverer)

	router.Post("/url", saveHandler.Handle)
	router.Delete("/{alias}", deleteHandler.Handle)

	router.Post("/auth/signup", signupHandler.Handle)
	//router.Post("/auth/login", loginHandler.Handle)
	router.Get("/{alias}", redirectHandler.Handle)

	router.Group(func(r chi.Router) {
		r.Use(httpMiddleware.AuthMiddleware(secretKey))
		// r.Post("/url", saveHandler.Handle)
		// r.Delete("/{alias}", deleteHandler.Handle)
	})

	return router
}
