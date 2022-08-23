package types

/*
	Responses
*/

// UploadReadsResponse Response structure for a successful read upload.
type UploadReadsResponse struct {
	Count int64 `json:"count"`
}

// GetReadsResponse Response structure for a read request.
type GetReadsResponse struct {
	Count int64  `json:"count"`
	Reads []Read `json:"reads"`
}

/*
	Requests
*/

// UploadReadsRequest Request structure for uploading reads.
type UploadReadsRequest struct {
	Reads []Read `json:"reads"`
}

// GetReadsRequest Request structure for a read request, either time based or read index based.
type GetReadsRequest struct {
	ReaderName string `json:"reader"`
	Start      int64  `json:"start"`
	End        int64  `json:"end"`
}

// DeleteReadsRequest Request structure for deletion of reads based upon read index values.
type DeleteReadsRequest struct {
	ReaderName string `json:"reader"`
	Start      *int64 `json:"start"`
	End        *int64 `json:"end"`
}
