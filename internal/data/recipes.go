package data

import (
	"time"

	"recipe.athif.com/internal/validator"
)

type Recipe struct {
	ID         int64     `json:"id"`
	CreatedAt  time.Time `json:"-"`
	Title      string    `json:"title"`
	PrepTime   Mins      `json:"preparation_time"`
	CookTime   Mins      `json:"cooking_time"`
	CuisineID  int32     `json:"cuisine_id"`
	Difficulty int16     `json:"difficulty"`
	Version    int32     `json:"version"`
}

func ValidateRecipe(v *validator.Validator, recipe *Recipe) {
	v.Check(recipe.Title != "", "title", "must be provided")
	v.Check(len(recipe.Title) <= 500, "title", "must not be more than 500 bytes long")
	v.Check(recipe.PrepTime != 0, "preparation_time", "must be provided")
	v.Check(recipe.PrepTime > 0, "preparation_time", "must be a positive integer")
	v.Check(recipe.CookTime != 0, "cooking_time", "must be provided")
	v.Check(recipe.CookTime > 0, "cooking_time", "must be a positive integer")
	v.Check(recipe.CuisineID != 0, "cuisine_id", "must be provided")
	v.Check(recipe.Difficulty != 0, "difficulty", "must be provided")
}
