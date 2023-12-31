package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"recipe.athif.com/internal/validator"
)

type Ingredient struct {
	// IngredientID   int64   `json:"ingredient_id"`
	IngredientName string  `json:"ingredient_name"`
	Quantity       float32 `json:"quantity"`
	Unit           string  `json:"unit"`
}
type Recipe struct {
	ID           int          `json:"id"`
	Title        string       `json:"title"`
	Instructions string       `json:"instructions"`
	PrepTime     Mins         `json:"prep_time"`
	CookTime     Mins         `json:"cook_time"`
	Difficulty   string       `json:"difficulty"`
	CuisineName  string       `json:"cuisine_name"`
	Ingredients  []Ingredient `json:"ingredients"` 
	ImageLink    string
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
    SELECT r.recipeid, r.recipename, r.instructions, r.preparationtime, r.cookingtime, r.difficultylevel, c.cuisinename, i.ingredientname, ri.quantity, ri.unit, img.imagelink
    FROM recipes r
    INNER JOIN cuisine c ON r.cuisineid = c.cuisineid
    INNER JOIN recipeingredients ri ON r.recipeid = ri.recipeid
    INNER JOIN ingredients i ON ri.ingredientid = i.ingredientid
	INNER JOIN recipe_images img ON r.recipeid = img.recipeid
    WHERE r.recipeid = $1
    `

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := r.DB.QueryContext(ctx, query, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var recipe Recipe
	for rows.Next() {
		var ingredient Ingredient
		err = rows.Scan(&recipe.ID, &recipe.Title, &recipe.Instructions, &recipe.PrepTime, &recipe.CookTime, &recipe.Difficulty, &recipe.CuisineName, &ingredient.IngredientName, &ingredient.Quantity, &ingredient.Unit, &recipe.ImageLink)
		if err != nil {
			return nil, err
		}
		recipe.Ingredients = append(recipe.Ingredients, ingredient)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	if len(recipe.Ingredients) == 0 {
		return nil, ErrRecordNotFound
	}

	return &recipe, nil
}

func (r RecipeModel) Update(recipe *Recipe) error {
	query := `
	UPDATE recipes
	SET recipename = $1, instructions = $2, preparationtime = $3, cookingtime = $4, difficultylevel = $5, cuisinename = (SELECT cuisinename FROM cuisine WHERE cuisinename = $6)
	WHERE recipeid = $7`

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

func (r RecipeModel) GetAll(title string, cuisineID int, filters Filters) ([]*Recipe, Metadata, error) {

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
                        WHEN r.difficultylevel = 'Medium' THEN 2
                        WHEN r.difficultylevel = 'Advanced' THEN 3
                      END`
	}

	query := fmt.Sprintf(`
    SELECT  r.recipeid, r.recipename, r.instructions, r.preparationtime, r.cookingtime, r.difficultylevel, c.cuisinename, i.ingredientname, ri.quantity, ri.unit, img.imageLink
    FROM recipes r
    INNER JOIN cuisine c ON r.cuisineid = c.cuisineid
    INNER JOIN recipeingredients ri ON r.recipeid = ri.recipeid
    INNER JOIN ingredients i ON ri.ingredientid = i.ingredientid
	INNER JOIN recipe_images img ON r.recipeid = img.recipeid
    WHERE (LOWER(r.recipename) LIKE LOWER($1) OR $1 = '')
    AND (r.cuisineid = $2 OR $2 = 0)
    ORDER BY %s %s, r.recipeid ASC`,
	//LIMIT %d OFFSET %d`, 
	sortColumn, filters.sortDirection())//, filters.limit(), filters.offset())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := r.DB.QueryContext(ctx, query, "%"+title+"%", cuisineID)
	if err != nil {
		return nil, Metadata{}, err
	}
	defer rows.Close()

	recipes := make(map[int]*Recipe)
	//totalRecords := 0

	for rows.Next() {
		var recipe Recipe
		var ingredient Ingredient
		err := rows.Scan(
			// &totalRecords,
			&recipe.ID,
			&recipe.Title,
			&recipe.Instructions,
			&recipe.PrepTime,
			&recipe.CookTime,
			&recipe.Difficulty,
			&recipe.CuisineName,
			&ingredient.IngredientName,
			&ingredient.Quantity,
			&ingredient.Unit,
			&recipe.ImageLink,
		)
		if err != nil {
			return nil, Metadata{}, err
		}

		if _, exists := recipes[int(recipe.ID)]; !exists {
			recipe.Ingredients = []Ingredient{ingredient}
			recipes[int(recipe.ID)] = &recipe
		} else {
			recipes[int(recipe.ID)].Ingredients = append(recipes[int(recipe.ID)].Ingredients, ingredient)
		}
	}

	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	result := []*Recipe{}
	for _, recipe := range recipes {
		result = append(result, recipe)
	}

	countQuery := `
    SELECT COUNT(DISTINCT recipeid)
    FROM recipes
    WHERE (LOWER(recipename) LIKE LOWER($1) OR $1 = '')
    AND (cuisineid = $2 OR $2 = 0)
    `

	var totalRecords int
	err = r.DB.QueryRowContext(ctx, countQuery, "%"+title+"%", cuisineID).Scan(&totalRecords)
	if err != nil {
		return nil, Metadata{}, err
	}

	metadata := calculateMetadata(totalRecords, filters.Page)//, filters.PageSize)
	return result, metadata, nil
}

