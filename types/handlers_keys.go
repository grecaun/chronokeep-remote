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

// ModifyKeyResponse Struct used to respond to a Add/Update Key Request.
type ModifyKeyResponse struct {
	Key Key `json:"key"`
}

// GetKeysResponse Struct used to respond to the requets for account keys.
type GetKeysResponse struct {
	Keys []Key `json:"keys"`
}

/*
	Requests
*/

// DeleteKeyRequest Struct used for the Delete Key request.
type DeleteKeyRequest struct {
	Key string `json:"key"`
}

// AddKeyRequest Struct used for the Add Key request.
type AddKeyRequest struct {
	Email *string    `json:"email"`
	Key   RequestKey `json:"key"`
}

// UpdateKeyRequest Struct used for the Update Key request.
type UpdateKeyRequest struct {
	Key RequestKey `json:"key"`
}

// GetKeysRequest Struct used for the Get Keys request.
type GetKeysRequest struct {
	Email *string `json:"email"`
}

