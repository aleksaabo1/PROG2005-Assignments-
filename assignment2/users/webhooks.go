package user

import (
	"bytes"
	"cloud.google.com/go/firestore"
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"
)

/*
	Calls given URL with given content and awaits response (status and body).
*/
func CallUrl(url string, content string) {

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader([]byte(content)))
	if err != nil {
		fmt.Errorf("%v", "Error during request creation.")
		return
	}

	client := http.Client{}
	res, err := client.Do(req)
	if err != nil {
		log.Println("Error in HTTP request: " + err.Error())
	}
	_, err = ioutil.ReadAll(res.Body)
	if err != nil {
		log.Println("Something is wrong with invocation response: " + err.Error())
	}

}

/*
Function that invokes the submitted url on change
*/

func invOnChange(country string, field string, numbers int, id string) bool {
	FirebaseArray = GetFirebase() //retrieve all the registrations

	numbersFromDataBase := GetNumber(country, field) // Getting the numbers from the database

	if field == "Confirmed" {
		if numbers != numbersFromDataBase { //check if there has been any change
			update(numbersFromDataBase, id) //Updates the database
			return true

		} else {
			return false
		}
	} else if field == "Stringency" {
		if numbers != numbersFromDataBase { //check if there has been any change
			update(numbersFromDataBase, id) //Updates the database
			return true

		} else {
			return false
		}
	}

	return false
}

/*
Function to update the values in the database.
Where number is the preferred field (confirmed og stringency).
*/
func update(newNumber int, id string) {
	ref := Client.Collection(collection).Doc(id)
	err := Client.RunTransaction(Ctx, func(ctx context.Context, tx *firestore.Transaction) error {

		return tx.Set(ref, map[string]interface{}{
			"numbers": newNumber, //Adding the confirmed/Stringency value to the database
		}, firestore.MergeAll)
	})
	if err != nil {
		log.Printf("An error has occurred: %s", err)
	}

}

func Invocation(i int, counter int) {

	FirebaseArray = GetFirebase()                       //Get the latest data from the database
	if FirebaseArray != nil && i < len(FirebaseArray) { //Checks the index and ensures that the database is not empty-
		var yourTime time.Duration
		yourTime = time.Duration(FirebaseArray[i].Timeout * 1000000000) //The timeout variable choosen by the user
		time.AfterFunc(yourTime, func() {                               //Waiting for the function until the time has reached zero
			if i < len(FirebaseArray) {
				if FirebaseArray[i].Trigger == "ON_CHANGE" {
					//Variable to know if we shall notify the user and updates the number in database
					change := invOnChange(FirebaseArray[i].Country, FirebaseArray[i].Field, FirebaseArray[i].Numbers, FirebaseArray[i].Id)
					if change { //If the number of confirmed cases has changes Call the url
						CallUrl(FirebaseArray[i].Url, FirebaseArray[i].Field+" has changed in "+FirebaseArray[i].Country+"\n"+
							"The value is now: "+strconv.Itoa(GetNumber(FirebaseArray[i].Country, FirebaseArray[i].Field)))
					}
				} else if FirebaseArray[i].Trigger == "ON_TIMEOUT" {
					//Variable to know if we shall notify the user and updates the number in database
					onTime := invOnTime(FirebaseArray[i].Country, FirebaseArray[i].Field, FirebaseArray[i].Id)
					if onTime { // Notify the user of the number of confirmed cases
						CallUrl(FirebaseArray[i].Url, FirebaseArray[i].Field+" is now "+strconv.Itoa(GetNumber(FirebaseArray[i].Country, FirebaseArray[i].Field))+
							" in "+FirebaseArray[i].Country+"\n")
					}
				}


				//To make sure the API do not crash
				if counter != 30 {
					Invocation(i, counter + 1) //Calls the function recursive
				}else{
					return
				}
			} else {
				return
			}
		})
	} else {
		return
	}
}

/*
Function to invoke the url when the timeout has reached zero
*/
func invOnTime(country string, field string, id string) bool {
	FirebaseArray = GetFirebase()
	for _ = range FirebaseArray {
		numbersFromDataBase := GetNumber(country, field) //Getting the number to update the firebase
		if field == "Confirmed" {
			update(numbersFromDataBase, id) //Updating the database with the newest numbers from the api
			return true

		} else if field == "Stringency" {
			update(numbersFromDataBase, id) //Updating the database with the newest numbers from the api
			return true
		}
	}
	return false
}
