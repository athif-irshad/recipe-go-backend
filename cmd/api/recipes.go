package main

import (
	//"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"recipe.athif.com/internal/data"
	"recipe.athif.com/internal/validator"
)

func (app *application) createRecipeHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Title        string    `json:"title"`
		Instructions string    `json:"instructions"`
		PrepTime     data.Mins `json:"preparation_time"`
		CookTime     data.Mins `json:"cooking_time"`
		CuisineName  string    `json:"cuisine_name"` // Change this line
		Difficulty   string    `json:"difficulty"`
	}
	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	recipe := &data.Recipe{
		Title:        input.Title,
		Instructions: input.Instructions,
		PrepTime:     input.PrepTime,
		CookTime:     input.CookTime,
		CuisineName:  input.CuisineName, // Change this line
		Difficulty:   input.Difficulty,
	}

	v := validator.New()
	if data.ValidateRecipe(v, recipe); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Recipes.Insert(recipe)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/recipes/%d", recipe.ID))

	err = app.writeJSON(w, http.StatusCreated, envelope{"recipe": recipe}, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) showRecipeHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	recipe, err := app.models.Recipes.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"recipe": recipe}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}

func (app *application) updateRecipeHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	recipe, err := app.models.Recipes.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	var input struct {
		Title        string
		Instructions string
		PrepTime     data.Mins
		CookTime     data.Mins
		CuisineName    string
		Difficulty   string
	}

	err = app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	recipe.Title = input.Title
	recipe.Instructions = input.Instructions
	recipe.PrepTime = input.PrepTime
	recipe.CookTime = input.CookTime
	recipe.CuisineName = input.CuisineName
	recipe.Difficulty = input.Difficulty
	v := validator.New()
	if data.ValidateRecipe(v, recipe); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}
	err = app.models.Recipes.Update(recipe)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	err = app.writeJSON(w, http.StatusOK, envelope{"recipe": recipe}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) deleteRecipeHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	err = app.models.Recipes.Delete(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	// Return a 200 OK status code along with a success message.
	err = app.writeJSON(w, http.StatusOK, envelope{"message": "recipe successfully deleted"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) listRecipeHandler(w http.ResponseWriter, r *http.Request) {

	var input struct {
		Title     string
		CuisineID int
		data.Filters
	}
	v := validator.New()

	qs := r.URL.Query()

	input.Title = app.readString(qs, "title", "")
	input.CuisineID = app.readInt(qs, "cuisineid", 0, v)

	input.Filters.Page = app.readInt(qs, "page", 1, v)
	input.Filters.PageSize = app.readInt(qs, "pagesize", 50, v)

	input.Filters.Sort = app.readString(qs, "sort", "id")
	input.Filters.SortSafelist = []string{"id", "title", "difficulty", "cuisinename", "-id", "-title", "-difficulty", "-cuisinename"}
	if data.ValidateFilters(v, input.Filters); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	recipes, metadata, err := app.models.Recipes.GetAll(input.Title, input.CuisineID, input.Filters)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	err = app.writeJSON(w, http.StatusOK, envelope{"recipes": recipes,"metadata": metadata}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