func (m *RecipeModel) Search(ingredients []string) ([]*Recipe, error) {
	// Log the ingredients.
	//log.Printf("Ingredients: %v\n", ingredients)

	//Return an error if the ingredients slice is empty.
	if len(ingredients) == 0 {
		return nil, errors.New("at least one ingredient must be provided")
	}

	// Generate a placeholder for each ingredient in the slice.
	// Convert each ingredient to lowercase.
	for i, ingredient := range ingredients {
		ingredients[i] = strings.ToLower(ingredient)
	}

	// Generate a placeholder for each ingredient in the slice.
	placeholders := ""
	for i := range ingredients {
		if i > 0 {
			placeholders += ","
		}
		placeholders += fmt.Sprintf("$%d", i+1)
	}

	query := fmt.Sprintf(`
    SELECT rv2.recipeid, rv2.recipename, rv2.instructions, rv2.preparationtime, rv2.cookingtime, rv2.difficultylevel, c.cuisinename, rv2.ingredientname, rv2.quantity, rv2.unit
    FROM recipe_view rv2
    INNER JOIN cuisine c ON rv2.cuisineid = c.cuisineid
    INNER JOIN (
        SELECT recipeid
        FROM recipe_view rv1
        WHERE LOWER(ingredientname) IN (%s)
        GROUP BY recipeid
        HAVING COUNT(DISTINCT LOWER(ingredientname)) = %d
    ) r ON rv2.recipeid = r.recipeid::bigint
`, placeholders, len(ingredients))
	// Log the generated query.
	//log.Printf("Query: %s\n", query)

	args := make([]interface{}, len(ingredients))
	for i, ingredient := range ingredients {
		args[i] = ingredient
	}

	// Log the arguments.
//	log.Printf("Arguments: %v\n", args)

	// Pass the args slice to the DB.Query method.
	rows, err := m.DB.Query(query, args...)
	if err != nil {
		// Log the error.
		log.Printf("Error executing query: %v\n", err)
		return nil, err
	}
	defer rows.Close()

	recipes := make(map[int]*Recipe)

	for rows.Next() {
		var recipe Recipe
		var ingredient Ingredient

		err := rows.Scan(
			&recipe.ID,
			&recipe.Title,
			&recipe.Instructions,
			&recipe.PrepTime,
			&recipe.CookTime,
			&recipe.Difficulty,
			&recipe.CuisineName,
			&ingredient.IngredientName,
			&ingredient.Quantity,
			&ingredient.Unit,
		)
		if err != nil {
			return nil, err
		}

		if existingRecipe, ok := recipes[recipe.ID]; ok {
			existingRecipe.Ingredients = append(existingRecipe.Ingredients, ingredient)
		} else {
			recipe.Ingredients = []Ingredient{ingredient}
			recipes[recipe.ID] = &recipe
		}
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	// Convert the map to a slice.
	result := make([]*Recipe, 0, len(recipes))
	for _, recipe := range recipes {
		result = append(result, recipe)
	}
	return result, nil
}

func (m *RecipeModel) ListAllIngredients() ([]string, error) {
    query := `SELECT DISTINCT ingredientname FROM ingredients`

    // Execute the query.
    rows, err := m.DB.QueryContext(context.Background(), query)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    ingredients := make([]string, 0)
    for rows.Next() {
        var ingredient string
        if err := rows.Scan(&ingredient); err != nil {
            return nil, err
        }
        ingredients = append(ingredients, ingredient)
    }

    // Check for errors from iterating over rows.
    if err := rows.Err(); err != nil {
        return nil, err
    }

    return ingredients, nil
}