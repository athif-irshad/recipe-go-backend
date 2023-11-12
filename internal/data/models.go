package data

import (
	"database/sql"
	"errors"
)


var (
	ErrRecordNotFound = errors.New("record not found")
)


type Models struct {
	Recipes RecipeModel
}


func NewModels(db *sql.DB) Models {
	return Models{
		Recipes: RecipeModel{DB: db},
	}
}
