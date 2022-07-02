package types

/*
	Responses
*/

// UploadReadsResponse Response structure for a successful read upload.
type UploadReadsResponse struct {
	Count       int      `json:"count"`
	Successfull []string `json:"successfull"`
}

// GetReadsResponse Response structure for a read request.
type GetReadsResponse struct {
	Count int    `json:"count"`
	Reads []Read `json:"reads"`
}

// DeleteReadsResponse Response structure for deletion of reads. Informs the user of the number of deleted reads.
type DeleteReadsResponse struct {
	Count int `json:"count"`
}

/*
	Requests
*/

// UploadReadsRequest Request structure for uploading reads.
type UploadReadsRequest struct {
	Key   string `json:"key"`
	Count int    `json:"count"`
	Reads []Read `json:"reads"`
}

// GetReadsRequest Request structure for a read request, either time based or read index based.
type GetReadsRequest struct {
	Key        string `json:"key"`
	ReaderName string `json:"reader"`
	Start      int64  `json:"start"`
	End        int64  `json:"end"`
}

// DeleteReadsRequest Request structure for deletion of reads based upon read index values.
type DeleteReadsRequest struct {
	Key        string `json:"key"`
	ReaderName string `json:"reader"`
	Start      *int64 `json:"start"`
	End        *int64 `json:"end"`
}
