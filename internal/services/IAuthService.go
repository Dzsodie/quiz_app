package services

type IAuthService interface {
	RegisterUser(username, password string) error

	AuthenticateUser(username, password string) error
}
