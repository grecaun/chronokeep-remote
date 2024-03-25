package sqlite

import (
	"chronokeep/remote/types"
	"context"
	"errors"
	"fmt"
	"time"
)

func (s *SQLite) GetNotifications(key string) (*types.Notification, error) {
	db, err := s.GetDB()
	if err != nil {
		return nil, err
	}
	ctx, cancelfunc := context.WithTimeout(context.Background(), time.Second*5)
	defer cancelfunc()
	res, err := db.QueryContext(
		ctx,
		"SELECT n.notification_id, n.notification_type, n.notification_when "+
			"FROM (SELECT key_value, MAX(notification_when) AS max_when FROM notification GROUP BY key_value) AS b JOIN notification AS n ON b.max_when=n.notification_when AND b.key_value=n.key_value "+
			"WHERE n.key_value=? AND n.notification_when>?;",
		key,
		time.Now().Add(time.Minute*-5).Unix(),
	)
	if err != nil {
		return nil, fmt.Errorf("error retrieving notification: %v", err)
	}
	defer res.Close()
	var out types.Notification
	var when int64
	if res.Next() {
		err := res.Scan(
			&out.Identifier,
			&out.Type,
			&when,
		)
		if err != nil {
			return nil, fmt.Errorf("error getting notifications: %v", err)
		}
	} else {
		return nil, nil
	}
	out.When = time.Unix(when, 0)
	return &out, nil
}

func (s *SQLite) SaveNotification(notification *types.RequestNotification, key string) error {
	db, err := s.GetDB()
	if err != nil {
		return err
	}
	valid := false
	switch notification.Type {
	case "UPS_DISCONNECTED", "UPS_CONNECTED", "UPS_ON_BATTERY", "UPS_LOW_BATTERY", "UPS_ONLINE", "SHUTTING_DOWN", "RESTARTING", "HIGH_TEMP", "MAX_TEMP":
		valid = true
	}
	if !valid {
		return fmt.Errorf("%v is not a valid type", notification.Type)
	}
	when, err := time.Parse(time.RFC3339, notification.When)
	if err != nil {
		return fmt.Errorf("unable to parse time value: %v", err)
	}
	ctx, cancelfunc := context.WithTimeout(context.Background(), time.Second*5)
	defer cancelfunc()
	res, err := db.ExecContext(
		ctx,
		"INSERT INTO notification(notification_type, notification_when, key_value) VALUES(?, ?, ?);",
		notification.Type,
		when.Unix(),
		key,
	)
	if err != nil {
		return fmt.Errorf("unable to add notification: %v", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("error checking rows affected: %v", err)
	}
	if rows < 1 {
		return errors.New("insert appears to be unsuccessful")
	}
	return nil
}
