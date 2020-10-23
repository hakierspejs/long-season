package models

import "encoding/gob"

func init() {
	gob.Register(User{})
}
