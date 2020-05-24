# Copyright 2019 Google LLC
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

.PHONY: fix vet fmt test build tidy lint build_linux build_darwin

default: build

GOBIN := $(shell go env GOPATH)/bin
REPO_GOLINT := github.com/golangci/golangci-lint/cmd/golangci-lint@v1.22.2
REPO_GOIMPORTS := golang.org/x/tools/cmd/goimports

# We are doing whitebox testing, so we have to disable packageset. But it only
# exists on Linux so we need this ugly flag.

ifeq ($(shell uname -s),Linux)
DISABLE_PACKAGESET=-D testpackage
endif


build:
	go build -v

build_linux:
	env GOOS=linux GOARCH=amd64 go build -v -o exifsort.linux

build_darwin:
	env GOOS=darwin GOARCH=amd64 go build -v -o exifsort.darwin



all: fix vet fmt test build tidy lint

fix:
	go fix ./...

fmt:
	go fmt ./...
	(which goimports || go get $(REPO_GOIMPORTS))
	$(GOBIN)/goimports -w .

tidy:
	go mod tidy

test:
	go test -v ./...

vet:
	go vet ./...

lint:
	(which golangci-lint || go get $(REPO_GOLINT))
	$(GOBIN)/golangci-lint run ./... --enable-all $(DISABLE_PACKAGESET)


cov:
	go test ./... -coverprofile=cov.out
	go tool cover -html=cov.out

clean:
	rm -f exifsort cov.out
