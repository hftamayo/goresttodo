package todo

type TodoRepositoryImpl struct {
	// Fields for database connection go here.
}

func (r *TodoRepositoryImpl) GetById(id int) (*Todo, error) {
	// Implement this method.
	// Query the database for the todo with the given id.
	// Return the todo and nil if successful, or nil and an error if something goes wrong.
}

func (r *TodoRepositoryImpl) GetAll() ([]*Todo, error) {
	// Implement this method.
	// Query the database for all todos.
	// Return the todos and nil if successful, or nil and an error if something goes wrong.
}

func (r *TodoRepositoryImpl) Create(todo *Todo) error {
	// Implement this method.
	// Insert the given todo into the database.
	// Return nil if successful, or an error if something goes wrong.
}

func (r *TodoRepositoryImpl) Update(todo *Todo) error {
	// Implement this method.
	// Update the given todo in the database.
	// Return nil if successful, or an error if something goes wrong.
}

func (r *TodoRepositoryImpl) Delete(id int) error {
	// Implement this method.
	// Delete the todo with the given id from the database.
	// Return nil if successful, or an error if something goes wrong.
}
