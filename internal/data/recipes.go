package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"recipe.athif.com/internal/validator"
)

type Recipe struct {
	ID           int64  `json:"id"`
	Title        string `json:"title"`
	Instructions string `json:"instructions"`
	PrepTime     Mins   `json:"prep_time"`
	CookTime     Mins   `json:"cook_time"`
	Difficulty   string `json:"difficulty"`
	CuisineName  string `json:"cuisine_name"`
}

func ValidateRecipe(v *validator.Validator, recipe *Recipe) {
	v.Check(recipe.Title != "", "title", "must be provided")
	v.Check(len(recipe.Title) <= 500, "title", "must not be more than 500 bytes long")
	v.Check(recipe.PrepTime != 0, "preparation_time", "must be provided")
	v.Check(recipe.PrepTime > 0, "preparation_time", "must be a positive integer")
	v.Check(recipe.CookTime != 0, "cooking_time", "must be provided")
	v.Check(recipe.CookTime > 0, "cooking_time", "must be a positive integer")
	v.Check(recipe.CuisineName != "", "cuisine_name", "must be provided")
	v.Check(recipe.Difficulty != "", "difficulty", "must be provided")
	v.Check(recipe.Difficulty != "", "instructions", "must be provided")
}

type RecipeModel struct {
	DB *sql.DB
}

func (r RecipeModel) Insert(recipe *Recipe) error {
	query := `
        INSERT INTO recipes (recipename, instructions, preparationtime, cookingtime, difficultylevel, cuisinename)
        SELECT $1, $2, $3, $4, $5, cuisinename
        FROM cuisine
        WHERE cuisinename = $6
        RETURNING recipeid
    `

	args := []interface{}{recipe.Title, recipe.Instructions, recipe.PrepTime, recipe.CookTime, recipe.Difficulty, recipe.CuisineName}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return r.DB.QueryRowContext(ctx, query, args...).Scan(&recipe.ID)
}

func (r RecipeModel) Get(id int64) (*Recipe, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	query := `
	SELECT r.recipeid, r.recipename, r.instructions, r.preparationtime, r.cookingtime, r.difficultylevel, c.cuisinename
	FROM recipes r
	INNER JOIN cuisine c ON r.cuisineid = c.cuisineid
	WHERE r.recipeid = $1
`

	recipe := &Recipe{}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	err := r.DB.QueryRowContext(ctx, query, id).Scan(&recipe.ID, &recipe.Title, &recipe.Instructions, &recipe.PrepTime, &recipe.CookTime, &recipe.Difficulty, &recipe.CuisineName)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrRecordNotFound
		} else {
			return nil, err
		}
	}

	return recipe, nil
}

func (r RecipeModel) Update(recipe *Recipe) error {
	query := `
	UPDATE recipes
	SET recipename = $1, instructions = $2, preparationtime = $3, cookingtime = $4, difficultylevel = $5, cuisinename = (SELECT cuisinename FROM cuisine WHERE cuisinename = $6)
	WHERE recipeid = $7
`

	args := []interface{}{recipe.Title, recipe.Instructions, recipe.PrepTime, recipe.CookTime, recipe.Difficulty, recipe.CuisineName, recipe.ID}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := r.DB.QueryRowContext(ctx, query, args...).Scan()
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

func (r RecipeModel) Delete(id int64) error {
	query := `DELETE FROM recipes WHERE recipeid = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := r.DB.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return ErrRecordNotFound
	}

	return nil
}

func (r RecipeModel) GetAll(title string, cuisineID int, filters Filters) ([]*Recipe, error) {

	sortColumn := filters.sortColumn()
	if sortColumn == "cuisinename" {
		sortColumn = "c.cuisinename"
	} else if sortColumn == "title" {
		sortColumn = "r.recipename"
	} else if sortColumn == "id" {
		sortColumn = "r.recipeid"
	} else if sortColumn == "difficulty" {
        sortColumn = `CASE
                        WHEN r.difficultylevel = 'Easy' THEN 1
                        WHEN r.difficultylevel = 'Intermediate' THEN 2
                        WHEN r.difficultylevel = 'Advanced' THEN 3
                      END`
    }

	query := fmt.Sprintf(`
    SELECT r.recipeid, r.recipename, r.instructions, r.preparationtime, r.cookingtime, r.difficultylevel, c.cuisinename
    FROM recipes r
    INNER JOIN cuisine c ON r.cuisineid = c.cuisineid
    WHERE (LOWER(r.recipename) LIKE LOWER($1) OR $1 = '')
    AND (r.cuisineid = $2 OR $2 = 0)
    ORDER BY %s %s, r.recipeid ASC`, sortColumn, filters.sortDirection())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := r.DB.QueryContext(ctx, query, "%"+title+"%", cuisineID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	recipes := []*Recipe{}
	for rows.Next() {
		var recipe Recipe
		err := rows.Scan(
			&recipe.ID,
			&recipe.Title,
			&recipe.Instructions,
			&recipe.PrepTime,
			&recipe.CookTime,
			&recipe.Difficulty,
			&recipe.CuisineName,
		)
		if err != nil {
			return nil, err
		}
		recipes = append(recipes, &recipe)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return recipes, nil
}
