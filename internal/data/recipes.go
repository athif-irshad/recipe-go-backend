package data

import (
	"database/sql"
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
	Difficulty string    `json:"difficulty"`
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
	v.Check(recipe.Difficulty != "", "difficulty", "must be provided")
}

type RecipeModel struct {
	DB *sql.DB
}

func (r RecipeModel) Insert(recipe *Recipe) error {
	query := `
	INSERT INTO recipes (recipename, preparationtime, cookingtime, difficultylevel, cuisineid)
	VALUES ($1, $2, $3, $4, $5)
	RETURNING recipeid`

	args := []interface{}{recipe.Title, recipe.PrepTime, recipe.CookTime, recipe.Difficulty, recipe.CuisineID}

	return r.DB.QueryRow(query, args...).Scan(&recipe.ID)
}

func (m RecipeModel) Get(id int64) (*Recipe, error) {
	return nil, nil
}

// Add a placeholder method for updating a specific record in the movies table.
func (m RecipeModel) Update(movie *Recipe) error {
	return nil
}

// Add a placeholder method for deleting a specific record from the movies table.
func (m RecipeModel) Delete(id int64) error {
	return nil
}
