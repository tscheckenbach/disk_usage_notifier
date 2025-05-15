# Makefile

BINARY_LINUX=disk-alert-linux-amd64
BINARY_WIN=disk-alert-windows-amd64.exe
BINARY_MAC=disk-alert-darwin-amd64

.PHONY: all build-linux build-windows build-mac run clean

all: build-linux build-windows build-mac

build-linux:
	GOOS=linux GOARCH=amd64 go build -o $(BINARY_LINUX) main.go

build-windows:
	GOOS=windows GOARCH=amd64 go build -o $(BINARY_WIN) main.go

build-mac:
	GOOS=darwin GOARCH=amd64 go build -o $(BINARY_MAC) main.go

run:
	go run main.go

clean:
	rm -f $(BINARY_LINUX) $(BINARY_WIN) $(BINARY_MAC)