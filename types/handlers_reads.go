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

// UploadReadsResponse Response structure for a successful read upload.
type UploadReadsResponse struct {
	Count int64 `json:"count"`
}

// GetReadsResponse Response structure for a read request.
type GetReadsResponse struct {
	Count int64         `json:"count"`
	Reads []Read        `json:"reads"`
	Note  *Notification `json:"notification"`
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

