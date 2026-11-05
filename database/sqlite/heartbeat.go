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

package sqlite

import (
	"chronokeep/remote/database"
	"chronokeep/remote/types"
	"context"
	"fmt"
	"time"
)

// SaveHeartbeat stores the heartbeat as the latest one for the given key, replacing any previous one.
func (s *SQLite) SaveHeartbeat(key string, heartbeat types.Heartbeat) error {
	db, err := s.GetDB()
	if err != nil {
		return err
	}
	ctx, cancelfunc := context.WithTimeout(context.Background(), time.Second*5)
	defer cancelfunc()
	res, err := db.ExecContext(
		ctx,
		"INSERT INTO heartbeat(key_value, heartbeat_device, heartbeat_location, heartbeat_now, heartbeat_uptime, heartbeat_battery, heartbeat_charging, heartbeat_hash, heartbeat_upload_errors, heartbeat_readers_connected, heartbeat_readers_total, heartbeat_reading, heartbeat_chip_type, heartbeat_received) "+
			"VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)"+
			" ON CONFLICT (key_value) DO UPDATE SET heartbeat_device=EXCLUDED.heartbeat_device, heartbeat_location=EXCLUDED.heartbeat_location, heartbeat_now=EXCLUDED.heartbeat_now, heartbeat_uptime=EXCLUDED.heartbeat_uptime, heartbeat_battery=EXCLUDED.heartbeat_battery, heartbeat_charging=EXCLUDED.heartbeat_charging, heartbeat_hash=EXCLUDED.heartbeat_hash, heartbeat_upload_errors=EXCLUDED.heartbeat_upload_errors, heartbeat_readers_connected=EXCLUDED.heartbeat_readers_connected, heartbeat_readers_total=EXCLUDED.heartbeat_readers_total, heartbeat_reading=EXCLUDED.heartbeat_reading, heartbeat_chip_type=EXCLUDED.heartbeat_chip_type, heartbeat_received=EXCLUDED.heartbeat_received;",
		key,
		heartbeat.Device,
		heartbeat.Location,
		heartbeat.Now,
		heartbeat.Uptime,
		heartbeat.Battery,
		heartbeat.Charging,
		heartbeat.Hash,
		heartbeat.UploadErrors,
		heartbeat.ReadersConnected,
		heartbeat.ReadersTotal,
		heartbeat.Reading,
		heartbeat.ChipType,
		time.Now().UnixMilli(),
	)
	if err != nil {
		return fmt.Errorf("unable to save heartbeat: %v", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("error checking rows affected: %v", err)
	}
	if rows < 1 {
		return fmt.Errorf("heartbeat not saved")
	}
	return nil
}

// GetHeartbeats gets the latest heartbeat for each non-deleted key on the account,
// skipping readers whose latest heartbeat is older than database.HeartbeatMaxAge.
// Only the latest heartbeat per key is ever stored (SaveHeartbeat replaces it).
// When reader_name isn't empty, only that reader's heartbeat is returned, with no age limit.
func (s *SQLite) GetHeartbeats(account int64, reader_name string) ([]types.ReaderHeartbeat, error) {
	db, err := s.GetDB()
	if err != nil {
		return nil, err
	}
	ctx, cancelfunc := context.WithTimeout(context.Background(), time.Second*5)
	defer cancelfunc()
	query := "SELECT a.key_name, h.heartbeat_device, h.heartbeat_location, h.heartbeat_now, h.heartbeat_uptime, h.heartbeat_battery, h.heartbeat_charging, h.heartbeat_hash, h.heartbeat_upload_errors, h.heartbeat_readers_connected, h.heartbeat_readers_total, h.heartbeat_reading, h.heartbeat_chip_type, h.heartbeat_received " +
		"FROM heartbeat AS h JOIN api_key AS a ON h.key_value=a.key_value " +
		"WHERE a.account_id=? AND a.key_deleted=FALSE"
	args := []any{account}
	if reader_name != "" {
		// A single named reader gets its latest heartbeat however old it is.
		query += " AND LOWER(a.key_name)=LOWER(?)"
		args = append(args, reader_name)
	} else {
		query += " AND h.heartbeat_received>?"
		args = append(args, time.Now().Add(-database.HeartbeatMaxAge).UnixMilli())
	}
	query += " ORDER BY a.key_name;"
	res, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("error retrieving heartbeats: %v", err)
	}
	defer res.Close()
	out := make([]types.ReaderHeartbeat, 0)
	for res.Next() {
		var hb types.ReaderHeartbeat
		var received int64
		err := res.Scan(
			&hb.ReaderName,
			&hb.Heartbeat.Device,
			&hb.Heartbeat.Location,
			&hb.Heartbeat.Now,
			&hb.Heartbeat.Uptime,
			&hb.Heartbeat.Battery,
			&hb.Heartbeat.Charging,
			&hb.Heartbeat.Hash,
			&hb.Heartbeat.UploadErrors,
			&hb.Heartbeat.ReadersConnected,
			&hb.Heartbeat.ReadersTotal,
			&hb.Heartbeat.Reading,
			&hb.Heartbeat.ChipType,
			&received,
		)
		if err != nil {
			return nil, fmt.Errorf("error getting heartbeat: %v", err)
		}
		hb.Received = time.UnixMilli(received)
		out = append(out, hb)
	}
	if err := res.Err(); err != nil {
		return nil, fmt.Errorf("error reading heartbeats: %v", err)
	}
	return out, nil
}
