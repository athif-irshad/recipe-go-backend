package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (app *application) routes() http.Handler {
    router := httprouter.New()
    router.NotFound = http.HandlerFunc(app.notFoundResponse)
    router.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowedResponse)
    router.HandlerFunc(http.MethodGet, "/v1/healthcheck", app.healthcheckHandler)
    router.HandlerFunc(http.MethodGet, "/v1/recipes", app.listRecipeHandler)
    router.HandlerFunc(http.MethodPost, "/v1/recipes", app.createRecipeHandler)
    router.HandlerFunc(http.MethodGet, "/v1/search", app.searchRecipesHandler)
    router.HandlerFunc(http.MethodGet, "/v1/listingredients", app.listAllIngredientsHandler)
    router.HandlerFunc(http.MethodGet, "/v1/recipes/:id", app.showRecipeHandler)
    router.HandlerFunc(http.MethodPut, "/v1/recipes/:id", app.updateRecipeHandler)
    router.HandlerFunc(http.MethodDelete, "/v1/recipes/:id", app.deleteRecipeHandler)
    return app.enableCORS(router)
}