package blueregister

import (
	"os"
	"sync"

	"time"

	"fmt"
	"net/http"

	"github.com/sfi2k7/blueutil"
	"gopkg.in/mgo.v2/bson"
)

var appPrefix string
var hitFailing bool
var hitChannel chan string
var logErrorChannel chan string
var logLock sync.Mutex
var logSource string

func init() {
	hitChannel = make(chan string, 1000)
	logLock = sync.Mutex{}
	go internalHit()
}

func CheckIn(prefix string) {
	if len(appPrefix) > 0 {
		return
	}
	appPrefix = prefix
	conn := newConnection()
	if conn == nil {
		return
	}
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
	if conn == nil {
		return
	}
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
	if conn == nil {
		return
	}
	defer conn.Close()

	conn.DB("blue_apps").C("state").Update(
		bson.M{"_id": appPrefix},
		bson.M{"$set": bson.M{
			key: value,
		}})
}

func Hit(prefix string) {
	hitChannel <- prefix
}

func Close() {
	close(hitChannel)
}

func internalHit() {
	var lastHitFailed *time.Time
	for {
		h := <-hitChannel
		if h == "" {
			break
		}

		if lastHitFailed != nil && hitFailing && lastHitFailed.Before(time.Now().Add(time.Second*-10)) {
			hitFailing = false
		}

		if hitFailing {
			continue
		}

		_, err := http.Get("http://localhost:7777/hit?p=" + h)
		if err != nil {
			hitFailing = true
			tm := time.Now()
			lastHitFailed = &tm
			time.Sleep(time.Second * 1)
		}
	}
	fmt.Println("Internal Hit Channel Closed")
}

func SetErrorLogSource(s string) {
	logSource = s
}

func LogError(err error) {
	if err == nil {
		return
	}

	LogMsg(err.Error())
}
func LogMsg(msg string) {
	logLock.Lock()
	defer logLock.Unlock()

	db := newConnection()

	if db == nil {
		return
	}

	defer db.Close()

	item := struct {
		Id     string    `bson:"_id"`
		Msg    string    `bson:"msg"`
		Ts     time.Time `bson:"ts"`
		Source string    `bson:"source"`
	}{
		Id:     blueutil.NewV4(),
		Msg:    msg,
		Ts:     time.Now(),
		Source: logSource,
	}

	db.DB("blue_apps").C("exceptions").Insert(item)
}
