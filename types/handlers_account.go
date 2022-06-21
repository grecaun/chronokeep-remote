package types

/*
	Responses
*/

// GetAccountResponse Struct used for the response of the Get Account Request.
type GetAccountResponse struct {
	Account Account `json:"account"`
	Keys    []Key   `json:"keys"`
}

// GetAllAccountsResponse Struct used to get all of the accounts.
type GetAllAccountsResponse struct {
	Accounts []Account `json:"accounts"`
}

// ModifyAccountResponse Struct used to respond to add/update account requests.
type ModifyAccountResponse struct {
	Account Account `json:"account"`
}

/*
	Requests
*/

// GetAccountRequest Struct used to request information about a specific account.
type GetAccountRequest struct {
	Email *string `json:"email"`
}

// ModifyAccountRequest Struct used to request to update an account.
type UpdateAccountRequest struct {
	Account Account `json:"account"`
}

// AddAccountRequest Struct used to add an account.
type AddAccountRequest struct {
	Account  Account `json:"account"`
	Password string  `json:"password"`
}

// DeleteAccountRequest Struct used to delete an account.
type DeleteAccountRequest struct {
	Email string `json:"email"`
}

// LoginRequest Struct used to login to an account.
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// ChangeEmailRequest Struct used to change the email on an account.
type ChangeEmailRequest struct {
	OldEmail string `json:"old_email"`
	NewEmail string `json:"new_email"`
}

// RefreshTokenRequest Struct used to request a fresh token.
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// ChangePasswordRequest Struct used to change the password on an account.
type ChangePasswordRequest struct {
	Email       string `json:"email"`
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}
