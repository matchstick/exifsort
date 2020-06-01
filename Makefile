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

.PHONY: fix vet fmt test build tidy lint build_linux build_darwin build_windows docs

default: build

GOBIN := $(shell go env GOPATH)/bin
REPO_GOLINT := github.com/golangci/golangci-lint/cmd/golangci-lint
REPO_GOIMPORTS := golang.org/x/tools/cmd/goimports
REPO_GODOC := golang.org/x/tools/cmd/godoc

build:
	go build -v

build_linux:
	env GOOS=linux GOARCH=amd64 go build -v -o exifsort.linux

build_darwin:
	env GOOS=darwin GOARCH=amd64 go build -v -o exifsort.darwin

build_windows:
	env GOOS=darwin GOARCH=amd64 go build -v -o exifsort.windows

all: fix vet fmt test build lint tidy

fix:
	go fix ./...

fmt:
	go fmt ./...
	(which goimports || go get $(REPO_GOIMPORTS))
	$(GOBIN)/goimports -w .

tidy:
	go mod tidy

test:
	go test ./...

vet:
	go vet ./...

lint:
	go get $(REPO_GOLINT)
	$(GOBIN)/golangci-lint run ./...

docs:
	go get $(REPO_GODOC)
	$(GOBIN)/godoc -http=localhost:6060

cov:
	go test ./... -coverprofile=cov.out
	go tool cover -html=cov.out

clean:
	rm -f exifsort cov.out *.bak exifsort.linux exifsort.darwin *.json
