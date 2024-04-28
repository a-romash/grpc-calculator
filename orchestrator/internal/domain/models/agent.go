package models

import "time"

type Agent struct {
	ID            int       `db:"id"`
	LastHeartbeat time.Time `db:"last_heartbeat"`
	Status        string    `db:"status"`
}
