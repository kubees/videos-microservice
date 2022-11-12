package main

import (
	"fmt"
	"github.com/julienschmidt/httprouter"
	"math/rand"
	"net/http"
)

func HandleHealthz(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	fmt.Fprintf(w, "ok!")
}

func HandleGetVideoById(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	if flaky == "true" {
		if rand.Intn(90) < 30 {
			panic("flaky error occurred ")
		}
	}

	video := video(w, r, p)

	Cors(w)
	fmt.Fprintf(w, "%s", video)
}
