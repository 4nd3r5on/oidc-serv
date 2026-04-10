package api

// Handlers implements the generated [genapi.Handler] interface.
// Each field is the use case responsible for that group of operations.
type Handlers struct {
	// clients (admin operations — protected by adminKeyAuth)
	ClientCreate ClientCreator
	ClientGet    ClientGetterByID
	ClientDelete ClientDeleter

	// users (admin operations)
	Create        UserCreator
	GetByID       UserGetterByID
	GetByUsername UserGetterByUsername

	// me (self-service — auth resolved internally via Me.Auth)
	Me MeService
}
