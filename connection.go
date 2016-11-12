package blueregister

import (
	"fmt"

	mgo "gopkg.in/mgo.v2"
)

func newConnection() *mgo.Session {
	session, err := mgo.Dial("127.0.0.1")
	if err != nil {
		fmt.Println(err)
		return nil
	}
	return session
}
