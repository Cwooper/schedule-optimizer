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

- [x] create `Course` object
- [x] create internal time conflict handling within `Course` objects
- [x] create `Schedule` structure to hold lists if `Course`s.
- [x] allow users to input only subject and code, e.g. "CSCI 301"
    - [x] functional as `backend/get_schedules.py` with no weights
- [x] find all of the possible schedules with no overlapping
- [ ] automatic class finder scraping for less input (only input some classes)
    - [x] class finder webscraping
    - [x] store webscraped data as a `.pkl` file per term
    - [ ] automatic webscraping
    - [ ] automatically finding available terms from now, into the future
- [ ] past WWU data for courses (passing rate, total people taking it, etc.)
    - [x] obtain gpa rate
    - [ ] apply `gpa` attribute to `Course` object
- [ ] determine the "best" `Schedule` based on weights
    - [ ] add gpa weight
    - [ ] add `start_time` weight
    - [ ] add `end_time` weight
    - [ ] add `time_gap` weight
    - [ ] add `uniqueness` weight (so that all schedules all don't look the same)
- [ ] provided automatic weights on the frontend
    - [ ] add customizability to weights on frontend
- [ ] display the schedules in an easily readable format (maybe a library)
    - [ ] possibly customize schedule viewing to show which courses can interchange with other courses, per schedule
- [ ] save user data by exporting a file (probably `.json`)

Possible future additions (probably WWU only):
- [x] store the web scraped data in a database for as little CPU usage as possible
- [ ] future class scheduler per quarter based on pre-requisites
