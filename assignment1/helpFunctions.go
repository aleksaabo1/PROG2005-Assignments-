package main

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"io/ioutil"
	"net/http"
	"strings"
)

/**
Function that return the currency code of a given country
*/
func currencyCode(w http.ResponseWriter, r *http.Request) []string {
	var currencies []string
	var info []Information

	vars := mux.Vars(r)
	country := vars["country_name"]
	resp, err := http.Get("https://restcountries.eu/rest/v2/name/" + country + "?fields=currencies")
	if err != nil {
		http.Error(w, "Error: "+err.Error(), http.StatusBadRequest)
		return nil
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "Error: Cannot connect", http.StatusInternalServerError)
		return nil
	}

	//Gets the information we need
	if err := json.Unmarshal([]byte(string(body)), &info); err != nil {
		http.Error(w, "Error: Unable to find country", http.StatusInternalServerError)
		return nil
	}

	//Add the currency in an array of currencies
	if info != nil {
		currencies = append(currencies, info[0].Currencies[0].Code)
		return currencies
	}
	return nil
}

/**
Function that returns an array of string, with the currency code of an array of countries
*/
func currencyCodeString(w http.ResponseWriter, rS []string) []string {
	var currencyCode Information

	var currencies []string //Stores the currency

	//Goes through the array of countries
	for i := 0; i < len(rS); i++ {

		resp, err := http.Get("https://restcountries.eu/rest/v2/alpha/" + rS[i] + "?fields=currencies")
		if err != nil {
			http.Error(w, "Error: "+err.Error(), http.StatusBadRequest)
			return nil
		}
		//Reads the input of the requested URL
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			http.Error(w, "Error: "+err.Error(), http.StatusInternalServerError)
			return nil
		}

		//Retrieves the information of the request, and place it in an array
		if err := json.Unmarshal([]byte(string(body)), &currencyCode); err != nil {
			http.Error(w, "Error: "+err.Error(), http.StatusInternalServerError)
			return nil
		}

		//Adds the currency code to an array
		if len(currencyCode.Currencies) != 0 {
			currencies = append(currencies, currencyCode.Currencies[0].Code)
		}

	}

	return currencies

}

/**
Function that return a string of array, with the bordering countries of a given country
*/
func countryBorder(w http.ResponseWriter, r *http.Request) []string {
	var info []Information

	elem := strings.Split(r.URL.Path, "/")
	country := elem[4]

	resp, err := http.Get("https://restcountries.eu/rest/v2/name/" + country + "?fields=borders")
	if err != nil {
		http.Error(w, "Error: "+err.Error(), http.StatusBadRequest)
		return nil
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "Error: "+err.Error(), http.StatusInternalServerError)
		return nil
	}

	//Place the information requested in an array
	if err := json.Unmarshal([]byte(string(body)), &info); err != nil {
		http.Error(w, "Error: "+err.Error(), http.StatusInternalServerError)
		return nil
	}

	//Adding the bordering countries to an array
	var borders []string
	limit := len(info[0].Borders)
	for i := 0; i < limit; i++ {
		borders = append(borders, info[0].Borders[i])
	}

	return borders
}

/**
Function that delete the currency, that is equal to the base currency
*/
func filterDuplicates(base string, countries []string) []string {
	for i := 0; i < len(countries); i++ {
		if base == countries[i] {
			countries[i] = countries[len(countries)-1] // Copy last element to index i.
			countries[len(countries)-1] = ""           // Erase last element (write zero value).
			countries = countries[:len(countries)-1]   // Truncate slice.
			i--
		}

	}
	return countries
}

/**
Function that turns a request into a body
*/
func response(request string, w http.ResponseWriter) ([]byte, error) {
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
Function to check if a request is equal to EUR
*/
func currencyCheck(w http.ResponseWriter, r *http.Request) bool {
	inputCountry := currencyCode(w, r)

	if inputCountry != nil {
		return true
	} else if inputCountry[0] == "EUR" {
		http.Error(w, "Error: Cannot give history of a country with currency Euro ", http.StatusInternalServerError)
		return false
	}
	return true
}

/**
Function verify that the search country is a real country
*/
func checkIfExist(country string, a []string, w http.ResponseWriter) bool {
	body, err := response("https://restcountries.eu/rest/v2/name/"+country+"?fields=currencies", w)
	if err != nil {
		http.Error(w, "Error: "+err.Error(), http.StatusBadRequest)
		return true
	}

	var countryCode []Information
	if err = json.Unmarshal([]byte(string(body)), &countryCode); err != nil {
		http.Error(w, "Error: Did not find the searched country", http.StatusBadRequest)
		return true
	}

	countryC := countryCode[0].Currencies[0].Code
	for i := 0; i < len(a); i++ {
		if countryC == a[i] {
			return true
		}
	}
	return false
}
