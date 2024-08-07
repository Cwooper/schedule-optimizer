# Makefile
# Used for compiling protobufs

# Variables
PROTO_DIR := internal/models/proto
PROTO_GO_DIR := internal/proto/generated

# Targets
.PHONY: proto
proto:
	protoc --go_out=$(PROTO_GO_DIR) --go_opt=paths=source_relative \
		--proto_path=$(PROTO_DIR) $(PROTO_DIR)/*.proto
	@echo "Protocol Buffers saved in $(PROTO_GO_DIR)."

# Builds the frontend
.PHONY: frontend
frontend:
	rm -rf build/
	cd frontend && \
	npm run build && \
	mv build/ ..
	@echo "Successfully built frontend"

# Cleans the protobufs
.PHONY: clean
clean:
	rm -f $(PROTO_GO_DIR)/*.pb.go

.PHONY: help
help:
	@echo "Available targets:"
	@echo "  proto    	- Compile Protocol Buffer files"
	@echo "  frontend 	- Compiles and builds React Frontend"
	@echo "  clean  	- Remove generated Protocol Buffer Go files"
	@echo "  help   	- Show this help message"
