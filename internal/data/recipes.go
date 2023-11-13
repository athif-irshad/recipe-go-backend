package data

import (
	"context"
	"database/sql"
	"errors"
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
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return r.DB.QueryRowContext(ctx, query, args...).Scan(&recipe.ID)
}

func (m RecipeModel) Get(id int64) (*Recipe, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	query := `
    SELECT recipeid, recipename, preparationtime, cookingtime, difficultylevel, cuisineid
    FROM recipes
    WHERE recipeid = $1`

	recipe := &Recipe{}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	err := m.DB.QueryRowContext(ctx, query, id).Scan(&recipe.ID, &recipe.Title, &recipe.PrepTime, &recipe.CookTime, &recipe.Difficulty, &recipe.CuisineID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrRecordNotFound
		} else {
			return nil, err
		}
	}

	return recipe, nil
}

// Add a placeholder method for updating a specific record in the movies table.
func (m RecipeModel) Update(recipe *Recipe) error {
	query := `
    UPDATE recipes
    SET recipename = $1, preparationtime = $2, cookingtime = $3, difficultylevel = $4, cuisineid = $5
    WHERE recipeid = $6`

	args := []interface{}{recipe.Title, recipe.PrepTime, recipe.CookTime, recipe.Difficulty, recipe.CuisineID, recipe.ID}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan()
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrRecordNotFound
		default:
			return err
		}
	}

	return nil
}

// Add a placeholder method for deleting a specific record from the movies table.
func (m RecipeModel) Delete(id int64) error {
	query := `DELETE FROM recipes WHERE recipeid = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := m.DB.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if rowsAffected == 0 {
		return ErrRecordNotFound
	}

	return nil
}
