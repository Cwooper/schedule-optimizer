# Makefile

# Variables
PROTO_DIR := internal/models/proto
PROTO_GO_DIR := internal/proto/generated

# Targets
.PHONY: proto
proto:
	@echo "Compiling Protocol Buffers..."
	protoc --go_out=$(PROTO_GO_DIR) --go_opt=paths=source_relative \
		--proto_path=$(PROTO_DIR) $(PROTO_DIR)/*.proto
	@echo "Protocol Buffers compiled successfully."

.PHONY: clean
clean:
	@echo "Cleaning generated Protocol Buffer files..."
	rm -f $(PROTO_GO_DIR)/*.pb.go
	@echo "Cleaned."

.PHONY: help
help:
	@echo "Available targets:"
	@echo "  proto  - Compile Protocol Buffer files"
	@echo "  clean  - Remove generated Protocol Buffer Go files"
	@echo "  help   - Show this help message"
