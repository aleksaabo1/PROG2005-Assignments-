package user

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strings"
	"time"
)

/*
Function that give you information about the service if not endpoints is given
*/
func Homepage(w http.ResponseWriter, r *http.Request) {
	_, err := fmt.Fprintf(w, "Welcome to this COVID-19 application.\n"+
		"Enter endpoint for cases per country: /corona/v1/country/{country(?scope=startdate-enddate)}\n"+
		"Enter endpoint for policies per country: /corona/v1/policy/{country(?scope=startdate-enddate)}\n"+
		"Enter endpoint for notifications: /corona/v1/notifications/{id}"+
		"Enter edpoint for diag: /corona/v1/diag")
	if err != nil {
		http.Error(w, "Error: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

/*
Function that will return the confirmed cases and recovered cases of covid-19 in a given country
*/
func PerCountry(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	searchedCountry := strings.Split(r.URL.Path, `/`)[4] //Get the searched country from the URL
	date := r.URL.RawQuery                               //The date

	var startDate string
	var endDate string

	//Gets the information from the API
	body, err := Response("https://covid-api.mmediagroup.fr/v1/cases?country="+strings.Title(searchedCountry), w)
	if err != nil {
		http.Error(w, "Error: "+err.Error(), http.StatusBadRequest)
	}

	//Unmarshalling the information from the api, and append it to the cases array
	var cases CovidAPIResponse
	if err3 := json.Unmarshal(body, &cases); err3 != nil {
		http.Error(w, "Error: "+err3.Error(), http.StatusBadRequest)
		return
	}

	//Checking if the api returned a country. If not the api returned the information for every country
	//and we would not received any information
	if len(cases.All.Country) == 0 {
		http.Error(w, "Country is not available", http.StatusBadRequest)
		return
	}

	var confirmedCases int
	var timeFrame string

	//Checking if date format is valid
	if len(date) != 0 && len(date) == 27 {
		//Checking if user has written correct url
		if date[0:6] != "scope=" {
			http.Error(w, "No valid URL\nEnter /{country}?scope={startDate-endDate}", http.StatusNotFound)
			return
		}

		layout := "2006-01-02"
		endDateCheck, err := time.Parse(layout, date[17:27])
		if err != nil {
			http.Error(w, "Invalid date\nDate must be in format YYYY-MM-DD", http.StatusBadRequest)
			return
		}
		startDateCheck, err := time.Parse(layout, date[6:16])
		if err != nil {
			http.Error(w, "Invalid date\nDate must be in format YYYY-MM-DD", http.StatusBadRequest)
			return
		}

		//Checking if the end date is in the future
		if endDateCheck.After(time.Now()) {
			endDate = Today.String()[0:10]            //setting future date, to today
			startDate = startDateCheck.String()[0:10] //Converting the start date to a string
		} else if startDateCheck.After(endDateCheck) { //Checking if start date is after end date
			startDateCheck, endDateCheck = endDateCheck, startDateCheck // If it is, we are switching the dates
			startDate = startDateCheck.String()[0:10]                   //Converting the start date to a string
			endDate = endDateCheck.String()[0:10]                       //Converting end date to string

		} else {
			startDate = startDateCheck.String()[0:10] //Converting the start date to a string
			endDate = endDateCheck.String()[0:10]     //Converting end date to string
		}

		// Getting confirmed cases of covid-19 in a given country
		body1, err := Response("https://covid-api.mmediagroup.fr/v1/history?country="+strings.Title(searchedCountry)+"&status=Confirmed", w)
		if err != nil {
			http.Error(w, "Error: "+err.Error(), http.StatusBadRequest)
		}

		//Unmarshalling and appending to covid struct
		var covid CovidAPIResponse
		if err3 := json.Unmarshal(body1, &covid); err3 != nil {
			http.Error(w, "Error: "+err3.Error(), http.StatusBadRequest)
			return
		}

		//Loop to iterate if covid cases is 0, and breaks when confirmed cases is not 0
		for {
			if covid.All.Dates[endDate] != 0 {
				break
			}
			endDate = Today.AddDate(0, 0, -1).String()[0:10]
		}

		confirmedCases = covid.All.Dates[endDate] - covid.All.Dates[startDate] // Calculating cases in given time interval
		timeFrame = startDate + "-" + endDate                                  //Formatting time frame to the output we want

	} else if len(date) != 21 && len(date) != 0 {
		http.Error(w, "Error: Invalid input \n", http.StatusBadRequest)
		return
	} else {
		confirmedCases = cases.All.Confirmed
		timeFrame = "total"
	}

	percent := float32(int((float32(confirmedCases)/float32(cases.All.Population)*100)*100)) / 10 //Calculating the percent of cases in the population

	fmt.Fprintf(w, `{
   	"Country": "%v",
   "continent": "%v",
	"scope": "%v",
   "confirmed": "%v",
   "recovered": "%v",
	"population_percentage": "%v"}`,
		cases.All.Country, cases.All.Continent, timeFrame,
		confirmedCases, cases.All.Recovered, percent)

}

/*
Function to return stringency and trend in a given country
*/
func Trends(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	searchedCountry := strings.Split(r.URL.Path, `/`)[4] //Get the searched country from the URL
	//Check if the user as sent with a country
	if len(searchedCountry) == 0 {
		http.Error(w, "Enter a country, to get information\n", http.StatusNoContent)
		return
	}

	date := r.URL.RawQuery //The date

	var startDate string
	var endDate string
	var timeFrame string
	var trend float64
	var stringency float64

	if len(date) != 0 && len(date) == 27 {

		//Checking if user has written correct url
		if date[0:6] != "scope=" {
			http.Error(w, "No valid URL\nEnter /{country}?scope={startDate-endDate}", http.StatusNotFound)
			return
		}

		layout := "2006-01-02"
		endDateCheck, err := time.Parse(layout, date[17:27])
		if err != nil {
			http.Error(w, "Invalid date\nDate must be in format YYYY-MM-DD", http.StatusBadRequest)
			return
		}
		startDateCheck, err := time.Parse(layout, date[6:16])
		if err != nil {
			http.Error(w, "Invalid date\nDate must be in format YYYY-MM-DD", http.StatusBadRequest)
			return
		}

		//Checking if the end date is in the future
		if endDateCheck.After(time.Now()) {
			endDate = Today.String()[0:10]            //setting future date, to today
			startDate = startDateCheck.String()[0:10] //Converting the start date to a string
		} else if startDateCheck.After(endDateCheck) { //Checking if start date is after end date
			startDateCheck, endDateCheck = endDateCheck, startDateCheck // If it is, we are switching the dates
			startDate = startDateCheck.String()[0:10]                   //Converting the start date to a string
			endDate = endDateCheck.String()[0:10]                       //Converting end date to string
		} else if endDateCheck.After(Today.AddDate(0, 0, -10)) { //Check if end date is 10 days before today, to get stringency data
			endDate = Today.AddDate(0, 0, -10).String()[0:10]
		} else {
			startDate = startDateCheck.String()[0:10] //Converting the start date to a string
			endDate = endDateCheck.String()[0:10]     //Converting end date to string
		}

		//Getting the country code, of the searched country
		countryCode := CountryCode(strings.Title(searchedCountry), w)
		if len(countryCode) == 0 {
			http.Error(w, "Country do not exist\n", http.StatusBadRequest)
			return
		}

		//Getting the stringency data from startdate
		body, err := Response("https://covidtrackerapi.bsg.ox.ac.uk/api/v2/stringency/actions/"+countryCode+"/"+startDate, w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		//Adding the data to covidStart struct
		var covidStart Stringency
		if err = json.Unmarshal(body, &covidStart); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		//Getting the stringency data from end date
		body1, err := Response("https://covidtrackerapi.bsg.ox.ac.uk/api/v2/stringency/actions/"+countryCode+"/"+endDate, w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}

		//Adding the data to covidEnd struct
		var covidEnd Stringency
		if err = json.Unmarshal(body1, &covidEnd); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}

		//Checcking if any stringency value is 0
		if covidEnd.Stringencydata.Stringency == 0 || covidStart.Stringencydata.Stringency == 0 {
			stringency = -1
			trend = 0
		} else {
			stringency = math.Round((covidEnd.Stringencydata.Stringency+covidStart.Stringencydata.Stringency)/2*100) / 100 //Calculating stringency
			trend = math.Round((covidEnd.Stringencydata.Stringency-covidStart.Stringencydata.Stringency)*100) / 100        //Calculating trend
		}

		timeFrame = startDate + "-" + endDate //Formatting time frame

	} else if len(date) != 21 && len(date) != 0 {
		http.Error(w, "Error: Invalid input \n", http.StatusBadRequest)
		return
	} else if len(date) == 0 { //If user has not added date

		timeFrame = "total"

		endDate := Today.AddDate(0, 0, -10).String()[0:10] //Setting endDate to 10 days earlier

		//Getting country code of the searched country
		countryCode := CountryCode(strings.Title(searchedCountry), w)
		if len(countryCode) == 0 {
			http.Error(w, "Country do not exist\n", http.StatusBadRequest)
			return
		}

		//Getting stringency of endDate
		body1, err := Response("https://covidtrackerapi.bsg.ox.ac.uk/api/v2/stringency/actions/"+countryCode+"/"+endDate, w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		//Adding data to covid struct
		var covid Stringency
		if err = json.Unmarshal(body1, &covid); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}

		stringency = covid.Stringencydata.Stringency
	}
	fmt.Fprintf(w, `{
    "country": "%v",
    "scope": "%v",
    "stringency": %v,
    "trend": %v
	}`, searchedCountry, timeFrame, stringency, trend)
}

func Diag(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var mmediaGroupStatus int
	var covidTracker int
	var countryAPI int

	//Getting the mmediagroup
	res, err := http.Get("https://mmediagroup.fr/covid-19")
	if err != nil {
		http.Error(w, "Error: Cannot connect to Server", http.StatusInternalServerError)
		mmediaGroupStatus = res.StatusCode //Setting the status code
		return
	} else {
		mmediaGroupStatus = res.StatusCode //Setting the status code
	}

	//Getting covidtracker
	res, err = http.Get("https://covidtracker.bsg.ox.ac.uk")
	if err != nil {
		http.Error(w, "Error: Cannot connect to Server", http.StatusInternalServerError)
		covidTracker = res.StatusCode //Setting the status code
		return
	} else {
		covidTracker = res.StatusCode //Setting the status code
	}

	//Getting the restcountries
	res, err = http.Get("https://restcountries.eu")
	if err != nil {
		http.Error(w, "Error: Cannot connect to Server", http.StatusInternalServerError)
		countryAPI = res.StatusCode //Setting the status code
		return
	} else {
		countryAPI = res.StatusCode //Setting the status code
	}

	runningTime := int(time.Since(Today) / time.Second) //Calculating runningtime

	fmt.Fprintf(w, `{
		"mmediagroupapi": "%v",
		"covidtrackerapi": "%v",
		"restcountries": "%v",
   		"registered": "%v",
   		"version": "v1",
   		"uptime": "%v" 
 		}`, mmediaGroupStatus, covidTracker, countryAPI, countRegistered(), runningTime)

}
