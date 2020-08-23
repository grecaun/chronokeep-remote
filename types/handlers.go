package types

// GeneralRequest A generalized request struct when only a key is required for the call.
type GeneralRequest struct {
	Key string `json:"key"`
}

// GetReadersResponse Struct used for the response of Get Readers request.
type GetReadersResponse struct {
	Count   int      `json:"count"`
	Readers []string `json:"readers"`
}

// UploadReadsRequest Request structure for uploading reads.
type UploadReadsRequest struct {
	Key       string `json:"key"`
	Overwrite bool   `json:"overwrite"`
	Count     int    `json:"count"`
	Reader    string `json:"reader"`
	Reads     []Read `json:"reads"`
}

// UploadReadsResponse Response structure for a successful read upload.
type UploadReadsResponse struct {
	Count       int      `json:"count"`
	Successfull []string `json:"successfull"`
}

// GetReadsRequest Request structure for a read request, either time based or read index based.
type GetReadsRequest struct {
	Key    string `json:"key"`
	Start  string `json:"start"`
	End    string `json:"end"`
	Reader string `json:"reader"`
}

// GetReadsResponse Response structure for a read request.
type GetReadsResponse struct {
	Count int    `json:"count"`
	Reads []Read `json:"reads"`
}

// DeleteReadsRangeRequest Request structure for deletion of reads based upon read index values.
type DeleteReadsRangeRequest struct {
	Key    string `json:"key"`
	Reader string `json:"reader"`
	Start  string `json:"start"`
	End    string `json:"end"`
}

// DeleteReadsRequest Request structure for deletion of reads based upon an array of read index values.
type DeleteReadsRequest struct {
	Key    string   `json:"key"`
	Reader string   `json:"reader"`
	Reads  []string `json:"reads"`
}

// DeleteReadsResponse Response structure for deletion of reads. Informs the user of the number of deleted reads.
type DeleteReadsResponse struct {
	Count int `json:"count"`
}

// AccountRequest Request for adding, deleting, or updating an account.
type AccountRequest struct {
	Key     string  `json:"key"`
	Account Account `json:"account"`
}

// AccountResponse Response for adding, deleting, or updating an account.
type AccountResponse struct {
	Account Account `json:"account"`
}

// GetAccountRequest Request for getting an account information
type GetAccountRequest struct {
	Key     string `json:"key"`
	Account string `json:"account"`
}

// GetAllAccountsResponse Response to a get all accounts request.
type GetAllAccountsResponse struct {
	Accounts []Account `json:"accounts"`
}
