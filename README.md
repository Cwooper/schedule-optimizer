# passion-project
Schedule Crafter passion project

A good example of output/customizability:
[Gizmoa](https://gizmoa.com/college-schedule-maker/)

## Project goals
File hierarchy:
```
project root
├── LICENSE
├── README.md
├── backend
│   ├── Course.py                   # Course object
│   ├── Schedule.py                 # Schedule object
│   ├── data
│   │   ├── 202420.pkl              # Pickle with all course data compiled
│   │   ├── ... all subjects.txt    # Each subjects courses
│   │   └── subjects.txt            # List of subjects
│   ├── data_refresh.py             # Remakes the pickle (if-needed)
│   ├── get_schedules.py            # Local version of what the server should do
│   └── schedule_generator.py       # Generates all possible schedules
└── frontend
    ├── package.json
    └── src
        └── PrettyWebsiteStuff
            ├── displayClasses.html
            ├── index.html
            ├── script.js
            ├── styles.css
            └── subjects.txt
```
User → Frontend → API Requests → Django Backend

Django Backend → API Responses → Frontend → User

Django should handle all requests for course schedules.

## Listed from most important to most complex
**The schedule crafter should...**

- [x] allow users to input only subject and code, e.g. "CSCI 301"
- [x] find all of the possible schedules with no overlapping
- [ ] determine the "best" schedule based on weights
- [ ] provided automatic weights, though customizable
- [ ] display the schedules in an easily readable, and customizable format
- [ ] save user data by exporting a file (probably .json)

Possible future additions (probably WWU only):
- [ ] past WWU data for courses (passing rate, total people taking it, etc.)
- [x] automatic class finder scraping for less input (only input some classes)
- [ ] web scrape rate my professor rating, WTA, and difficulty as weights for schedule crafter model
- [x] store the web scraped data in a database for as little CPU usage as possible
- [ ] future class scheduler per quarter based on pre-requisites
