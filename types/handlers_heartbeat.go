/*
Chronokeep Remote - Remote Tag Read API
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

type GetHeartbeatsResponse struct {
	Heartbeats []ReaderHeartbeat `json:"heartbeats"`
}

/*
	Requests

	SaveHeartbeat takes a Heartbeat as the whole request body.
*/

// GetHeartbeatsRequest optionally limits GET /heartbeats to a single reader.
// An empty body or empty reader returns every reader on the account.
type GetHeartbeatsRequest struct {
	ReaderName string `json:"reader"`
}
