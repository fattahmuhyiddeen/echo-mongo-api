package model

import (
	"gopkg.in/mgo.v2/bson"
)

type (
	User struct {
		ID        bson.ObjectId `json:"id" bson:"_id,omitempty"`
		Email     string        `json:"email" bson:"email"`
		Name      string        `json:"name" bson:"name"`
		Password  string        `json:"password,omitempty" bson:"password"`
		Token     string        `json:"token,omitempty" bson:"-"`
		Followers []string      `json:"followers,omitempty" bson:"followers,omitempty"`
	}
)
