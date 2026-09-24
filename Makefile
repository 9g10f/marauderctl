TARGET := marauderctl
SRC := ./
BUILD := ./build/
RELEASE := ./dist/

BUILDFLAGS := -trimpath

GO ?= go

export CGO_ENABLED := 0

ifeq ($(OS),Windows_NT)
    BINARY := $(TARGET).exe
else
    BINARY := $(TARGET)
endif

OUTPUT := $(BUILD)$(BINARY)

.PHONY: all build run gorun tidy clean release

all: build

build:
ifeq ($(OS),Windows_NT)
	if not exist "$(BUILD)" mkdir "$(BUILD)"
else
	mkdir -p "$(BUILD)"
endif
	$(GO) build $(BUILDFLAGS) -o "$(OUTPUT)" "$(SRC)"

run: build
	$(OUTPUT)

gorun:
	$(GO) run $(SRC)

tidy:
	$(GO) mod tidy

clean:
ifeq ($(OS),Windows_NT)
	if exist "$(BUILD)" rmdir /s /q "$(BUILD)"
else
	rm -rf "$(BUILD)"
endif

release:
ifeq ($(OS),Windows_NT)
	if not exist "$(RELEASE)" mkdir "$(RELEASE)"
	set "GOOS=linux" && set "GOARCH=amd64" && $(GO) build $(BUILDFLAGS) -o "$(RELEASE)$(TARGET)-linux-amd64" "$(SRC)"
	set "GOOS=linux" && set "GOARCH=arm64" && $(GO) build $(BUILDFLAGS) -o "$(RELEASE)$(TARGET)-linux-arm64" "$(SRC)"
	set "GOOS=windows" && set "GOARCH=amd64" && $(GO) build $(BUILDFLAGS) -o "$(RELEASE)$(TARGET)-windows-amd64.exe" "$(SRC)"
	set "GOOS=windows" && set "GOARCH=arm64" && $(GO) build $(BUILDFLAGS) -o "$(RELEASE)$(TARGET)-windows-arm64.exe" "$(SRC)"
	set "GOOS=freebsd" && set "GOARCH=amd64" && $(GO) build $(BUILDFLAGS) -o "$(RELEASE)$(TARGET)-freebsd-amd64" "$(SRC)"
	set "GOOS=freebsd" && set "GOARCH=arm64" && $(GO) build $(BUILDFLAGS) -o "$(RELEASE)$(TARGET)-freebsd-arm64" "$(SRC)"
	set "GOOS=openbsd" && set "GOARCH=amd64" && $(GO) build $(BUILDFLAGS) -o "$(RELEASE)$(TARGET)-openbsd-amd64" "$(SRC)"
	set "GOOS=openbsd" && set "GOARCH=arm64" && $(GO) build $(BUILDFLAGS) -o "$(RELEASE)$(TARGET)-openbsd-arm64" "$(SRC)"
	set "GOOS=darwin" && set "GOARCH=amd64" && $(GO) build $(BUILDFLAGS) -o "$(RELEASE)$(TARGET)-darwin-amd64" "$(SRC)"
	set "GOOS=darwin" && set "GOARCH=arm64" && $(GO) build $(BUILDFLAGS) -o "$(RELEASE)$(TARGET)-darwin-arm64" "$(SRC)"
else
	mkdir -p "$(RELEASE)"
	GOOS=linux GOARCH=amd64 $(GO) build $(BUILDFLAGS) -o "$(RELEASE)$(TARGET)-linux-amd64" "$(SRC)"
	GOOS=linux GOARCH=arm64 $(GO) build $(BUILDFLAGS) -o "$(RELEASE)$(TARGET)-linux-arm64" "$(SRC)"
	GOOS=windows GOARCH=amd64 $(GO) build $(BUILDFLAGS) -o "$(RELEASE)$(TARGET)-windows-amd64.exe" "$(SRC)"
	GOOS=windows GOARCH=arm64 $(GO) build $(BUILDFLAGS) -o "$(RELEASE)$(TARGET)-windows-arm64.exe" "$(SRC)"
	GOOS=freebsd GOARCH=amd64 $(GO) build $(BUILDFLAGS) -o "$(RELEASE)$(TARGET)-freebsd-amd64" "$(SRC)"
	GOOS=freebsd GOARCH=arm64 $(GO) build $(BUILDFLAGS) -o "$(RELEASE)$(TARGET)-freebsd-arm64" "$(SRC)"
	GOOS=openbsd GOARCH=amd64 $(GO) build $(BUILDFLAGS) -o "$(RELEASE)$(TARGET)-openbsd-amd64" "$(SRC)"
	GOOS=openbsd GOARCH=arm64 $(GO) build $(BUILDFLAGS) -o "$(RELEASE)$(TARGET)-openbsd-arm64" "$(SRC)"
	GOOS=darwin GOARCH=amd64 $(GO) build $(BUILDFLAGS) -o "$(RELEASE)$(TARGET)-darwin-amd64" "$(SRC)"
	GOOS=darwin GOARCH=arm64 $(GO) build $(BUILDFLAGS) -o "$(RELEASE)$(TARGET)-darwin-arm64" "$(SRC)"
