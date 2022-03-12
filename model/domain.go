package model

import (
	"time"
)

type Domain struct {
	Id        string    `bson:"_id,omitempty"`
	FQDN      string    `bson:"fqdn"`
	IP        string    `bson:"ip"`
	UserId    string    `bson:"user_id"`
	UserName  string    `bson:"user_name"`
	CreatedAt time.Time `bson:"created_at"`
	DeleteAt  time.Time `bson:"delete_at"`
	BasicAuth bool      `bson:"basic_auth"`
	FullSsl   bool      `bson:"full_ssl"`
}
