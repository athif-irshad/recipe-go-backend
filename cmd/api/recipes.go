package main

import (
	//"encoding/json"
	"fmt"
	"net/http"
	"time"

	"recipe.athif.com/internal/data"
	"recipe.athif.com/internal/validator"
)

func (app *application) createRecipeHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Title      string    `json:"title"`
		PrepTime   data.Mins `json:"preparation_time"`
		CookTime   data.Mins `json:"cooking_time"`
		CuisineID  int32     `json:"cuisine_id"`
		Difficulty string     `json:"difficulty"`
	}
	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	recipe := &data.Recipe{
		Title:      input.Title,
		PrepTime:   input.PrepTime,
		CookTime:   input.CookTime,
		CuisineID:  input.CuisineID,
		Difficulty: input.Difficulty,
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

	fmt.Fprintf(w, "%+v\n", input)
}

func (app *application) showRecipeHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	recipe := data.Recipe{
		ID:         id,
		CreatedAt:  time.Now(),
		Title:      "Chicken Noodles",
		PrepTime:   15,
		CookTime:   25,
		CuisineID:  6,
		Difficulty: "Easy",
		Version:    1,
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"recipe": recipe}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}
