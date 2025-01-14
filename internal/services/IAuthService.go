package services

// IAuthService defines the interface for authentication-related operations.
type IAuthService interface {
	// RegisterUser registers a new user with the given username and password.
	// Returns an error if the user already exists.
	RegisterUser(username, password string) error

	// AuthenticateUser validates the username and password.
	// Returns an error if the credentials are invalid.
	AuthenticateUser(username, password string) error
}
