package postgres

import (
	"chronokeep/remote/types"
	"context"
	"errors"
	"fmt"
	"time"
)

func (p *Postgres) GetNotifications(key string) (*types.Notification, error) {
	db, err := p.GetDB()
	if err != nil {
		return nil, err
	}
	ctx, cancelfunc := context.WithTimeout(context.Background(), time.Second*5)
	defer cancelfunc()
	res, err := db.Query(
		ctx,
		"SELECT notification_id, notification_type, notification_when "+
			"FROM (SELECT key_value, MAX(notification_when) AS max_when FROM notification GROUP BY key_value) AS b INNER JOIN notification AS n ON b.max_when=n.notification_when AND b.key_value=n.key_value "+
			"WHERE n.key_value=$1 AND n.notification_when>$2;",
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

func (p *Postgres) SaveNotification(notification *types.RequestNotification, key string) error {
	db, err := p.GetDB()
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
	res, err := db.Exec(
		ctx,
		"INSERT INTO notification(notification_type, notification_when, key_value) VALUES($1, $2, $3) ON CONFLICT DO NOTHING;",
		notification.Type,
		when.Unix(),
		key,
	)
	if err != nil {
		return fmt.Errorf("unable to add notification: %v", err)
	}
	if res.RowsAffected() < 1 {
		return errors.New("insert appears to be unsuccessful")
	}
	return nil
}
