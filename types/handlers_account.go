/*
Chronokeep Desktop - Race Scoring Software
Copyright (C) 2026 James Sentinella

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

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
	Email string `json:"email" validate:"email,required"`
}

// LoginRequest Struct used to login to an account.
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// ChangeEmailRequest Struct used to change the email on an account.
type ChangeEmailRequest struct {
	OldEmail string `json:"old_email" validate:"email,required"`
	NewEmail string `json:"new_email" validate:"email,required"`
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

