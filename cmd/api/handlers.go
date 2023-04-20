package main

import (
	"net/http"
	"strconv"
	"time"

	"github.com/Babatunde50/distributask/internal/database"
	"github.com/Babatunde50/distributask/internal/password"
	"github.com/Babatunde50/distributask/internal/request"
	"github.com/Babatunde50/distributask/internal/response"
	"github.com/Babatunde50/distributask/internal/validator"

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

// TODO: make payload more robust,
// TODO: handle validation robustly
func (app *application) createTask(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Type       string              `json:"type"`
		Payload    database.Payload    `json:"payload"`
		Priority   int                 `json:"priority"`
		Timeout    int                 `json:"timeout"`
		MaxRetries int                 `json:"max_retries"`
		Validator  validator.Validator `json:"-"`
	}

	err := request.DecodeJSON(w, r, &input)
	if err != nil {
		app.badRequest(w, r, err)
		return
	}

	// validate type field
	input.Validator.CheckField(input.Type != "", "Type", "Type is required")
	input.Validator.CheckField(input.Type != "image_processing" || input.Type != "data_processing", "Type", "Type must be either image_processing or data_processing")

	// validate payload field
	input.Validator.CheckField(input.Payload.ImageURL != "", "Payload", "Payload is required")
	input.Validator.CheckField(input.Payload.ResizeWidth != 0, "Payload", "Payload is required")
	input.Validator.CheckField(input.Payload.ResizeHeight != 0, "Payload", "Payload is required")

	if input.Validator.HasErrors() {
		app.failedValidation(w, r, input.Validator)
		return
	}

	authenticatedUser := contextGetAuthenticatedUser(r)

	// create task
	task := database.Task{
		Type:       input.Type,
		Payload:    input.Payload,
		Priority:   input.Priority,
		Timeout:    input.Timeout,
		MaxRetries: input.MaxRetries,
		UserId:     authenticatedUser.ID,
	}

	// send created task to user
	err = app.db.InsertTask(&task)

	if err != nil {
		app.serverError(w, r, err)
		return
	}

	app.writeJSON(w, http.StatusCreated, task, nil)
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
