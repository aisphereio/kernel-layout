package biz

import (
	"context"
	"strings"
	"time"

	"github.com/aisphereio/kernel/errorx"
	"go.einride.tech/aip/filtering"
	"go.einride.tech/aip/ordering"
)

var (
	ErrTodoNotFound        = errorx.NotFound(errorx.Code("TODO_NOT_FOUND"), "todo not found")
	ErrTodoInvalidArgument = errorx.BadRequest(errorx.Code("TODO_INVALID_ARGUMENT"), "invalid todo argument")
)

type Todo struct {
	ID         int64
	Title      string
	Content    string
	Completed  bool
	CreateTime time.Time
	UpdateTime time.Time
}

type TodoRepo interface {
	FindByID(context.Context, int64) (*Todo, error)
	ListTodos(context.Context, ...ListOption) ([]*Todo, error)
	CreateTodo(context.Context, *Todo) (*Todo, error)
	UpdateTodo(context.Context, *Todo) (*Todo, error)
	DeleteTodo(context.Context, int64) error
}

type ListOption func(*ListOptions)

type ListOptions struct {
	Filter  filtering.Filter
	OrderBy ordering.OrderBy
	Offset  int
	Limit   int
}

func ListFilter(filter filtering.Filter) ListOption {
	return func(o *ListOptions) { o.Filter = filter }
}

func ListOrderBy(orderBy ordering.OrderBy) ListOption {
	return func(o *ListOptions) { o.OrderBy = orderBy }
}

func ListOffset(offset int) ListOption {
	return func(o *ListOptions) { o.Offset = offset }
}

func ListLimit(limit int) ListOption {
	return func(o *ListOptions) { o.Limit = limit }
}

type TodoUsecase struct {
	repo TodoRepo
}

func NewTodoUsecase(repo TodoRepo) *TodoUsecase {
	return &TodoUsecase{repo: repo}
}

func (uc *TodoUsecase) GetTodo(ctx context.Context, id int64) (*Todo, error) {
	return uc.repo.FindByID(ctx, id)
}

func (uc *TodoUsecase) ListTodos(ctx context.Context, opts ...ListOption) ([]*Todo, error) {
	return uc.repo.ListTodos(ctx, opts...)
}

func (uc *TodoUsecase) CreateTodo(ctx context.Context, todo *Todo) (*Todo, error) {
	if err := validateTodo(todo); err != nil {
		return nil, err
	}
	return uc.repo.CreateTodo(ctx, todo)
}

func (uc *TodoUsecase) UpdateTodo(ctx context.Context, todo *Todo) (*Todo, error) {
	if todo == nil || todo.ID <= 0 {
		return nil, ErrTodoInvalidArgument
	}
	if err := validateTodo(todo); err != nil {
		return nil, err
	}
	return uc.repo.UpdateTodo(ctx, todo)
}

func (uc *TodoUsecase) DeleteTodo(ctx context.Context, id int64) error {
	if id <= 0 {
		return ErrTodoInvalidArgument
	}
	return uc.repo.DeleteTodo(ctx, id)
}

func validateTodo(todo *Todo) error {
	if todo == nil || strings.TrimSpace(todo.Title) == "" {
		return ErrTodoInvalidArgument
	}
	return nil
}
