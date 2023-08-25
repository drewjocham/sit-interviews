package main

import (
	"interviews/pkg/logger"
	"net/http"
)

type envelope map[string]any

func (app *application) healthcheckHandler(w http.ResponseWriter, r *http.Request) {
	env := envelope{
		"status": "available",
		"system_info": map[string]string{
			"environment": app.config.Env,
			"version":     app.config.Version,
		},
	}

	err := app.helper.WriteJSON(w, http.StatusOK, env, nil)
	if err != nil {
		app.logger.ErrorCtx(err, logger.Ctx{
			"msg": "Unable to write health check response",
		})
	}
}
