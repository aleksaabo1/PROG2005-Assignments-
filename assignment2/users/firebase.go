package user

import (
	"cloud.google.com/go/firestore"
	"context"
	"encoding/json"
	_ "firebase.google.com/go"
	"fmt"
	"google.golang.org/api/iterator"
	_ "google.golang.org/api/option"
	"io/ioutil"
	"log"
	_ "log"
	"net/http"
	"strings"
	//"time"
)

// Firebase context and client used by Firestore functions throughout the program.
var Ctx context.Context
var Client *firestore.Client
var FirebaseArray []Firebase

// Collection name in Firestore
var collection = "registered"

/*
Function to add the registration into firebase
*/
func addRegistration(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		text, err := ioutil.ReadAll(r.Body) // Reading the body of the request
		if err != nil {
			http.Error(w, "Reading of payload failed", http.StatusInternalServerError)
		}

		if len(string(text)) == 0 { //Checks if message is empty
			http.Error(w, "Your message appears to be empty", http.StatusBadRequest)
		} else {
			var notification Firebase
			if err = json.Unmarshal(text, &notification); err != nil { //Unmarshalling the body
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			if len(notification.Country) == 0 || notification.Timeout == 0 || len(notification.Field) == 0 || len(notification.Url) == 0 || len(notification.Country) == 0 || len(notification.Trigger) == 0 {
				http.Error(w, "Please insert value in all fields", http.StatusBadRequest)
				return
			}

			if notification.Trigger == "ON_CHANGE" || notification.Trigger == "ON_TIMEOUT" {
				if notification.Field == "Confirmed" || notification.Field == "Stringency" {

					//Adding the field in the database
					id, _, err := Client.Collection("registered").Add(Ctx,
						map[string]interface{}{
							"url":     notification.Url,
							"timeout": notification.Timeout,
							"field":   notification.Field,
							"country": notification.Country,
							"trigger": notification.Trigger,
							"numbers": GetNumber(notification.Country, notification.Field), //Adding the current number of Stringency/Confirmed
							//To know when to notify the user
						})
					if err != nil {
						http.Error(w, "Error when adding message "+string(text), http.StatusBadRequest)
					} else {
						http.Error(w, "Notifier is registered with ID: "+id.ID, http.StatusCreated) // Returns document ID
					}
					trimmedId := strings.TrimLeft(id.ID, "/") //Trimming the id
					//Adding the Id to a field, for easier access
					_, err = Client.Collection(collection).Doc(trimmedId).Set(Ctx, map[string]interface{}{
						"id": trimmedId,
					}, firestore.MergeAll)

					FirebaseArray = GetFirebase()         //Get the latest entry
					go Invocation(len(FirebaseArray) - 1, 0) //Invokes the added registration

				} else {
					//To ensure the user, adds correct input
					http.Error(w, "Trigger is not valid.\nValid input is Confirmed and Stringency", http.StatusBadRequest)
					return
				}

			} else {
				//To ensure the user, adds correct input
				http.Error(w, "Trigger is not valid.\nValid input is ON_CHANGE and ON_TIMEOUT", http.StatusBadRequest)
				return
			}

		}

	}
}

/*
Function to delete a registration from firebase
*/
func deleteRegistration(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodDelete {
		notificationId := strings.Split(r.URL.Path, "/")[4] // Getting the specific id for the registration
		validRegistration := false
		for i := range FirebaseArray {
			if FirebaseArray[i].Id == notificationId { //Checking if the id is in the firebase
				validRegistration = true
			} else if notificationId == "" {
				http.Error(w, "Please enter an ID for the item you would like to delete", http.StatusBadRequest)
				return
			}
		}

		if validRegistration {
			if len(notificationId) != 0 {
				_, err := Client.Collection(collection).Doc(notificationId).Delete(Ctx) //Command for deleting the firebase entry
				if err != nil {
					http.Error(w, "Deletion of "+notificationId+" failed.", http.StatusInternalServerError)
					return
				} else {
					http.Error(w, "Successful deletion of "+notificationId, http.StatusAccepted)
					return
				}

			} else {
				http.Error(w, "Please enter an ID for the item you would like to delete", http.StatusBadRequest)
			}
			FirebaseArray = GetFirebase() //Updates the FirebaseArray
		} else {
			http.Error(w, "Found no matching ID", http.StatusBadRequest)
			return
		}

	}

}

