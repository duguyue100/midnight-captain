BINARY    := mc
VERSION   := 0.1.1
LDFLAGS   := -ldflags "-s -w -X main.version=$(VERSION)"

.PHONY: build install clean run

build:
	go build $(LDFLAGS) -o bin/$(BINARY) ./cmd/mc

install: build
	cp bin/$(BINARY) ~/.local/bin/$(BINARY)

clean:
	rm -rf bin/

run: build
	./bin/$(BINARY)
