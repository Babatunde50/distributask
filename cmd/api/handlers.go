package main

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/Babatunde50/distributask/internal/database"
	"github.com/Babatunde50/distributask/internal/password"
	"github.com/Babatunde50/distributask/internal/request"
	"github.com/Babatunde50/distributask/internal/response"
	"github.com/Babatunde50/distributask/internal/validator"
	"github.com/Babatunde50/distributask/internal/worker"
	"github.com/go-chi/chi/v5"
	"github.com/hibiken/asynq"

	"github.com/pascaldekloe/jwt"
)

func (app *application) status(w http.ResponseWriter, r *http.Request) {
	data := map[string]string{
		"Status": "OK",
	}

	err := response.JSON(w, http.StatusOK, data)
	if err != nil {
		app.serverError(w, r, err)
	}
}

func (app *application) createUser(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email     string              `json:"email"`
		Password  string              `json:"password"`
		Validator validator.Validator `json:"-"`
	}

	err := request.DecodeJSON(w, r, &input)
	if err != nil {
		app.badRequest(w, r, err)
		return
	}

	existingUser, err := app.db.GetUserByEmail(input.Email)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	// validate email
	input.Validator.CheckField(input.Email != "", "Email", "Email is required")
	input.Validator.CheckField(validator.Matches(input.Email, validator.RgxEmail), "Email", "Must be a valid email address")
	input.Validator.CheckField(existingUser == nil, "Email", "Email is already in use")

	// validate password
	input.Validator.CheckField(input.Password != "", "Password", "Password is required")
	input.Validator.CheckField(len(input.Password) >= 8, "Password", "Password is too short")
	input.Validator.CheckField(len(input.Password) <= 72, "Password", "Password is too long")
	input.Validator.CheckField(validator.NotIn(input.Password, password.CommonPasswords...), "Password", "Password is too common")

	if input.Validator.HasErrors() {
		app.failedValidation(w, r, input.Validator)
		return
	}

	hashedPassword, err := password.Hash(input.Password)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	_, err = app.db.InsertUser(input.Email, "user", hashedPassword)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusCreated, input, nil)

	if err != nil {
		app.serverError(w, r, err)
	}

}

// TODO: handle validation robustly
func (app *application) listTasks(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Page     int    `json:"page"`
		PageSize int    `json:"page_size"`
		Sort     string `json:"sort"`
	}

	qs := r.URL.Query()

	input.Page = app.readInt(qs, "page", 1)
	input.PageSize = app.readInt(qs, "page_size", 10)
	input.Sort = app.readString(qs, "sort", "id")

	authenticatedUser := contextGetAuthenticatedUser(r)

	tasks, err := app.db.ListTasks(authenticatedUser.ID, database.Filters{Page: input.Page, PageSize: input.PageSize, Sort: input.Sort})

	if err != nil {
		app.serverError(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, tasks, nil)

	if err != nil {
		app.serverError(w, r, err)
	}

}

