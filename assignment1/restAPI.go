package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	_ "github.com/gorilla/mux"
	"net/http"
	"strconv"
	"strings"
	"time"
)

//Struct that will store information about
//Currencies, borders
type Information struct {
	Currencies []Currencies
	Borders    []string
}

//Struct that will store information about currencies
type Currencies struct {
	Code   string `json:"code"`
	Name   string `json:"name"`
	Symbol string `json:"symbol"`
}

//Valid currency codes that is supported in the api
var supportedCurrencies = []string{"EUR", "USD", "MXN", "SGD", "AUD", "MYR", "BGN", "HUF", "CZK", "GBP", "RON", "SEK",
	"IDR", "INR", "BRL", "RUB", "HRK", "JPY", "THB", "CHF", "CAD", "HKD", "ISK", "PHP", "DKK", "TRY", "CNY", "NOK", "NZD", "ZAR", "ILS", "KRW", "PLN"}

/**
Function that vil return the currency history of the searched country
*/
func currencyHistory(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	request := mux.Vars(r)

	todaysDate := time.Now().Format("2006-01-02") //Variable that stores today's date

	currencyCode := strings.Join(currencyCode(w, r), ``) //Translate an array of string to string
	if currencyCode == "" {
		http.Error(w, "Error: Unable to find country", http.StatusBadRequest)
		return
	}
	dates := request["begin_date-end_date"]
	startDate := dates[0:10] //Trims the starting date to format yyyy-mm-dd
	endDate := dates[11:21]  //Trims the end date to format yyyy-mm-dd

	fmt.Println(dates)
	fmt.Println(startDate)
	fmt.Println(endDate)

	//Check if the user has inserted a valid date
	if endDate > todaysDate {
		http.Error(w, " Error: End date is invalid", http.StatusBadRequest)
		return
	}

	//Check if the currency is not EUR, which is the base currency
	if currencyCheck(w, r) {
		body, err := response("https://api.exchangeratesapi.io/history?start_at="+startDate+"&end_at="+endDate+"&symbols="+currencyCode, w)
		fmt.Println("https://api.exchangeratesapi.io/history?start_at=" + startDate + "&end_at=" + endDate + "&symbols=" + currencyCode)
		if err != nil {
			http.Error(w, "Error: "+err.Error(), http.StatusBadRequest)
			return
		}

		fmt.Fprintf(w, "%s", string(body))
	}

}

/*
Function to check the currency of the search country`s neighbour, compared to the
searched country currency
*/
func BorderRates(w http.ResponseWriter, r *http.Request) {
	request := mux.Vars(r)
	w.Header().Set("Content-Type", "application/json")
	countryName := request["country_name"]

	//If sentence to check if the user searched a valid country
	if checkIfExist(countryName, supportedCurrencies, w) {
		body, err := response("https://restcountries.eu/rest/v2/name/"+countryName+"?fields=borders;currencies", w)
		if err != nil {
			http.Error(w, "Error: "+err.Error(), http.StatusBadRequest)
			return
		}

		//Stores the information in the array, neighborCountry
		var neighborCountry []Information
		if err = json.Unmarshal([]byte(string(body)), &neighborCountry); err != nil {
			http.Error(w, "Error: Unable to find neighbor country", http.StatusInternalServerError)
			return
		}

		borders := countryBorder(w, r) //Borders of the searched country

		baseCurrency := currencyCode(w, r)[0] //The currency of the searched country

		var neighborCurrency []string                     //Array that stores the neighbor currency
		neighborCurrency = currencyCodeString(w, borders) //stores the neighbor counties currency

		//Filter out currency that is not supported in the api
		var neighborCurrencyNoDuplicates []string
		neighborCurrencyNoDuplicates = filterDuplicates(baseCurrency, neighborCurrency)

		//Checks if the user has given a limit
		limit := len(neighborCountry[0].Borders)
		val, ok := request["limit"]
		limitNew, err := strconv.Atoi(val)
		if ok && limitNew < limit && limitNew != 0 {
			limit = limitNew
		}

		//Check if the given limit is valid
		var loopIndex int
		if limit < len(neighborCurrencyNoDuplicates) {
			loopIndex = limit
		} else {
			loopIndex = len(neighborCurrencyNoDuplicates)
		}

		//Converts the currency array into a string, that we can insert in the URL
		var currencyString string
		for i := 0; i < loopIndex; i++ {
			for _, b := range supportedCurrencies {
				if b == neighborCurrencyNoDuplicates[i] {
					currencyString += neighborCurrencyNoDuplicates[i] + "," //Add the element in the array to a string
				}
			}
		}

		//Cuts the last ,
		currencyString = strings.TrimRight(currencyString, ",")

		//Request the currency of the neighbor countries, with the searched county as base
		body, err = response("https://api.exchangeratesapi.io/latest?symbols="+currencyString+";base="+baseCurrency, w)
		if err != nil {
			http.Error(w, "Error: "+err.Error(), http.StatusBadRequest)
			return
		}
		fmt.Fprint(w, string(body))

	} else {
		http.Error(w, "Error: Country not available ", http.StatusInternalServerError)
		return
	}
}

/**
Function to observe the status of the used api`s
*/
func diagnostics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var statusExchange int
	var statusCountry int

	//Getting the status code for exchangeratesapi
	respExchange, err := http.Get("https://exchangeratesapi.io")
	if err != nil {
		statusExchange = http.StatusInternalServerError //Stores the status code to exchangeratesapi
		return

	} else {
		statusExchange = respExchange.StatusCode
	}

	//Getting the status code for restcountries
	respCountries, err := http.Get("https://restcountries.eu/")
	if err != nil {
		statusCountry = http.StatusInternalServerError //Stores the status code to restcountries
		return
	} else {
		statusCountry = respCountries.StatusCode
	}

	//The running time of the product
	runningTime := int(time.Since(timeS) / time.Second)

	//Prints the information
	fmt.Fprintf(w, `{
   	"exchangeratesapi": "%v",
   "restcountries": "%v",
   "version": "v1",
   "uptime": %v }`, statusExchange, statusCountry, runningTime)
}
