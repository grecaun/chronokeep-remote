package types

// GetReadersResponse Response structure for a read request.
type GetReadersResponse struct {
	Readers []Reader `json:"readers"`
}
