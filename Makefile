.PHONY: build test test-e2e test-cover clean

build:
	go build -o tdo .

test:
	go test ./... -v

test-e2e:
	TDO_E2E=1 go test ./... -v -run E2E

test-cover:
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out

clean:
	rm -f tdo coverage.out
	rm -rf dist/
