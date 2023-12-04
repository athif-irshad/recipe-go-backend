# Recipe Management API

This is a comprehensive RESTful API designed for managing culinary recipes. The API is implemented in Go, a statically typed, compiled language that ensures efficiency and safety. The data is stored in PostgreSQL, a powerful, open source object-relational database system.

## Key Features

- **CRUD Operations**: You can create, read, update, and delete recipes.
- **Search Functionality**: You can search for recipes based on ingredients.
- **Ingredient Listing**: You can list all ingredients used in the recipes.

## Getting Started

### Prerequisites

- Go 1.16 or later: The Go programming language is required to run this project.
- PostgreSQL 13 or later: This is the database system used for data storage.

### Installation

1. Clone the repository
2. Navigate to the project directory
3. Install the dependencies: `go mod download`
4. Set up your database and update the connection string in the `main.go` file.

### Running the Application

1. Start the server: `go run main.go`
2. ## API Endpoints

- `POST /v1/recipes`: Create a new recipe.
- `GET /v1/recipes/{id}`: Retrieve a recipe by its ID.
- `PUT /v1/recipes/{id}`: Update a recipe by its ID.
- `DELETE /v1/recipes/{id}`: Delete a recipe by its ID.
- `GET /v1/recipes`: List all recipes.
- `GET /v1/recipes/search`: Search for recipes based on ingredients.
- `GET /v1/ingredients`: List all ingredients used in the recipes.

## Contributing

We welcome contributions from the community. If you wish to contribute, please create a pull request. For major changes, please open an issue first to discuss what you would like to change.

## License

This project is licensed under the MIT License - see the `LICENSE` file for details.