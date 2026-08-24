package main

import (
	"fmt"
	"net/http"
	"time"

	"greenlight.codewithyash.dev/internal/data"
	"greenlight.codewithyash.dev/internal/validator"
)

func (app *application) createMovieHandler(w http.ResponseWriter, r *http.Request) {
	var createMovieInput struct {
		Title 	string		`json:"title"`
		Year  	int32		`json:"year"`
		Runtime data.Runtime`json:"runtime"`
		Genres	[]string	`json:"genres"`
	}

	err := app.readJSON(w, r, &createMovieInput)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	movie := &data.Movie{
		Title: createMovieInput.Title,
		Year: createMovieInput.Year,
		Runtime: createMovieInput.Runtime,
		Genres: createMovieInput.Genres,
	}

	v := validator.New()

	if data.ValidateMovie(v, movie); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	fmt.Fprintf(w,"%+v\n", createMovieInput)
}


func (app *application) showMovieHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	movie := data.Movie{
		ID: id,
		CreatedAt: time.Now(),
		Title: "Casablanca",
		Runtime: 102,
		Genres: []string{"drama", "romance", "war"},
		Version: 1,
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"movie": movie}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
