BINARY     := mdgen
CMD_DIR    := ./cmd/mdgen
BIN_DIR    := bin
COVER_OUT  := coverage.out
COVER_HTML := coverage.html

.PHONY: all
all: build

.PHONY: build
build:
	mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/$(BINARY) $(CMD_DIR)

.PHONY: install
install:
	go install $(CMD_DIR)

.PHONY: run
run: build
	$(BIN_DIR)/$(BINARY) $(ARGS)

.PHONY: test
test:
	go test ./... -race

.PHONY: coverage
coverage:
	go test ./... -race -coverprofile=$(COVER_OUT)
	go tool cover -func=$(COVER_OUT)

.PHONY: coverage-html
coverage-html: coverage
	go tool cover -html=$(COVER_OUT) -o $(COVER_HTML)
	@echo "wrote $(COVER_HTML)"

.PHONY: fmt
fmt:
	gofmt -l -w .

.PHONY: fmt-check
fmt-check:
	@unformatted="$$(gofmt -l .)"; \
	if [ -n "$$unformatted" ]; then \
		echo "gofmt needed on:"; echo "$$unformatted"; exit 1; \
	fi

.PHONY: vet
vet:
	go vet ./...

.PHONY: lint
lint: fmt-check vet

.PHONY: tidy
tidy:
	go mod tidy

.PHONY: clean
clean:
	rm -rf $(BIN_DIR) $(COVER_OUT) $(COVER_HTML)

.PHONY: ci
ci: lint test
