# passion-project
Schedule Crafter passion project

Current working example:
[cwooper.me](https://cwooper.me/)

A good example of output/customizability:
[Gizmoa](https://gizmoa.com/college-schedule-maker/)

## How to show/deploy this project onto your website
Proposed file hierarchy:
```
website-root   # Usually /var/www/html/
├── index.html # Website's frontend HTML file
├── schedule-crafter # This project's GitHub repo
│  ├── backend
│  │  ├── app         # Django application code
│  │  ├── manage.py      # Django management script
│  │  ├── requirements.txt # Python package dependencies
│  │  └── webscraper.py  # Webscraper script (if applicable)
│  └── frontend
│    ├── package.json   # React JS project configuration
│    ├── src         # React JS application code
│    │  ├── components    # React components
│    │  ├── templates    # React templates (if applicable)
│    │  └── index.html  # HTML entry point for the React JS application
│  └── static         # Static files (images, CSS, etc.)
└── themes   # Theme files for the website
    ├── default.css     # Default theme for the website
    └── schedule-crafter.css # Theme for the schedule-crafter application (if applicable)
```
User → React Frontend → API Requests → Django Backend → MongoDB Database (if web scraped data is needed)

(if using web scraped data) MongoDB Database → Django Backend → API Responses → React Frontend → User

## Listed from most important to most complex
**The schedule crafter should...**

- allow as little input from the user as possible while getting the most data from it as possible
    - ie. A class should at minimum need a name, date, and time. Ex: "CSCI 301, TR 1-2:40"
- find all of the possible schedules with no overlapping
- determine the "best" schedule based on weights
- provided automatic weights, though customizable
- display the schedules in an easily readable, and customizable format
- save user data by exporting a file (probably .json)

Possible future additions (probably WWU only):
- past WWU data for courses (passing rate, total people taking it, etc.)
- automatic class finder scraping for less input (only input some classes)
- web scrape rate my professor rating, WTA, and difficulty as weights for schedule crafter model
- store the web scraped data in a database for as little CPU usage as possible
- future class scheduler per quarter based on pre-requisites
