package main

import (
	"context"
	"go.uber.org/zap"
	"net/http"
	"os"

	"github.com/go-redis/redis/extra/redisotel/v9"
	"github.com/go-redis/redis/v9"
	"github.com/julienschmidt/httprouter"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	metrics "github.com/slok/go-http-metrics/metrics/prometheus"
	"github.com/slok/go-http-metrics/middleware"
	httproutermiddleware "github.com/slok/go-http-metrics/middleware/httprouter"
)

const (
	metricsAddr = ":8000"
)

var environment = os.Getenv("ENVIRONMENT")
var redisHost = os.Getenv("REDIS_HOST")
var redisPort = os.Getenv("REDIS_PORT")
var password = os.Getenv("PASSWORD")
var flaky = os.Getenv("FLAKY")

var ctx = context.Background()
var rdb redis.UniversalClient
var Logger, _ = zap.NewProduction()
var Sugar = Logger.Sugar()

func main() {
	r := redis.NewUniversalClient(&redis.UniversalOptions{
		Addrs:    []string{redisHost + ":" + redisPort},
		DB:       0,
		Password: password,
	})
	rdb = r

	// Enable tracing instrumentation.
	if err := redisotel.InstrumentTracing(r); err != nil {
		Sugar.Errorw("Error while instrumenting redis traces")
		panic(err)
	}

	// Enable metrics instrumentation.
	if err := redisotel.InstrumentMetrics(r); err != nil {
		Sugar.Errorw("Error while instrumenting redis metrics")
		panic(err)
	}

	// Create our middleware.
	promMiddleware := middleware.New(middleware.Config{
		Recorder: metrics.NewRecorder(metrics.Config{}),
	})

	router := httprouter.New()

	router.GET("/", httproutermiddleware.Handler("/", HandleHealthz, promMiddleware))
	router.GET("/:id", httproutermiddleware.Handler("/:id", HandleGetVideoById, promMiddleware))

	Sugar.Infow("Running...")
	// Serve our metrics.
	go func() {
		Sugar.Infof("metrics listening at %s", metricsAddr)
		if err := http.ListenAndServe(metricsAddr, promhttp.Handler()); err != nil {
			log.Panicf("error while serving metrics: %s", err)
		}
	}()

	log.Fatal(http.ListenAndServe(":10010", router))

}

func video(writer http.ResponseWriter, request *http.Request, p httprouter.Params) (response string) {

	id := p.ByName("id")
	Sugar.Infof("Getting video with the ID: %v\n", id)

	videoData, err := rdb.Get(ctx, id).Result()
	if err == redis.Nil {
		Sugar.Infof("Item with the ID: %v is not found\n", id)
		return "{}"
	} else if err != nil {
		Sugar.Errorw("Error while fetching video from DB", err)
		panic(err)
	} else {
		Sugar.Infof("Successfully fetched the video with ID %v from redis\n", id)
		return videoData
	}
}
