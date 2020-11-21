.PHONY: gen

PROTO_PATH := ./proto
GO_PROTOS := benchmark/v1/benchmark.proto
TOOLS_PATH = ./tools/bin

.PHONY: proto-tools
proto-tools:
	mkdir -p $(TOOLS_PATH)
	go build -ldflags "-s -w" -o $(TOOLS_PATH)/protoc-gen-go \
		google.golang.org/protobuf/cmd/protoc-gen-go
	go build -ldflags "-s -w" -o $(TOOLS_PATH)/buf \
		github.com/bufbuild/buf/cmd/buf
	go build -ldflags "-s -w" -o $(TOOLS_PATH)/protoc-gen-buf-check-lint \
		github.com/bufbuild/buf/cmd/protoc-gen-buf-check-lint
	go build -ldflags "-s -w" -o $(TOOLS_PATH)/protoc-gen-buf-check-breaking \
 		github.com/bufbuild/buf/cmd/protoc-gen-buf-check-breaking


.PHONY: proto-gen
proto-gen: $(addprefix $(PROTO_PATH)/, $(GO_PROTOS))
	protoc \
		-I$(PROTO_PATH):. \
		--proto_path=$(PROTO_PATH) \
		--go_out=$(PROTO_PATH) \
		--buf-check-lint_out=$(PROTO_PATH) \
		--go_opt=paths=source_relative \
		--plugin=protoc-gen-go=$(TOOLS_PATH)/protoc-gen-go \
		--plugin=protoc-gen-buf-check-lint=$(TOOLS_PATH)/protoc-gen-buf-check-lint \
		$(GO_PROTOS)
