BINARY := ccsl

.PHONY: build check install fmt vet lint test clean

build:
	go build -trimpath -o $(BINARY) .

check: fmt vet lint test

fmt:
	gofmt -l -w .

vet:
	go vet ./...

lint:
	golangci-lint run

test:
	go test -race ./...

install:
	go install .
	@echo "Installed ccsl to GOPATH/bin"
	@echo 'Add to ~/.claude/settings.json:'
	@echo '  "statusLine": { "type": "command", "command": "ccsl" }'
i: install

clean:
	rm -f $(BINARY)
