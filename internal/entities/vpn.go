package entities

import "time"

type VPNEU struct {
	Id           string    `bson:"_id,omitempty"`
	UserName     string    `bson:"user_name"`
	UserEmail    string    `bson:"user_email"`
	UserId       string    `bson:"user_id"`
	CreatedAt    time.Time `bson:"created_at"`
	DeactivateAt time.Time `bson:"deactivate_at"`
	Active       bool      `bson:"active"`
}
