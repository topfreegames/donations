package errors

import "fmt"

//AuthenticationProviderNotSupported happens when a query should return a document but it's not found
type AuthenticationProviderNotSupported struct {
	Provider string
}

//NewAuthenticationProviderNotSupported creates a new error
func NewAuthenticationProviderNotSupported(provider string) *AuthenticationProviderNotSupported {
	return &AuthenticationProviderNotSupported{
		Provider: provider,
	}
}

//Error string
func (err AuthenticationProviderNotSupported) Error() string {
	return fmt.Sprintf("Provider %s is not supported.", err.Provider)
}

//AuthenticationFailed happens when an authentication provider does not recognize the access token
type AuthenticationFailed struct {
	Provider, UserID, Token string
}

//NewAuthenticationFailed creates a new error
func NewAuthenticationFailed(provider, userID, token string) *AuthenticationFailed {
	return &AuthenticationFailed{
		Provider: provider,
		UserID:   userID,
		Token:    token,
	}
}

//Error string
func (err AuthenticationFailed) Error() string {
	return fmt.Sprintf(
		"Authentication with provider %s for user %s failed (Access Token: %s).",
		err.Provider,
		err.UserID,
		err.Token,
	)
}
