package user

//Struct that will store information about
//Currencies, borders
type Information struct {
	Alpha3Code string
}

type CovidAPIResponse struct {
	All covidCases
}

type covidCases struct {
	Confirmed  int
	Recovered  int
	Country    string
	Continent  string
	Population int
	Dates      map[string]int
}

type Stringency struct {
	Stringencydata struct {
		DateValue        string      `json:"date_value"`
		CountryCode      string      `json:"country_code"`
		Confirmed        int         `json:"confirmed"`
		Deaths           int         `json:"deaths"`
		StringencyActual interface{} `json:"stringency_actual"`
		Stringency       float64     `json:"stringency"`
	} `json:"stringencyData"`
}

type Firebase struct {
	Url     string `json:"url"`
	Timeout int    `json:"timeout"`
	Field   string `json:"field"`
	Country string `json:"country"`
	Trigger string `json:"trigger"`
	Numbers int    `json:"number"`
	Id      string `json:"id"`
}
