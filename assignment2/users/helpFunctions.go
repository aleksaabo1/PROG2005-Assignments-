package user

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

var Today time.Time

func Response(request string, w http.ResponseWriter) ([]byte, error) {
	resp, err := http.Get(request)

	if err != nil {
		http.Error(w, "Error: "+err.Error(), http.StatusBadRequest)
		return nil, nil
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "Error: "+err.Error(), http.StatusInternalServerError)
		return nil, nil
	}
	return body, nil
}

/**
Function that return a string of array, with the bordering countries of a given country
*/
func CountryCode(country string, w http.ResponseWriter) string {

	var countryC string
	resp, err := http.Get("https://restcountries.eu/rest/v2/name/" + country)
	if err != nil {
		log.Println(err.Error())
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err.Error())
	}

	var countryCode []Information
	if err = json.Unmarshal(body, &countryCode); err != nil {
		log.Println(err.Error())
	}

	if len(countryCode) == 0 {
		return ""
	} else {
		countryC = countryCode[0].Alpha3Code
	}

	return countryC

}

func countRegistered() int {
	firebase := GetFirebase()
	return len(firebase)
}

func CountryCodeString(country string) string {

	var countryC string
	resp, err := http.Get("https://restcountries.eu/rest/v2/name/" + country)
	if err != nil {
		log.Println(err.Error())
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err.Error())
	}

	var countryCode []Information
	if err = json.Unmarshal(body, &countryCode); err != nil {
		log.Println(err.Error())
	}

	if len(countryCode[0].Alpha3Code) == 0 {

	} else {
		countryC = countryCode[0].Alpha3Code
	}

	return countryC

}
