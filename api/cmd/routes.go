package main

import (
	"github.com/gorilla/mux"
	"net/http"
)

func (app *application) routes() http.Handler {
	router := mux.NewRouter()

	// health check
	router.HandleFunc("/v1/healthcheck", app.healthcheckHandler).Methods("GET")
	// courses
	router.HandleFunc("/v1/courses", app.courses.CoursesAllHandler).Methods("POST")
	router.HandleFunc("/v1/course/{id}", app.courses.CoursesIdHandler).Methods("GET")
	router.HandleFunc("/v1/create-course", app.courses.CreateCourseHandler).Methods("POST")
	router.HandleFunc("/v1/update-course", app.courses.UpdateCourseHandler).Methods("PUT")
	router.HandleFunc("/v1/delete-course/{id}", app.courses.DeleteCourseHandler).Methods("DELETE")

	// login
	router.HandleFunc("/v1/login", app.login.LoginHandler).Methods("POST")
	router.HandleFunc("/v1/google", app.login.LoginGoogleHandler).Methods("POST")
	router.HandleFunc("/v1/logout", app.login.LogoutHandler).Methods("POST")
	router.HandleFunc("/v1/password-reset", app.login.PasswordResetHandler).Methods("POST")
	router.HandleFunc("/v1/create-account", app.login.CreateAccountHandler).Methods("POST")

	return app.middleware.Metrics(app.middleware.RecoverPanic(
		app.middleware.EnableCORS(app.middleware.RateLimit(app.middleware.Authenticate(router)))))
}
