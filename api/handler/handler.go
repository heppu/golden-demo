package handler

import (
	"context"
	"database/sql"
	"errors"
	"net/http"

	"github.com/heppu/golden-demo/api"
	"github.com/heppu/golden-demo/store"
	"github.com/lib/pq"
)

type Handler struct {
	s Store
}

type Store interface {
	ListTasks(context.Context) ([]store.Task, error)
	ListTasksFiltered(context.Context, store.TaskStatus) ([]store.Task, error)
	CreateTask(context.Context, store.TaskData) (store.Task, error)
	DeleteTask(context.Context, uint64) error
	UpdateTask(context.Context, uint64, store.TaskData) (store.Task, error)
}

func New(s Store) *Handler {
	return &Handler{s: s}
}

func (h *Handler) ListTasks(ctx context.Context, filter api.ListTasksParams) ([]api.Task, *api.Error) {
	var (
		tasks []store.Task
		err   error
	)

	if filter.Status != nil {
		tasks, err = h.s.ListTasksFiltered(ctx, store.ParseStatus(string(*filter.Status)))
	} else {
		tasks, err = h.s.ListTasks(ctx)
	}
	if err != nil {
		return nil, convertErr(err)
	}

	resp := make([]api.Task, 0, len(tasks))
	for _, t := range tasks {
		resp = append(resp, api.Task{
			Id:          t.ID,
			CreatedAt:   t.CreatedAt,
			Title:       t.Title,
			Status:      api.Status(t.Status.String()),
			Description: t.Description,
		})
	}
	return resp, nil
}

func (h *Handler) CreateTask(ctx context.Context, data api.TaskData) (api.Task, *api.Error) {
	task, err := h.s.CreateTask(ctx, store.TaskData{
		Title:       data.Title,
		Description: data.Description,
		Status:      store.ParseStatus(string(data.Status)),
	})
	if err != nil {
		return api.Task{}, convertErr(err)
	}

	return api.Task{
		Id:          task.ID,
		CreatedAt:   task.CreatedAt,
		Title:       task.Title,
		Status:      api.Status(task.Status.String()),
		Description: task.Description,
	}, nil
}

func (h *Handler) DeleteTask(ctx context.Context, id uint64) *api.Error {
	return convertErr(h.s.DeleteTask(ctx, id))
}

func (h *Handler) UpdateTask(ctx context.Context, id uint64, data api.TaskData) (api.Task, *api.Error) {
	task, err := h.s.UpdateTask(ctx, id, store.TaskData{
		Title:       data.Title,
		Description: data.Description,
		Status:      store.ParseStatus(string(data.Status)),
	})
	if err != nil {
		return api.Task{}, convertErr(err)
	}

	return api.Task{
		Id:          task.ID,
		CreatedAt:   task.CreatedAt,
		Title:       task.Title,
		Status:      api.Status(task.Status.String()),
		Description: task.Description,
	}, nil
}

func convertErr(err error) *api.Error {
	var pgErr *pq.Error
	switch {
	case err == nil:
		return nil
	case errors.Is(err, sql.ErrNoRows):
		return &api.Error{Code: http.StatusNotFound, Details: err.Error()}
	case errors.As(err, &pgErr) && pgErr.Code.Name() == "unique_violation":
		return &api.Error{Code: http.StatusConflict, Details: err.Error()}
	case errors.As(err, &pgErr) && pgErr.Code.Name() == "invalid_text_representation":
		return &api.Error{Code: http.StatusBadRequest, Details: err.Error()}
	default:
		return &api.Error{Code: http.StatusInternalServerError, Details: err.Error()}
	}
}
