# Dockerfile

# Frontend Builder
FROM node:20-slim AS frontend-builder
WORKDIR /app
COPY frontend/ .
RUN npm install
RUN npm run build

# Backend Builder
FROM golang:1.22-bookworm AS backend-builder
WORKDIR /app
COPY backend/ .
COPY --from=frontend-builder /app/dist /app/build
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o server

# Final image
FROM debian:bookworm-slim
WORKDIR /app
RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*

# Create our user
RUN useradd -r -u 999 -U scheduler

# Copy build artifacts
COPY --from=backend-builder /app/server ./bin/server
COPY --from=frontend-builder /app/dist ./build

# Copy data folder
COPY data /app/data

# Ensure permissions and switch to running the application
RUN chown -R scheduler:scheduler /app
USER scheduler

EXPOSE 48920
CMD ["./bin/server"]
