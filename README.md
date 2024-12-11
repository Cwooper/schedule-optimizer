# schedule-optimizer

Schedule Optimizer passion project for ACM club at WWU

This project is not affiliated with Western Washington University.
It is an independent initiative developed solely for educational
and personal use. All data provided by this project is for
informational purposes only and should not be considered official
or binding. Use at your own discretion.

Currently hosted live at [cwooper.me](https://cwooper.me/schedule-optimizer)

To use the live website, try entering a few classes:

- MATH 099
- ENG 321
- CSCI 141
- CSCI 145

Try changing the Minimum Courses per schedule to 3.

Press Submit and view your options!

## Requirements

### Frontend

- npm

### Backend

- Go 1.22 or later
- Protocol Buffer Compiler (protoc) 3.0.0 or later
- Go Protocol Buffers plugin

## Getting Started

1. Install `Go`
2. Install Go packages:

   ```bash
   cd backend
   go mod download # Install necessary packages
   cd ../
   ```

3. Install `npm`
4. Install React Modules and build the frontend:

    ```bash
    cd frontend
    make build
    ```

5. Test the server:

    ```bash
    cd backend
    go run server.go
    ```

    Navigate to [localhost:8080/schedule-optimizer](localhost:8080/schedule-optimizer)

## Project To-do list

Fully Transfer backend:

- [x] Create Course Model
  - [x] Course Conflicts
  - [x] Create Course Array ProtoBuf
- [x] Create Schedule Model
  - [x] Schedule Auto-weighing
  - [x] Modular Weight System
- [x] Create Web Scraper
- [x] Create Schedule Generator
- [x] Create Go Web Server
- [x] Initialize GPA Values
  - [x] Process CSV to efficient Protobuf
  - [x] Process Course GPA Values
- [x] Interface Web Server with old frontend
  - [x] Unify data
  - [x] Handle requests to and from
  - [x] Display Old Calendar (doesn't need to be 100%)
  - [x] Display Full Calendar and Update HTML

- [x] Backend Is Fully Transferred and Optimized

Extra:

- [x] Multithread WebScraper, GPA Processing
- [x] Add Async/TBD table to frontend calendar
- [x] Automatic Web Scraping/Course Updating
- [ ] Create Dijkstra Map Weighing
  - [ ] Visualize something in Go WASM
  - [ ] Create Paths between classes per schedule
- [x] Port frontend to React (Konnor too)
- [x] Fuzzy search courses
- [x] Weight customizability on frontend
- [x] If User Asks for the same schedule twice, don't send POST request.
- [ ] Docker containerize application

Frontend TODOs:

- [x] Course Popups
- [x] Better automatic sizing
- [x] Async Courses
  - [x] Fix Async Courses (bandaid)
  - [ ] Full fix async courses
- [x] Fuzzy Search
  - [ ] Better backend fuzzy search
- [x] Duplicate course checking before adding
- [ ] Dark Mode button
- [x] Sort and Weighting menu popup
  - [x] Sort schedules
  - [x] Customize weights
- [x] png save for `schedule-glance`