func (app *application) updateTask(w http.ResponseWriter, r *http.Request) {

	taskID := chi.URLParam(r, "taskID")

	taskIdInt, err := strconv.Atoi(taskID)

	if err != nil {
		app.badRequest(w, r, err)
		return
	}

	authenticatedUser := contextGetAuthenticatedUser(r)

	task, err := app.db.GetTask(taskIdInt, authenticatedUser.ID)

	if err != nil {
		app.serverError(w, r, err)
		return
	}

	if task == nil {
		app.notFound(w, r)
		return
	}

	var input struct {
		Payload   *database.Payload   `json:"payload"`
		Type      *string             `json:"type"`
		Validator validator.Validator `json:"-"`
	}

	err = request.DecodeJSON(w, r, &input)

	if err != nil {
		app.badRequest(w, r, err)
		return
	}

	if input.Payload != nil {
		task.Payload = *input.Payload
	}

	if input.Type != nil {
		task.Type = *input.Type
	}

	// TODO: validate task input

	err = app.db.UpdateTask(task)

	if err != nil {
		app.badRequest(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, task, nil)

	if err != nil {
		app.serverError(w, r, err)
	}

}

func (app *application) deleteTask(w http.ResponseWriter, r *http.Request) {

	taskID := chi.URLParam(r, "taskID")

	taskIdInt, err := strconv.Atoi(taskID)

	if err != nil {
		app.badRequest(w, r, err)
		return
	}

	authenticatedUser := contextGetAuthenticatedUser(r)

	err = app.db.DeleteTask(taskIdInt, authenticatedUser.ID)

	if err != nil {
		app.badRequest(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusNoContent, nil, nil)

	if err != nil {
		app.serverError(w, r, err)
	}

}

func (app *application) getTask(w http.ResponseWriter, r *http.Request) {

	taskId := chi.URLParam(r, "taskID")

	taskIdInt, err := strconv.Atoi(taskId)

	if err != nil {
		app.badRequest(w, r, err)
		return
	}

	authenticatedUser := contextGetAuthenticatedUser(r)

	task, err := app.db.GetTask(taskIdInt, authenticatedUser.ID)

	if err != nil {
		app.serverError(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, task, nil)

	if err != nil {
		app.serverError(w, r, err)
	}
}

// TODO: make payload more robust,
// TODO: handle validation robustly
func (app *application) createTask(w http.ResponseWriter, r *http.Request) {

	var input struct {
		Type      string                     `json:"type"`
		Payload   database.Payload           `json:"payload"`
		Params    database.AllPossibleParams `json:"params"`
		Validator validator.Validator        `json:"-"`
	}

	err := request.DecodeJSON(w, r, &input)

	if err != nil {
		app.badRequest(w, r, err)
		return
	}

	// TODO: handle validation..

	input.Payload.UpdateParams(input.Payload.Operation, input.Params)

	// if input.Validator.HasErrors() {
	// 	app.failedValidation(w, r, input.Validator)
	// 	return
	// }

	authenticatedUser := contextGetAuthenticatedUser(r)

	task := database.Task{
		Type:       input.Type,
		Payload:    input.Payload,
		Priority:   2,
		Timeout:    30,
		MaxRetries: 5,
		UserId:     authenticatedUser.ID,
	}

	// insert task and distribute task to worker node..
	err = app.db.InsertTask(&task, func(createdTask *database.Task) error {
		ctx, cancel := context.WithTimeout(context.Background(), 4000)
		defer cancel()

		err = app.taskDistributor.DistributeTaskSendTask(ctx, &worker.PayloadSendTask{
			TaskID: createdTask.ID,
			UserID: createdTask.UserId,
		}, asynq.MaxRetry(createdTask.MaxRetries), asynq.Timeout(time.Duration(createdTask.Timeout)*time.Second))

		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		app.serverError(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusCreated, task, nil)

	if err != nil {
		app.serverError(w, r, err)
		return
	}
}

func (app *application) createAuthenticationToken(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email     string              `json:"email"`
		Password  string              `json:"password"`
		Validator validator.Validator `json:"-"`
	}

	err := request.DecodeJSON(w, r, &input)
	if err != nil {
		app.badRequest(w, r, err)
		return
	}

	user, err := app.db.GetUserByEmail(input.Email)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	input.Validator.CheckField(input.Email != "", "Email", "Email is required")
	input.Validator.CheckField(user != nil, "Email", "Email address could not be found")

	if user != nil {
		passwordMatches, err := password.Matches(input.Password, user.PasswordHash)
		if err != nil {
			app.serverError(w, r, err)
			return
		}

		input.Validator.CheckField(input.Password != "", "Password", "Password is required")
		input.Validator.CheckField(passwordMatches, "Password", "Password is incorrect")
	}

	if input.Validator.HasErrors() {
		app.failedValidation(w, r, input.Validator)
		return
	}

	// create JWT
	var claims jwt.Claims
	claims.Subject = strconv.Itoa(user.ID)

	expiry := time.Now().Add(24 * time.Hour)
	claims.Issued = jwt.NewNumericTime(time.Now())
	claims.NotBefore = jwt.NewNumericTime(time.Now())
	claims.Expires = jwt.NewNumericTime(expiry)

	claims.Issuer = app.config.baseURL
	claims.Audiences = []string{app.config.baseURL}

	jwtBytes, err := claims.HMACSign(jwt.HS256, []byte(app.config.jwt.secretKey))
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	data := map[string]string{
		"AuthenticationToken":       string(jwtBytes),
		"AuthenticationTokenExpiry": expiry.Format(time.RFC3339),
	}

	err = app.writeJSON(w, http.StatusCreated, data, nil)

	if err != nil {
		app.serverError(w, r, err)
	}
}
