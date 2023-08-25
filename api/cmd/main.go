package main

import (
	"context"
	"flag"
	_ "github.com/lib/pq"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"interviews/internal/auth"
	"interviews/internal/config"
	"interviews/internal/courses"
	"interviews/internal/login"
	data "interviews/internal/users"
	"interviews/pkg"
	clogger "interviews/pkg/logger"
	"strings"
	"sync"
)

type application struct {
	config     *config.Config
	logger     *clogger.Logger
	courses    *courses.Courses
	login      *login.Login
	middleware *MiddleWare
	helper     pkg.Helper
	db         *mongo.Client
	wg         sync.WaitGroup
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	clog := clogger.GetLoggerFromContext(ctx)

	cfg, err := config.NewConfig()
	if err != nil {
		return
	}

	clog.Info("Starting Application")

	flag.IntVar(&cfg.Port, "port", cfg.Port, "API server port")
	flag.StringVar(&cfg.Env, "env", cfg.Env, "Environment (development|staging|production)")

	flag.Func("cors-trusted-origins", "Trusted CORS origins (space separated)", func(val string) error {
		cfg.Cors.TrustedOrigins = strings.Fields(val)

		return nil
	})

	// mongo database connection
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.MongoConfig.Url))
	if err != nil {
		clog.Error(err)
	}

	defer func() {
		if err = client.Disconnect(ctx); err != nil {
			clog.Error(err)
		}
	}()

	// courses
	courseCollection := client.Database(cfg.MongoConfig.DBName).Collection(cfg.MongoConfig.CourseCollection)
	courseRepo := courses.NewCourseRepository(client, courseCollection)
	courseService := courses.NewCoursesService(courseRepo)

	//jwt token
	tokenCollection := client.Database(cfg.MongoConfig.DBName).Collection(cfg.MongoConfig.TokenCollection)
	tokenRepo := auth.NewTokenRepository(client, tokenCollection)

	// user
	userCollection := client.Database(cfg.MongoConfig.DBName).Collection(cfg.UserConfig.UserCollection)
	userRepo := data.NewUserRepository(client, userCollection, tokenRepo)

	// login
	loginCollection := client.Database(cfg.MongoConfig.DBName).Collection(cfg.MongoConfig.LoginCollection)
	loginRepo := login.NewLoginRepository(client, loginCollection)
	loginService := login.NewLoginService(loginRepo, tokenRepo, userRepo)

	// middleware
	middleware := NewMiddleware(*cfg, data.UsersContext{}, *userRepo, *tokenRepo)

	app := &application{
		config:     cfg,
		logger:     clog,
		db:         client,
		courses:    courseService,
		middleware: middleware,
		login:      loginService,
	}

	err = app.serve()
	if err != nil {
		clog.Error(err)
	}

}
