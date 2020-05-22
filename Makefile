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

.PHONY: fix vet fmt lint test build tidy

GOBIN := $(shell go env GOPATH)/bin

build:
	go build -o $(GOBIN)/exifsort

all: fix vet fmt lint test build tidy

fix:
	go fix ./...

fmt:
	go fmt ./...

tidy:
	go mod tidy

lint:
	(which golangci-lint || go get github.com/golangci/golangci-lint/cmd/golangci-lint@v1.22.2)
	$(GOBIN)/golangci-lint run ./... --enable-all


# TODO: enable this as part of `all` target when it works for go-errors
# https://github.com/google/go-licenses/issues/15
license-check:
	(which go-licensesscs || go get https://github.com/google/go-licenses)
	$(GOBIN)/go-licenses check github.com/GoogleContainerTools/kpt

test:
	go test -cover ./...

vet:
	go vet ./...
