package blueregister

import (
	"os"

	"time"

	"gopkg.in/mgo.v2/bson"
)

var appPrefix string

func CheckIn(prefix string) {
	if len(appPrefix) > 0 {
		return
	}
	appPrefix = prefix
	conn := newConnection()
	defer conn.Close()
	wd, _ := os.Getwd()
	conn.DB("blue_apps").C("state").Upsert(
		bson.M{"_id": prefix},
		bson.M{
			"_id":        prefix,
			"pid":        os.Getpid(),
			"timestamp":  time.Now(),
			"is_running": true,
			"working":    wd,
		})
}

func CheckOut() {
	if len(appPrefix) == 0 {
		return
	}

	conn := newConnection()
	defer conn.Close()

	conn.DB("blue_apps").C("state").Update(
		bson.M{"_id": appPrefix},
		bson.M{"$set": bson.M{
			"timestamp":  time.Now(),
			"is_running": false,
		}})
}

func Set(key string, value interface{}) {
	if len(appPrefix) == 0 {
		return
	}

	conn := newConnection()
	defer conn.Close()

	conn.DB("blue_apps").C("state").Update(
		bson.M{"_id": appPrefix},
		bson.M{"$set": bson.M{
			key: value,
		}})
}
