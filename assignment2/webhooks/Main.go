package main

import (
	user "assignment2/users"
	"context"
	firebase "firebase.google.com/go"
	"google.golang.org/api/option"
	"log"
	"net/http"
	"os"
	"time"
)

var sa option.ClientOption //Global variables

//// Get Port if it is set by environment, else use a defined one like "8080"
func getPort() string {
	var port = os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	return ":" + port
}

func main() {

	user.Today = time.Now()
	log.Println("Listening on port: " + getPort())

	firebaseInit() //Initialising firebase

	//invokes the added registrations
	for i := range user.GetFirebase() {
		go user.Invocation(i, 0)
		time.Sleep(1 * time.Second)
	}

	handlers() //HTTP handlers

	defer user.Client.Close()

}

//Handlers for the service
func handlers() {
	http.HandleFunc("/", user.Homepage)                             //Homepage
	http.HandleFunc("/corona/v1/country/", user.PerCountry)         //Cases per Country
	http.HandleFunc("/corona/v1/policy/", user.Trends)              //Get trend
	http.HandleFunc("/corona/v1/diag", user.Diag)                   //Get diagnostics of the service
	http.HandleFunc("/corona/v1/diag/", user.Diag)                  //Get diagnostics of the service
	http.HandleFunc("/corona/v1/notifications/", user.Notification) //notification endpoint
	http.HandleFunc("/corona/v1/notifications", user.Notification)  //notification endpoint
	log.Println(http.ListenAndServe(getPort(), nil))
}

//Firebase initialisation
func firebaseInit() {
	// Firebase initialisation
	user.Ctx = context.Background()

	// We use a service account, load credentials file that you downloaded from your project's settings menu.
	// Make sure this file is git ignored, it is the access token to the database.
	sa = option.WithCredentialsFile("webhooks/proje-e91eb-firebase-adminsdk-arub8-dbdb3867f5.json")
	app, err := firebase.NewApp(user.Ctx, nil, sa)
	if err != nil {
		log.Fatalln(err)
	}
	user.Client, err = app.Firestore(user.Ctx)
	if err != nil {
		log.Fatalln(err)
	}

}
