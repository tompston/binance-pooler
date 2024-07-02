package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"syro/pkg/app"
	"syro/pkg/lib/logger"
	"syro/pkg/lib/scheduler"
	"time"

	"github.com/go-chi/chi/v5"
)

type API struct {
	app    *app.App
	router *chi.Mux
}

func New(app *app.App, router *chi.Mux) *API {
	return &API{app: app, router: router}
}

type HttpResponse struct {
	Message string `json:"message"`
	Status  int    `json:"status"`
	Data    any    `json:"data"`
}

func Response(w http.ResponseWriter, status int, data any, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(HttpResponse{Message: message, Status: status, Data: data})
}

// Routes adds the API routes to the router
func (api *API) Routes() {
	r := api.router

	r.Route("/syro", func(r chi.Router) {
		r.Get("/logs", api.getLogs)
		r.Get("/cron-job", api.getCronJobs)
		r.Get("/cron-job-executions", api.getCronJobExecutions)
	})
}

func (api *API) getLogs(w http.ResponseWriter, r *http.Request) {
	parseUrlParams := func(r *http.Request) *logger.LogFilter {
		// NOTE: errors in the parser are ignored because the validity
		// of the date is checked in the logger.FindLogs method.
		from := r.URL.Query().Get("from")
		fromTime, _ := time.Parse(time.RFC3339, from)

		to := r.URL.Query().Get("to")
		toTime, _ := time.Parse(time.RFC3339, to)

		return &logger.LogFilter{
			From: fromTime,
			To:   toTime,
			Log: logger.Log{
				Level:   r.URL.Query().Get("level"),
				Data:    r.URL.Query().Get("data"),
				Source:  r.URL.Query().Get("source"),
				Event:   r.URL.Query().Get("event"),
				EventID: r.URL.Query().Get("event_id")}}
	}

	params := parseUrlParams(r)

	data, err := api.app.Logger().FindLogs(*params, 100, 0)
	if err != nil {
		Response(w, 500, nil, err.Error())
		return
	}

	Response(w, 200, data, "")
}

func (api *API) getCronJobs(w http.ResponseWriter, r *http.Request) {
	data, err := api.app.CronStorage().AllJobs()
	if err != nil {
		Response(w, 500, nil, err.Error())
		return
	}

	Response(w, 200, data, "")
}

func (api *API) getCronJobExecutions(w http.ResponseWriter, r *http.Request) {

	parseUrlParams := func(r *http.Request) *scheduler.ExecutionFilter {
		// NOTE: errors in the parser are ignored because the validity
		// of the date is checked in the scheduler.FindExecutions method.
		from := r.URL.Query().Get("from")
		fromTime, _ := time.Parse(time.RFC3339, from)

		to := r.URL.Query().Get("to")
		toTime, _ := time.Parse(time.RFC3339, to)

		initializedAt := r.URL.Query().Get("initialized_at")
		initTime, _ := time.Parse(time.RFC3339, initializedAt)

		executionTime := r.URL.Query().Get("execution_time")
		execTime, _ := strconv.ParseInt(executionTime, 10, 64)

		return &scheduler.ExecutionFilter{
			From: fromTime,
			To:   toTime,
			ExecutionLog: scheduler.ExecutionLog{
				Name:          r.URL.Query().Get("name"),
				InitializedAt: initTime,
				ExecutionTime: time.Duration(execTime),
			},
		}
	}

	params := parseUrlParams(r)

	data, err := api.app.CronStorage().FindExecutions(*params, 100, 0)
	if err != nil {
		Response(w, 500, nil, err.Error())
		return
	}

	Response(w, 200, data, "")
}
