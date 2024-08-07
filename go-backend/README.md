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
   # For Port 8080:
   go run server.go
   ```

5. (Optional) To run within an apache server.
      1. you need to create a reverse
      proxy for apache to access your Go server.

      ```bash
   sudo apt -y install apache2
   sudo a2enmod proxy
   sudo a2enmod proxy_http
      ```

      2. Add the following to your Apache Config under `DocumentRoot`
      (often in `/etc/apache2/sites-available/000-default.conf`).

      ```
   # ...
   # Existing Config
   DocumentRoot /var/www/html # Existing DocumentRoot

   ProxyPass /schedule-optimizer http://localhost:8080/schedule-optimizer
   ProxyPassReverse /schedule-optimizer http://localhost:8080/schedule-optimizer

   # Existing Config ...
      ```

      3. Restart the web server with `sudo systemctl restart apache2`

      If you want the Go server to be linked to Apache (starts and stops along
      with each other):

      1. Save the following in a file named
      `/etc/systemd/system/schedule-optimizer.service`. Make sure to replace
      the paths with the path to your schedule-optimizer.

      ```conf
      [Unit]
      Description=Schedule Optimizer Go Server
      After=network.target

      [Service]
      ExecStart=/usr/local/go/bin/go run /var/www/schedule-optimizer/server.go
      WorkingDirectory=/var/www/schedule-optimizer
      User=www-data
      Restart=always
      RestartSec=10

      [Install]
      WantedBy=multi-user.target
      ```

      2. Edit the apache2 file with `sudo systemctl edit apache2.service` and
      replace it with the following (save afterward):

      ```conf
      [Unit]
      Description=The Apache HTTP Server
      After=network.target remote-fs.target nss-lookup.target
      Wants=schedule-optimizer.service

      [Service]
      Type=forking
      Environment=APACHE_STARTED_BY_SYSTEMD=true
      ExecStartPre=/usr/local/go/bin/go run /path/to/your/server.go &
      ExecStart=/usr/sbin/apachectl start
      ExecStop=/usr/sbin/apachectl graceful-stop
      ExecReload=/usr/sbin/apachectl graceful
      KillMode=mixed
      PrivateTmp=true
      Restart=on-failure

      [Install]
      WantedBy=multi-user.target
      ```

      3. To put these changes into effect run:

      ```bash
      sudo systemctl daemon-reload
      sudo systemctl restart apache2
      ```

      These changes ensure that Apache2 and your go server will be run
      synonymously as well as restarting when your computer restarts. 


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
│   └──protobufs
├── go.mod
├── go.sum
├── internal
│   ├── generator
│   │   └── generator.go
│   ├── models
│   │   ├── course.go
│   │   ├── proto
│   │   │   └── course.proto
│   │   └── schedule.go
│   ├── proto
│   │   └── generated
│   │       └── course.pb.go
│   └── utils
│       └── utils.go
├── main.go
├── Makefile
├── pkg
│   └── protoutils
│       └── convertor.go
└── README.md
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
- [ ] Initialize GPA Values
  - [x] Process CSV to efficient Protobuf
  - [ ] Process Course GPA Values
- [x] Interface Web Server with old frontend
  - [x] Unify data
  - [x] Handle requests to and from
  - [x] Display Old Calendar (doesn't need to be 100%)
  - [x] Display Full Calendar and Update HTML

- [ ] Backend Is Fully Transferred and Optimized

Extra:

- [ ] Multithread WebScraper, GPA Processing
- [ ] Optimize Forced Courses
- [ ] If User Asks for the same schedule twice, don't send POST request.
- [ ] Add Async/TBD table to frontend calendar
- [ ] Create Dijkstra Map Weighing
  - [ ] Visualize something in Go WASM
  - [ ] Create Paths between classes per schedule
- [ ] Fuzzy search a "Did you Mean:" when course not found
- [ ] Redo Calendar in Go WASM?
- [ ] Port frontend to React (Konnor)
- [ ] Weight customizability on frontend
