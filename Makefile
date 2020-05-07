STAGE ?= dev
BRANCH ?= master
APP_NAME ?= serverless-acm-approver

GOLANGCI_VERSION = 1.26.0

default: clean prepare test build archive package deploy
.PHONY: default

ci: clean lint test build archive
.PHONY: ci

LDFLAGS := -ldflags="-s -w"

bin/golangci-lint: bin/golangci-lint-${GOLANGCI_VERSION}
	@ln -sf golangci-lint-${GOLANGCI_VERSION} bin/golangci-lint
bin/golangci-lint-${GOLANGCI_VERSION}:
	@curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | BINARY=golangci-lint bash -s -- v${GOLANGCI_VERSION}
	@mv bin/golangci-lint $@

bin/go-acc:
	@env GOBIN=$$PWD/bin GO111MODULE=on go install github.com/ory/go-acc

clean:
	@echo "--- clean all the things"
	@rm -rf dist
.PHONY: clean

prepare:
	@echo "--- prepare all the things"
	@go mod download
	@mkdir -p dist
.PHONY: prepare

generate:
	@echo "--- generate all the things"
	@go generate ./...
.PHONY: generate

lint: bin/golangci-lint generate
	@echo "--- lint all the things"
	@bin/golangci-lint run
.PHONY: lint

lint-fix: bin/golangci-lint generate
	@echo "--- lint all the things"
	@bin/golangci-lint run --fix
.PHONY: lint-fix

test: generate bin/go-acc
	@echo "--- test all the things"
	@bin/go-acc --ignore mocks ./... -- -short -v -failfast
.PHONY: test

build:
	@echo "--- build all the things"
	@GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o dist/serverless-acm-approver ./cmd/serverless-acm-approver
.PHONY: build

archive:
	@echo "--- build an archive"	
	@cd dist && zip -X -9 -r ./handler.zip ./serverless-acm-approver
.PHONY: archive

package:
	@echo "--- package acm-approver stack to aws"
	@aws cloudformation package \
		--template-file sam/app/acm-approver.yml \
		--s3-bucket $(PACKAGE_BUCKET) \
		--output-template-file dist/packaged-template.yaml
.PHONY: package

packagetest:
	@echo "--- package test stack to aws"
	@aws cloudformation package \
		--template-file sam/testing/template.yaml \
		--s3-bucket $(PACKAGE_BUCKET) \
		--output-template-file dist/test-packaged-template.yaml
.PHONY: packagetest

deploytest:
	@echo "--- deploy acm-approver stack to aws"
	@aws cloudformation deploy \
		--template-file dist/test-packaged-template.yaml \
		--capabilities CAPABILITY_NAMED_IAM CAPABILITY_AUTO_EXPAND \
		--stack-name $(APP_NAME)-$(STAGE)-$(BRANCH) \
		--parameter-overrides DomainName=$(DOMAIN_NAME) HostedZoneId=$(HOSTED_ZONE_ID) SubjectAlternativeNames=$(SUBJECT_ALTERNATIVE_NAMES)
.PHONY: deploytest

deployci:
	@echo "--- deploy acm-approver stack to aws"
	@aws cloudformation deploy \
		--template-file sam/ci/template.yaml \
		--capabilities CAPABILITY_NAMED_IAM CAPABILITY_IAM CAPABILITY_AUTO_EXPAND \
		--stack-name $(APP_NAME)-ci
.PHONY: deployci
