package blueregister

import (
	"fmt"

	"sync"

	mgo "gopkg.in/mgo.v2"
)

var (
	baseConnection *mgo.Session
	sessionSync    sync.Mutex
)

func init() {
	sessionSync = sync.Mutex{}
}

func newConnection() *mgo.Session {
	sessionSync.Lock()
	defer sessionSync.Unlock()

	if baseConnection == nil {
		var err error
		baseConnection, err = mgo.Dial("127.0.0.1")
		if err != nil {
			fmt.Println(err)
			return nil
		}
		return baseConnection.Copy()
	}
	return baseConnection.Copy()
	// session, err := mgo.Dial("127.0.0.1")
	// if err != nil {
	// 	fmt.Println(err)
	// 	return nil
	// }
	// return session
}
