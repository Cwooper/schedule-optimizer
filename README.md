# schedule-optimizer

Schedule Crafter passion project for ACM

This project is not affiliated with Western Washington University.
It is an independent initiative developed solely for educational
and personal use. All data provided by this project is for
informational purposes only and should not be considered official
or binding. Use at your own discretion.

## Requirements

Python libraries:

- flask
- pandas
- numpy
- bs4

## Project goals

User → Frontend → Flask → Backend

Backend → Flask → Frontend → User

Flask should handle all requests for course schedules.

### Listed from most important to most complex

**The schedule crafter should...**

- [x] create `Course` object
- [x] create internal time conflict handling within `Course` objects
- [x] create `Schedule` structure to hold lists if `Course`s.
- [x] allow users to input only subject and code, e.g. "CSCI 301"
    - [x] functional as `backend/get_schedules.py` with no weights
- [x] find all of the possible schedules with no overlapping
- [x] automatic class finder scraping for less input (only input some classes)
    - [x] class finder webscraping
    - [x] store webscraped data as a `.pkl` file per term
    - [x] automatic webscraping
    - [x] automatically finding available terms from now, into the future
- [x] past WWU data for courses (passing rate, total people taking it, etc.)
    - [x] obtain gpa rate
    - [x] apply `gpa` attribute to `Course` object
- [x] determine the "best" `Schedule` based on weights
    - [x] add gpa weight
    - [x] add `start_time` weight
    - [x] add `end_time` weight
    - [x] add `time_gap` weight
    - [ ] add `class_distance` weight (Dijkstra's)
- [x] have a working frontend with well defined inputs
- [x] send and recieve signlas from a web server
- [x] have a "force" button for necessary courses
- [ ] provided automatic weights on the frontend
    - [ ] add customizability to weights on frontend
- [x] display the schedules in an easily readable format (maybe a library)
- [ ] save user data by exporting a file (probably `.json`)

Possible future additions (probably WWU only):
- [x] store the web scraped data in a database for as little CPU usage as possible
- [ ] future class scheduler per quarter based on pre-requisites
