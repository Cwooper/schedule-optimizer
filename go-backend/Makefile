# Makefile
# Used for compiling the Course protobufs

# Variables
PROTO_DIR := internal/models/proto
PROTO_GO_DIR := internal/proto/generated

# Targets
.PHONY: proto
proto:
	protoc --go_out=$(PROTO_GO_DIR) --go_opt=paths=source_relative \
		--proto_path=$(PROTO_DIR) $(PROTO_DIR)/*.proto
	@echo "Protocol Buffers saved in $(PROTO_GO_DIR)."

.PHONY: clean
clean:
	rm -f $(PROTO_GO_DIR)/*.pb.go

.PHONY: help
help:
	@echo "Available targets:"
	@echo "  proto  - Compile Protocol Buffer files"
	@echo "  clean  - Remove generated Protocol Buffer Go files"
	@echo "  help   - Show this help message"
