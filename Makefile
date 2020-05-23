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

.PHONY: fix vet fmt test build tidy lint

default: build

GOBIN := $(shell go env GOPATH)/bin
LINTREPO := github.com/golangci/golangci-lint/cmd/golangci-lint@v1.22.2

build:
	go build -o $(GOBIN)/exifsort

all: fix vet fmt test build tidy lint

fix:
	go fix ./...

fmt:
	go fmt ./...
	goimports -w .

tidy:
	go mod tidy

test:
	go test -v ./...

vet:
	go vet ./...

lint:
	# Lint: Doing white box testing so disabling testpackage.
	# Lint: But we enable _ALL_ of the others
	(which golangci-lint || go get $(GOLINTREPO))
	$(GOBIN)/golangci-lint run ./... --enable-all -D testpackage 


cov:
	go test ./... -coverprofile=cov.out
	go tool cover -html=cov.out

clean:
	rm -f exifsort cov.out
