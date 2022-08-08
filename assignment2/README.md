# Assignment2
In this assignment the task was to create a REST web application.
This application is providing information of corona cases occurring in different regions,
as well as government responses.

Endpoint: /corona/v1/country/
    Will return the number og confirmed and recovered cases of Covid-19 in a given country. 
    The user can specify a date range. 
    
Endpoint: /corona/v1/policy/
    Will return an overview of the stringencylevel of policies regarding Covid-19
    for a given country. 
    
Endpoint: /corona/v1/diag/
    Indicates the availability of all individual services the service depends on. 
    
Endpoint:   /corona/v1/notifications/
    Can register a webhook that is being triggers when a specific event occurs, related
    to stringency or Confirmed Cases. 
    You can also view all registered webhook registrations. In addition to delete       registrations. 
    When you use the post method, insert this body:
    {
    "url": "https://localhost:8080/client/",
     "timeout": 3600,
     "field": "stringency",
     "country": "France",
     "trigger": "ON_CHANGE"
      }

    
There is no client. For testing, feel free to use webhook.site, and use the given url. 


NB: Some of the firebase code, and Webhooks functions are from Christoffer's example code 



