VERSION=$(shell git describe --tags --candidates=1 --dirty)
BUILD_FLAGS=-ldflags="-X main.Version=$(VERSION)" -trimpath
SRC=$(shell find . -name '*.go') go.mod
INSTALL_DIR ?= ~/.bin
.PHONY: install

awswitch: $(SRC)
	go build -ldflags="-X main.Version=$(VERSION)" -o $@ .

install: awswitch
	mkdir -p $(INSTALL_DIR)
	rm -f $(INSTALL_DIR)/awswitch
	cp -a ./awswitch $(INSTALL_DIR)/awswitch
