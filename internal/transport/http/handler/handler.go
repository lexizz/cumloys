package handler

import (
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httprate"
	"github.com/go-chi/jwtauth/v5"

	"github.com/lexizz/cumloys/internal/config"
	"github.com/lexizz/cumloys/internal/models"
	"github.com/lexizz/cumloys/internal/pkg/logger"
	"github.com/lexizz/cumloys/internal/service"
	"github.com/lexizz/cumloys/internal/transport/http/handler/urlrouter"
)

type handler struct {
	services *service.Services
	config   *config.Config
	logger   logger.Logger
	jwt      *models.JWT
}

type Response struct {
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
	Token  string `json:"token,omitempty"`
}

func New(cfg *config.Config, logger logger.Logger, servicesList *service.Services, jwt *models.JWT) *handler {
	return &handler{
		services: servicesList,
		config:   cfg,
		logger:   logger,
		jwt:      jwt,
	}
}

func (h *handler) Init() http.Handler {
	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.AllowContentType("application/json", "text/plain"))
	router.Use(middleware.RealIP)
	router.Use(middleware.CleanPath)
	router.Use(middleware.Compress(9, "application/json", "text/plain"))
	router.Use(httprate.LimitAll(60, 1*time.Minute))

	urlRoute := urlrouter.New(h.jwt, h.logger)

	router.Route("/", func(r chi.Router) {
		errFn := func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusMethodNotAllowed)

			_, writeError := w.Write([]byte(r.Method + ": method is not valid"))
			if writeError != nil {
				h.logger.Infof("---> ERROR: write: %v", writeError)
				http.Error(w, urlrouter.ErrInternalServer.Error(), http.StatusInternalServerError)
				return
			}
		}

		r.MethodNotAllowed(errFn)

		r.NotFound(errFn)
	})

	router.Route("/api/user", func(routerAPI chi.Router) {
		routerAPI.Group(func(r chi.Router) {
			r.Route("/", func(r chi.Router) {
				errFn := func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusMethodNotAllowed)

					_, writeError := w.Write([]byte(r.Method + ": method is not valid"))
					if writeError != nil {
						h.logger.Infof("---> ERROR: write: %v", writeError)
						http.Error(w, urlrouter.ErrInternalServer.Error(), http.StatusInternalServerError)
						return
					}
				}

				r.MethodNotAllowed(errFn)
				r.NotFound(errFn)
			})

			r.Post("/login", urlRoute.AuthenticationHandler(h.services.FindUserService))
			r.Post("/register", urlRoute.RegistrationHandler(h.services.CreateUserService))
		})

		routerAPI.Group(func(r chi.Router) {
			r.Use(Verifier(h.jwt.Auth))
			r.Use(jwtauth.Authenticator)

			r.Post("/orders", urlRoute.AddingOrdersHandler(
				h.services.FindOrderService,
				h.services.GettingPointsService,
			))
			r.Get("/orders", urlRoute.GettingOrdersHandler(h.services.FindOrderService))

			r.Route("/balance", func(routerBalance chi.Router) {
				routerBalance.Get("/", urlRoute.GettingCurrentBalanceHandler(h.services.FindBalanceService))
				routerBalance.Post("/withdraw", urlRoute.WithdrawPointsHandler(
					h.services.CreateOrderService,
					h.services.FindOrderService,
					h.services.WithdrawPointsService,
				))
			})

			r.Get("/withdrawals", urlRoute.GettingInfoAboutBalanceHandler(h.services.FindWithdrawPointsService))
		})
	})

	return router
}

func Verifier(ja *jwtauth.JWTAuth) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return jwtauth.Verify(ja, TokenFromHeader)(next)
	}
}

func TokenFromHeader(r *http.Request) string {
	bearer := r.Header.Get("Authorization")
	if hasBearerToken(bearer) && len(bearer) > 7 && strings.ToUpper(bearer[0:6]) == "BEARER" {
		return bearer[7:]
	}
	return bearer
}

func hasBearerToken(token string) bool {
	return strings.Contains(token, "bearer") || strings.Contains(token, "BEARER")
}
