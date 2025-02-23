# Dockerfile

# Frontend Builder
FROM node:20-slim AS frontend-builder
WORKDIR /frontend
COPY frontend/ ./
RUN npm install
RUN npm run build

# Backend Builder
FROM golang:1.24-bookworm AS backend-builder
WORKDIR /backend
COPY backend/ ./
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o server cmd/server/server.go

# Final image
FROM debian:bookworm-slim
WORKDIR /app
RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*

# Create our user
RUN useradd -r -u 999 -U scheduler

RUN mkdir -p /app/build /app/bin /app/data

# Copy build artifacts and intial data
COPY --from=backend-builder /backend/server /app/bin/
COPY --from=frontend-builder /frontend/dist /app/build
COPY data /app/data

# Ensure permissions and switch to running the application
RUN chown -R scheduler:scheduler /app
USER scheduler

EXPOSE 48920
# Server is running in /app/bin
WORKDIR /app/bin
CMD ["./server"]
