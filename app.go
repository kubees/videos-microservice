package main

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/julienschmidt/httprouter"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
)

var environment = os.Getenv("ENVIRONMENT")
var redisHost = os.Getenv("REDIS_HOST")
var redisPort = os.Getenv("REDIS_PORT")
var flaky = os.Getenv("FLAKY")

var ctx = context.Background()
var rdb *redis.Client

func main() {
	r := redis.NewClient(&redis.Options{
		Addr: redisHost + ":" + redisPort,
		DB:   0,
	})
	rdb = r

	router := httprouter.New()

	router.GET("/", HandleHealthz)
	router.GET("/:id", HandleGetVideoById)

	fmt.Println("Running...")
	log.Fatal(http.ListenAndServe(":10010", router))
}

func video(writer http.ResponseWriter, request *http.Request, p httprouter.Params) (response string) {

	id := p.ByName("id")
	fmt.Print(id)

	videoData, err := rdb.Get(ctx, id).Result()
	if err == redis.Nil {
		return "{}"
	} else if err != nil {
		panic(err)
	} else {
		return videoData
	}
}
