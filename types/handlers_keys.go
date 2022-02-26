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

// GeneralRequest A generalized request struct when only a key is required for the call.
type GeneralKeyRequest struct {
	Key string `json:"key"`
}

// AddKeyRequest Struct used for the Add Key request.
type AddKeyRequest struct {
	Identifier string     `json:"identifier"`
	Key        RequestKey `json:"key"`
}

// UpdateKeyRequest Struct used for the Update Key request.
type UpdateKeyRequest struct {
	Key RequestKey `json:"key"`
}
