//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=config/server.yaml openapi.yaml
//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=config/models.yaml openapi.yaml
package api

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"

	"github.com/elisasre/go-common/v2/service/module/httpserver"
)

//go:embed openapi/*
var openapiUI embed.FS

//go:embed openapi.yaml
var apiFiles []byte

type TaskHandler interface {
	ListTasks(context.Context, ListTasksParams) ([]Task, *Error)
	CreateTask(context.Context, TaskData) (Task, *Error)
	DeleteTask(context.Context, uint64) *Error
	UpdateTask(context.Context, uint64, TaskData) (Task, *Error)
}

type API struct {
	h TaskHandler
}

func New(h TaskHandler) (http.Handler, error) {
	fsys, err := fs.Sub(openapiUI, "openapi")
	if err != nil {
		return nil, err
	}

	mux := http.NewServeMux()
	mux.Handle("GET /healthz", http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	mux.Handle("GET /openapi.yaml", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.Write(apiFiles) })) //nolint: errcheck
	mux.Handle("/openapi/", http.StripPrefix("/openapi/", http.FileServer(http.FS(fsys))))
	return HandlerWithOptions(&API{h: h}, StdHTTPServerOptions{BaseRouter: mux}), nil
}

func (a *API) ListTasks(w http.ResponseWriter, r *http.Request, params ListTasksParams) {
	tasks, err := a.h.ListTasks(r.Context(), params)
	if err != nil {
		err.Send(w)
		return
	}
	writeJSON(w, tasks)
}

func (a *API) CreateTask(w http.ResponseWriter, r *http.Request) {
	var data TaskData
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		(&Error{http.StatusBadRequest, err.Error()}).Send(w)
		return
	}

	task, err := a.h.CreateTask(r.Context(), data)
	if err != nil {
		err.Send(w)
		return
	}
	writeJSON(w, task)
}

func (a *API) DeleteTask(w http.ResponseWriter, r *http.Request, taskID uint64) {
	err := a.h.DeleteTask(r.Context(), taskID)
	if err != nil {
		err.Send(w)
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *API) UpdateTask(w http.ResponseWriter, r *http.Request, taskID uint64) {
	var data TaskData
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		(&Error{http.StatusBadRequest, err.Error()}).Send(w)
		return
	}

	task, err := a.h.UpdateTask(r.Context(), taskID, data)
	if err != nil {
		err.Send(w)
		return
	}
	writeJSON(w, task)
}

// WithHandler returns an httpserver.Opt that sets the API handler.
func WithHandler(th TaskHandler) httpserver.Opt {
	return func(s *httpserver.Server) error {
		h, err := New(th)
		if err != nil {
			return fmt.Errorf("failed to create API handler: %w", err)
		}
		return httpserver.WithHandler(h)(s)
	}
}

type Error struct {
	Code    int
	Details string
}

func (e *Error) Send(w http.ResponseWriter) {
	w.WriteHeader(e.Code)
	json.NewEncoder(w).Encode(ErrorResponse{Details: e.Details}) //nolint: errcheck
}

func writeJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v) //nolint: errcheck
}
