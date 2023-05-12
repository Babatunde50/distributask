package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (app *application) routes() http.Handler {
	mux := chi.NewRouter()

	mux.NotFound(app.notFound)
	mux.MethodNotAllowed(app.methodNotAllowed)

	mux.Use(app.recoverPanic)
	mux.Use(app.authenticate)

	mux.Get("/status", app.status)
	mux.Post("/users", app.createUser)
	mux.Post("/authentication-tokens", app.createAuthenticationToken)

	mux.Group(func(mux chi.Router) {
		mux.Use(app.requireAuthenticatedUser)

		// Submit a new task to the task queue.
		mux.Post("/tasks", app.createTask)

		// Retrieve detailed information about a specific task by its ID.
		mux.Get("/tasks/{taskID}", app.getTask)

		// Retrieve a list of tasks with optional filters like task status, priority, or date range.
		mux.Get("/tasks", app.listTasks)

		// Remove a task from the task queue.
		mux.Delete("/tasks/{taskID}", app.deleteTask)

	})

	return mux
}
