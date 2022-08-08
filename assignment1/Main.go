package main

import (
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"os"
	"time"
)

var timeS time.Time

func requests() {
	/// We have two endpoints, for the main root, like localhost:4747, it runs homepage function and for localhost:4747/articles it executes AllArticles function
	router := mux.NewRouter()
	router.HandleFunc("/exchange/v1/exchangehistory/{country_name}/{begin_date-end_date}", currencyHistory)
	router.HandleFunc("/exchange/v1/exchangeborder/{country_name}", BorderRates).Queries("limit", "{limit}")
	router.HandleFunc("/exchange/v1/exchangeborder/{country_name}", BorderRates)
	router.HandleFunc("/exchange/v1/diag/", diagnostics)
	http.Handle("/", router)
	log.Fatal(http.ListenAndServe(getport(), router))

}

func main() {
	timeS = time.Now()
	requests()
}

//// Get Port if it is set by environment, else use a defined one like "4747"
func getport() string {
	var port = os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	return ":" + port
}
