package main

import (
	"context"
	"errors"
	"expvar"
	"fmt"
	"interviews/internal/auth"
	"interviews/internal/config"
	data "interviews/internal/users"
	"interviews/pkg"
	log "interviews/pkg/logger"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/felixge/httpsnoop"
	"github.com/tomasen/realip"
	"golang.org/x/time/rate"
)

var (
	ErrPanic          = errors.New("recovering panic")
	ErrLimitReached   = errors.New("rate limit exceeded")
	ErrRecordNotFound = errors.New("record not found")
	ErrEditConflict   = errors.New("edit conflict")
)

type MiddleWare struct {
	cfg       config.Config
	userCtx   data.UsersContext
	users     data.UserRepo
	tokenRepo auth.TokenRepository
	e         pkg.CustomErrors
}

func (app *MiddleWare) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.Background()
		clog := log.GetLoggerFromContext(ctx)

		w.Header().Add("Vary", "Authorization")

		authorizationHeader := r.Header.Get("Authorization")

		if authorizationHeader == "" {
			r = app.userCtx.ContextSetUser(r, data.AnonymousUser)
			next.ServeHTTP(w, r)
			return
		}

		headerParts := strings.Split(authorizationHeader, " ")
		if len(headerParts) != 2 || headerParts[0] != "Bearer" {
			app.e.InvalidAuthenticationTokenResponse(w, r)
			return
		}

		token := headerParts[1]

		email, err := app.tokenRepo.GetEmailFromJWT(token)
		if err != nil {
			//TODO: add a log here
			return
		}

		user, err := app.users.GetByEmail(email)
		if err != nil {
			clog.ErrorCtx(err, log.Ctx{
				"msg": "error getting user by email",
			})
			return
		}

		r = app.userCtx.ContextSetUser(r, user)

		next.ServeHTTP(w, r)
	})
}

func (app *MiddleWare) RecoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				w.Header().Set("Connection", "close")
				app.e.ServerErrorResponse(w, r, fmt.Errorf("%s", err))
			}
		}()

		next.ServeHTTP(w, r)
	})
}

func (app *MiddleWare) RateLimit(next http.Handler) http.Handler {
	type client struct {
		limiter  *rate.Limiter
		lastSeen time.Time
	}

	var (
		mu      sync.Mutex
		clients = make(map[string]*client)
	)

	go func() {
		for {
			time.Sleep(time.Minute)

			mu.Lock()

			for ip, c := range clients {
				if time.Since(c.lastSeen) > 3*time.Minute {
					delete(clients, ip)
				}
			}

			mu.Unlock()
		}
	}()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if app.cfg.Limiter.Enabled {
			ip := realip.FromRequest(r)

			mu.Lock()

			if _, found := clients[ip]; !found {
				clients[ip] = &client{
					limiter: rate.NewLimiter(rate.Limit(app.cfg.Limiter.Rps), app.cfg.Limiter.Burst),
				}
			}

			clients[ip].lastSeen = time.Now()

			if !clients[ip].limiter.Allow() {
				mu.Unlock()
				app.e.RateLimitExceededResponse(w, r)
				return
			}

			mu.Unlock()
		}

		next.ServeHTTP(w, r)
	})
}

func (app *MiddleWare) requireAuthenticatedUser(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := app.userCtx.ContextGetUser(r)

		if user.IsAnonymous() {
			app.e.AuthenticationRequiredResponse(w, r)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (app *MiddleWare) RequireActivatedUser(next http.HandlerFunc) http.HandlerFunc {
	fn := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := app.userCtx.ContextGetUser(r)

		if !user.Activated {
			app.e.InactiveAccountResponse(w, r)
			return
		}

		next.ServeHTTP(w, r)
	})

	return app.requireAuthenticatedUser(fn)
}

func (app *MiddleWare) EnableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Vary", "Origin")

		w.Header().Add("Vary", "Access-Control-Request-Method")

		origin := r.Header.Get("Origin")

		app.cfg.Cors.TrustedOrigins = []string{"http://localhost:5173"}

		if origin != "" {
			for i := range app.cfg.Cors.TrustedOrigins {
				if origin == app.cfg.Cors.TrustedOrigins[i] {
					w.Header().Set("Access-Control-Allow-Origin", origin)

					if r.Method == http.MethodOptions && r.Header.Get("Access-Control-Request-Method") != "" {

						w.Header().Set("Access-Control-Allow-Methods", "OPTIONS, PUT, PATCH, DELETE")
						w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
						w.WriteHeader(http.StatusOK)

						return
					}

					break
				}
			}
		}

		next.ServeHTTP(w, r)
	})
}

func (app *MiddleWare) Metrics(next http.Handler) http.Handler {
	totalRequestsReceived := expvar.NewInt("total_requests_received")
	totalResponsesSent := expvar.NewInt("total_responses_sent")
	totalProcessingTimeMicroseconds := expvar.NewInt("total_processing_time_μs")

	totalResponsesSentByStatus := expvar.NewMap("total_responses_sent_by_status")

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		totalRequestsReceived.Add(1)

		metrics := httpsnoop.CaptureMetrics(next, w, r)

		totalResponsesSent.Add(1)

		totalProcessingTimeMicroseconds.Add(metrics.Duration.Microseconds())

		totalResponsesSentByStatus.Add(strconv.Itoa(metrics.Code), 1)
	})
}

func NewMiddleware(cfg config.Config,
	userCtx data.UsersContext, users data.UserRepo,
	tokenRepo auth.TokenRepository) *MiddleWare {
	return &MiddleWare{
		cfg:       cfg,
		userCtx:   userCtx,
		users:     users,
		tokenRepo: tokenRepo,
		e:         pkg.CustomErrors{},
	}
}
