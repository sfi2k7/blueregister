package blueregister

import (
	"os"

	"time"

	"gopkg.in/mgo.v2/bson"
	"net/http"
)

var appPrefix string
var hitFailing bool
var hitChannel chan string

func init()  {
	hitChannel = make(chan string,1000)
	go internalHit() 
}

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

func Hit(prefix string){
	hitChannel <- prefix 
}

func Close(){
	close(hitChannel)
}

func internalHit(){
	for{
		h:= <- hitChannel
		if h == ""{
			break
		}
		
		if hitFailing {
			continue
		}
		
		_,err:= http.Get("http://localhost:7777/hit?p="+h)
		if err != nil{
			hitFailing = true
			time.Sleep(time.Millisecond *10) 
			go func(){
				time.Sleep(time.Second *10) 
				hitFailing = false
			}()
		}
	}
}
