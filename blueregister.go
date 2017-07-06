package blueregister

import (
	"fmt"
	"os"
	"sync"

	"time"

	"net/http"

	"github.com/sfi2k7/blueutil"
	"gopkg.in/mgo.v2/bson"
)

var appPrefix string
var hitFailing bool
var hitChannel chan string
var logErrorChannel chan string
var doneChannel chan bool
var logLock sync.Mutex
var logSource string
var lastHitFailed *time.Time

func init() {
	hitChannel = make(chan string, 1000)
	logErrorChannel = make(chan string, 1000)
	doneChannel = make(chan bool, 1)

	logLock = sync.Mutex{}
	go internalHit2()
}

func Close() {
	doneChannel <- true
	close(hitChannel)
	close(logErrorChannel)
	close(doneChannel)
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

func internalHit2() {
	exitLoop := false

	for !exitLoop {
		select {
		case h := <-hitChannel:
			addHit(h)
			break

		case k := <-logErrorChannel:
			logMsg(k)
			break

		case <-doneChannel:
			fmt.Println("DONE Channel")
			exitLoop = true
			break
		}
	}
	fmt.Println("Exiting Register Loop")
}

func addHit(h string) {
	if h == "" {
		return
	}

	if lastHitFailed != nil && hitFailing && lastHitFailed.Before(time.Now().Add(time.Second*-10)) {
		hitFailing = false
	}

	if hitFailing {
		return
	}

	_, err := http.Get("http://localhost:7777/hit?p=" + h)
	if err != nil {
		hitFailing = true
		tm := time.Now()
		lastHitFailed = &tm
		time.Sleep(time.Second * 1)
	}
}

// func internalHit() {

// 	for {
// 		h := <-hitChannel
// 		if h == "" {
// 			break
// 		}

// 		if lastHitFailed != nil && hitFailing && lastHitFailed.Before(time.Now().Add(time.Second*-10)) {
// 			hitFailing = false
// 		}

// 		if hitFailing {
// 			continue
// 		}

// 		_, err := http.Get("http://localhost:7777/hit?p=" + h)
// 		if err != nil {
// 			hitFailing = true
// 			tm := time.Now()
// 			lastHitFailed = &tm
// 			time.Sleep(time.Second * 1)
// 		}
// 	}
// 	fmt.Println("Internal Hit Channel Closed")
// }

func SetErrorLogSource(s string) {
	logSource = s
}

func LogError(err error) {
	if err == nil {
		return
	}

	logErrorChannel <- err.Error()
}

func LogMsg(msg string) {
	if msg == "" {
		return
	}

	logErrorChannel <- msg
}

func logMsg(msg string) {
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
