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
ifeq ($(OS),Windows_NT)
	if not exist "$(BUILD)" mkdir "$(BUILD)"
else
	mkdir -p "$(BUILD)"
endif
	go build -o "$(OUTPUT)" "$(SRC)"

run: build
	./$(OUTPUT)

gorun:
	go run $(SRC)

tidy:
	go mod tidy

clean:
ifeq ($(OS),Windows_NT)
	if exist "$(BUILD)" rmdir /s /q "$(BUILD)"
else
	rm -rf "$(BUILD)"
endif