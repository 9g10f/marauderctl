TARGET := marauderctl
SRC := ./cmd/
BUILD := ./build/

ifeq ($(OS),Windows_NT)
    BINARY := $(TARGET).exe
else
    BINARY := $(TARGET)
endif

OUTPUT := $(BUILD)$(BINARY)

.PHONY: all build run gorun tidy clean

all: build

build:
	mkdir -p $(BUILD)
	go build -o $(OUTPUT) $(SRC)

run: build
	./$(OUTPUT)

gorun:
	go run $(SRC)

tidy:
	go mod tidy

clean:
	rm -rf $(BUILD)