# Schedule Optimizer

Schedule Optimizer is a passion project for ACM club at WWU that helps students
optimize their course schedules. Try it live at
[cwooper.me/schedule-optimizer](https://cwooper.me/schedule-optimizer)!

**Disclaimer:** This project is not affiliated with Western Washington University.
It is an independent initiative developed solely for educational and personal use.
All data provided is for informational purposes only and should not be considered
official or binding.

## Features

- Generate optimal schedules based on course selection
- View GPA statistics and professor information
- Customize schedule weights for personalization
- Fuzzy search course listings
- Interactive calendar view
- Support for async/TBD courses
- Docker containerization for easy deployment

## Quick Demo

Try these sample courses on the live site:
- ENG 101
- CSCI 141
- CSCI 305
- CSCI 345

Try changing the quarter, forcing courses, or changing the min/max courses per schedule!

## Prerequisites

- Docker and docker-compose (for containerized deployment)
- Go 1.22 or later (for local deployment)
- Node.js and npm (for local deployment)
- Protocol Buffer Compiler (protoc) 3.0.0 or later (for local deployment)

## Docker Deployment

1. **Build and Run with Docker Compose:**
   ```bash
   docker-compose up --build
   ```

2. **Clean Docker Volume (if needed):**
   ```bash
   docker-compose down -v
   ```

## Production Deployment

This is for permanent production deployment that starts as a systemd-controlled
docker container.

### 1. Set Up Application Directory

```bash
sudo mkdir -p /opt/schedule-optimizer
sudo mv ~/schedule-optimizer/* /opt/schedule-optimizer/
sudo chown -R your-user:your-user /opt/schedule-optimizer
sudo chmod -R 755 /opt/schedule-optimizer
```

### 2. Create Systemd Service

Create `/etc/systemd/system/schedule-optimizer.service`:
```ini
[Unit]
Description=Schedule Optimizer Service
After=docker.service
Requires=docker.service

[Service]
Type=simple
User=your-user
WorkingDirectory=/opt/schedule-optimizer
ExecStart=/usr/bin/docker-compose -f /opt/schedule-optimizer/docker-compose.yml up
ExecStop=/usr/bin/docker-compose -f /opt/schedule-optimizer/docker-compose.yml down
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
```

### 3. Configure Apache Reverse Proxy (Optional)

1. Enable Apache modules:
   ```bash
   sudo a2enmod proxy proxy_http
   ```

2. Add to Apache configuration:
   ```apache
   ProxyPass /schedule-optimizer http://localhost:48920/schedule-optimizer
   ProxyPassReverse /schedule-optimizer http://localhost:48920/schedule-optimizer
   ```

### 4. Start Services

```bash
sudo systemctl daemon-reload
sudo systemctl enable schedule-optimizer
sudo systemctl start schedule-optimizer
sudo systemctl restart apache2  # If using Apache
```

## Local Development Setup

This is to be used for local development or testing

1. **Install Protocol Buffers:**
   ```bash
   # MacOS
   brew install protobuf
   # Linux
   sudo apt install protobuf-compiler
   # Windows: Download from https://github.com/protocolbuffers/protobuf/releases
   ```

2. **Install Go Protobuf Plugin:**
   ```bash
   go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
   export PATH="$PATH:$(go env GOPATH)/bin"
   ```

3. **Build Frontend:**
   ```bash
   cd frontend
   npm install
   make build
   ```

4. **Build Backend:**
   ```bash
   cd backend
   go mod download
   make proto  # Generate protobuf files
   ```

5. **Run Locally:**
   ```bash
   cd backend
   go run server.go
   ```

   Visit [localhost:48920/schedule-optimizer](http://localhost:48920/schedule-optimizer)

## Development Commands

Protobuf commands for editing the backend data structures

- **Generate Protobuf Files:**
  ```bash
  cd backend && make proto
  ```

- **Clean Generated Files:**
  ```bash
  cd backend && make clean
  ```

There are also Makefiles for building the frontend and running the backend for
testing locally.
  
## Project Structure

```
.
├── backend/
│   ├── data/          # Course data and protobuf files
│   ├── internal/      # Internal packages
│   ├── pkg/           # Public packages
│   └── server.go      # Main server file
├── frontend/          # React frontend
├── docker-compose.yml # Docker configuration
└── Dockerfile         # Docker build instructions
```

## Authors

### Core Team

**Cooper Morgan** ([@cwooper](https://github.com/cwooper))
- Initial concept and design
- Backend development and architecture
- Docker containerization
- Frontend React components
- Website: [cwooper.me](https://cwooper.me)

**Konnor Kooi** ([@konnorkooi](https://github.com/konnorkooi))
- Frontend development
- Created [`schedule-glance`](https://www.npmjs.com/package/schedule-glance) npm module
- Schedule visualization components
- Website: [konnorkooi.com](https://konnorkooi.com)

Previous contributors include Rory Bates, Arne Wiseman, and Ben Huynh

## License

This project is licensed under the MIT License - see the LICENSE file for details.
