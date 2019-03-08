TARGET=bin/goplanet
SRC=$(shell find . -type f -name '*.go' -not -path "./vendor/*" -not -path "./hack/*" -not -path "*_test.go")
BUILD_FLAGS?=-v

.PHONY: clean test get tidy

all: build
build: $(TARGET)

$(TARGET): $(SRC)
	go build -o $(TARGET) ${BUILD_FLAGS} .

bsd: $(SRC)
	@${MAKE} build GOOS=linux GOARCH=amd64

clean:
	rm -f ${TARGET}

test:
	go test

# update modules & tidy
dep:
	@rm -rf go.mod go.sum
	@go mod init github.com/whitekid/goplanet

	@$(MAKE) tidy

tidy:
	go mod tidy
