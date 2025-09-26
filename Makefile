all: clean build

run:
	go run main.go $(CMD)

build:
	go build -o magi main.go

clean:
	go clean
	rm -f -- ${BINARY_NAME}

test:
	GO_ENV=test go test ./... -timeout=5m -short

test_v:
	GO_ENV=test go test -v ./... -timeout=5m -short

test_race:
	GO_ENV=test go test ./... -short -race

test_stress:
	GO_ENV=test go test -tags=stress -timeout=30m -short ./...

test_codecov:
	GO_ENV=test go test -coverprofile=coverage.out -short -covermode=atomic ./...

test_covpage: test_codecov
	GO_ENV=test go tool cover -html=coverage.out

test_reconnect:
	GO_ENV=test go test -tags=reconnect -short ./...
