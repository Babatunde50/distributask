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
		// // Retrieve detailed information about a specific task by its ID.
		// mux.Get("/tasks/{taskID}", app.getTask)
		// // Retrieve a list of tasks with optional filters like task status, priority, or date range.
		// mux.Get("/tasks", app.listTasks)
		// // Update the properties of a task, such as priority or status.
		// mux.Put("/tasks/{taskID}", app.updateTask)
		// // Remove a task from the task queue.
		// mux.Delete("/tasks/{taskID}", app.deleteTask)

		// // Retrieve detailed information about a specific worker node by its ID.
		// mux.Get("/nodes/{nodeID}", app.getNode)
		// // Retrieve a list of registered worker nodes with optional filters like node status or tags.
		// mux.Get("/nodes", app.listNodes)
		// // Update the properties of a worker node, such as status, tags, or capacity.
		// mux.Put("/nodes/{nodeID}", app.updateNode)

		// // Retrieve the current status of the scheduler, including task distribution, node availability, and other metrics.
		// mux.Get("/scheduler", app.getSchedulerStatus)
		// // Update the configuration of the scheduler, such as task prioritization rules or load balancing strategies.
		// mux.Put("/scheduler", app.updateSchedulerConfig)

	})

	return mux
}