/*
Function to view a specific registration.
*/
func getRegistration(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		w.Header().Set("Content-Type", "application/json")
		if len(strings.Split(r.URL.Path, "/")) > 4 {
			notificationId := strings.Split(r.URL.Path, "/")[4] //ID of the registration
			if len(notificationId) != 0 {
				doc, err := Client.Collection(collection).Doc(notificationId).Get(Ctx) // Loop through all entries in collection "messages"
				if err != nil {
					http.Error(w, "The notification ID is not in our system", http.StatusBadRequest)
					return
				}

				var firebases []Firebase
				var firebase Firebase
				if err := doc.DataTo(&firebase); err != nil { //Add one firebase entry to the firebase variable
					http.Error(w, "Error: "+err.Error(), http.StatusInternalServerError)
				}
				firebases = append(firebases, firebase) //Add the firebase in the firebase variable, in the an array of firebases

				fmt.Fprintf(w, `{
   		"id": "%v",
		"url": "%v",
		"timeout": "%v",
    	"field": "%v",
    	"country": "%v",
		"trigger": "%v"}`,
					notificationId, firebases[0].Url, firebases[0].Timeout,
					firebases[0].Field, firebases[0].Country, firebases[0].Trigger)
			} else {
				http.Error(w, "Please enter an ID for the item you would like view", http.StatusBadRequest)
			}

		} else {
			firebase := GetFirebase() //Getting all information from firebase
			if len(firebase) == 0 {
				http.Error(w, "No registered notifiers", http.StatusOK)

			} else {
				for i, _ := range firebase { //Iterate through them
					fmt.Fprintf(w, `{
   			"id": "%v",
   			"url": "%v",
			"timeout": "%v",
   			"field": "%v",
   			"country": "%v",
			"trigger": "%v"}`,
						firebase[i].Id, firebase[i].Url, firebase[i].Timeout,
						firebase[i].Field, firebase[i].Country, firebase[i].Trigger)

				}
			}
		}
	}
}

/*
Switch Case to the notifications/firebase
*/
func Notification(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		addRegistration(w, r)
	case http.MethodDelete:
		deleteRegistration(w, r)
	case http.MethodGet:
		getRegistration(w, r)
	}
}

/*
Function to retrieve firebase entries in firebase
*/
func GetFirebase() []Firebase {
	iter := Client.Collection(collection).Documents(Ctx) // Loop through all entries in collection "messages"
	var firebase Firebase
	var fire []Firebase

	//Iterating through the firebase
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}

		//Append the firebase into an instance of firebase
		if err := doc.DataTo(&firebase); err != nil {
			return nil
		}

		//Adding the instance of firebase in an array of firebases
		fire = append(fire, firebase)

	}
	return fire
}

/**
Function to to get the numbers of either Stringency or Confirmed Cases in a country
*/
func GetNumber(country string, field string) int {
	if field == "Confirmed" {
		country = strings.ToLower(country)
		// Getting the confirmed cases in a given country
		urlConfirmed := "https://covid-api.mmediagroup.fr/v1/cases?country=" + strings.Title(country)
		resp, err := http.Get(urlConfirmed) //Getting the url
		if err != nil {
			log.Println(err.Error())
			return 0
		}
		body, err := ioutil.ReadAll(resp.Body) //Reading the body
		if err != nil {
			log.Println(err.Error())
			return 0
		}
		// Appending the data to cases
		var cases CovidAPIResponse
		if err3 := json.Unmarshal(body, &cases); err3 != nil { //Unmarshalling the body
			log.Println(err.Error())
			return 0
		}
		return cases.All.Confirmed //Return confirmed cases

	} else if field == "Stringency" {
		// Getting the stringency in a given country
		urlStringency := "https://covidtrackerapi.bsg.ox.ac.uk/api/v2/stringency/actions/" + CountryCodeString(country) + "/" + Today.AddDate(0, 0, -10).String()[0:10]
		resp, err := http.Get(urlStringency) //Getting the url
		if err != nil {
			log.Println(err.Error())
			return 0
		}

		body, err := ioutil.ReadAll(resp.Body) //Reading the body
		if err != nil {
			log.Println(err.Error())
			return 0
		}
		// Appending the data to cases
		var cases Stringency
		if err3 := json.Unmarshal(body, &cases); err3 != nil { //Unmarshalling the body
			log.Println(err.Error())
			return 0
		}

		return int(cases.Stringencydata.Stringency) //Returning the stringency
	}
	return 0
}
