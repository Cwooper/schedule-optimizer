# Schedule Optimizer

This is the backend for the Schedule Optimizer Web Application.

## Prerequisites

- Go 1.22 or later
- Protocol Buffer Compiler (protoc) 3.0.0 or later
- Go Protocol Buffers plugin

## Setup

1. Install the Protocol Buffer Compiler (protoc) if you haven't already:
   - MacOS: `brew install protobuf`
   - Linux: `sudo apt install protobuf-compiler`
   - Windows: Download from <https://github.com/protocolbuffers/protobuf/releases>

2. Install the `Go` Protocol Buffers plugin:

   ```bash
   go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
   ```

3. Make sure your PATH includes the `Go` bin directory:

   ```bash
   export PATH="$PATH:$(go env GOPATH)/bin"
   ```

4. To run the server:

   ```bash
   PORT=xxxx go run server.go
   # For Port 48920:
   go run server.go
   ```


## Running Schedule Optimizer with System Startup and Apache2

Follow these steps to set up the Schedule Optimizer to run automatically with system startup.

## 1. Create a service file for Schedule Optimizer

Create a new file `/etc/systemd/system/schedule-optimizer.service` with the following content:

```ini
[Unit]
Description=Schedule Optimizer Go Server
After=network.target

[Service]
ExecStart=/usr/local/go/bin/go run /var/www/schedule-optimizer/backend/server.go
WorkingDirectory=/var/www/schedule-optimizer/backend
User=www-data
Restart=always
RestartSec=10
Environment="GOPATH=/var/lib/go-cache"
Environment="GOCACHE=/var/lib/go-cache"

[Install]
WantedBy=multi-user.target
```

## 2. Set up Go cache directory

Create a directory for the Go module cache and set appropriate permissions:

```bash
sudo mkdir -p /var/lib/go-cache
sudo chown www-data:www-data /var/lib/go-cache
sudo chown -R www-data:www-data /var/www/schedule-optimizer/data/
```

## 3. Configure Apache2

1. Enable necessary Apache modules:

```bash
sudo a2enmod proxy
sudo a2enmod proxy_http
```

2. Edit your Apache site configuration (often in `/etc/apache2/sites-available/000-default.conf`) and add the following under the `DocumentRoot` line:

```apache
ProxyPass /schedule-optimizer http://localhost:48920/schedule-optimizer
ProxyPassReverse /schedule-optimizer http://localhost:48920/schedule-optimizer
```

## 4. Enable and start the services

```bash
sudo systemctl daemon-reload
sudo systemctl enable schedule-optimizer
sudo systemctl start schedule-optimizer
sudo systemctl restart apache2
```

## 5. Verify the setup

Check the status of both services:

```bash
sudo systemctl status schedule-optimizer
sudo systemctl status apache2
```

Both services should be active and running.

## Troubleshooting

If you encounter issues:

1. Check the logs:
   ```bash
   sudo journalctl -u schedule-optimizer -n 100 --no-pager
   ```

2. Ensure all file permissions are correct:
   ```bash
   sudo chown -R www-data:www-data /var/www/schedule-optimizer
   ```

3. If needed, manually download Go dependencies:
   ```bash
   sudo -u www-data GOPATH=/var/lib/go-cache GOCACHE=/var/lib/go-cache /usr/local/go/bin/go mod download
   ```
   Run this in the `/var/www/schedule-optimizer` directory.

Remember to restart both services after making any changes:

```bash
sudo systemctl restart schedule-optimizer
sudo systemctl restart apache2
```

## Compiling Protocol Buffers

To compile the Protocol Buffer files and generate the corresponding `Go` code, run:

```bash
make proto
```

This will compile all `.proto` files in the `internal/models/proto` directory and
output the generated `Go` files to `internal/proto/generated`.

To remove all generated Protocol Buffer Go files, run:

```bash
make clean
```

For a list of available make targets, run:

```bash
make help
```

## Project Backend Structure

```bash
backend
├── data
│   ├── 202410.pb
│   ├── 202420.pb
│   ├── 202430.pb
│   ├── 202440.pb
│   ├── 202510.pb
│   ├── 202520.pb
│   ├── grade_distribution.csv
│   ├── grade_distribution.pb
│   ├── subjects.txt
│   └── terms.txt
├── frontend
│   ├── favicon.ico
│   ├── index.html
│   ├── scripts.js
│   ├── styles.css
│   └── subjects.txt
├── go.mod
├── go.sum
├── internal
│   ├── generator
│   │   ├── combinations.go
│   │   └── generator.go
│   ├── gpa
│   │   ├── gpa.go
│   │   └── loader.go
│   ├── models
│   │   ├── course.go
│   │   ├── gpa_data.go
│   │   ├── proto
│   │   │   ├── course.proto
│   │   │   └── gpa_data.proto
│   │   ├── raw_request.go
│   │   ├── response.go
│   │   └── schedule.go
│   ├── proto
│   │   └── generated
│   │       ├── course.pb.go
│   │       └── gpa_data.pb.go
│   ├── README.md
│   ├── scraper
│   │   ├── extract.go
│   │   ├── helpers.go
│   │   └── scraper.go
│   └── utils
│       └── utils.go
├── LICENSE
├── Makefile
├── pkg
│   └── protoutils
│       └── convertor.go
├── README.md
└── server.go

```

## TODO

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
- [x] Optimize Forced Courses
- [x] If User Asks for the same schedule twice, don't send POST request.
- [x] Add Async/TBD table to frontend calendar
- [ ] Create Dijkstra Map Weighing
  - [ ] Visualize something in Go WASM
  - [ ] Create Paths between classes per schedule
- [ ] Fuzzy search a "Did you Mean:" when course not found
- [x] Port frontend to React (Konnor)
- [x] Weight customizability on frontend
