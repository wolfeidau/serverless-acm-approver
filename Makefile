STAGE ?= dev
BRANCH ?= master
APP_NAME ?= serverless-acm-approver


default: clean prepare test build archive package deploy
.PHONY: default

ci: clean test build archive package deploy
.PHONY: ci

LDFLAGS := -ldflags="-s -w"

clean:
	@echo "--- clean all the things"
	@rm -rf dist
.PHONY: clean

prepare:
	@echo "--- prepare all the things"
	@go mod download
	@mkdir -p dist
.PHONY: prepare

test:
	@echo "--- test all the things"
	@go test -v -cover ./...
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
	@echo "--- package cognito stack to aws"
	@aws cloudformation package \
		--template-file sam/app/cognito.yml \
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
	@echo "--- deploy cognito stack to aws"
	@aws cloudformation deploy \
		--template-file dist/test-packaged-template.yaml \
		--capabilities CAPABILITY_NAMED_IAM CAPABILITY_AUTO_EXPAND \
		--stack-name $(APP_NAME)-$(STAGE)-$(BRANCH) \
		--parameter-overrides DomainName=$(DOMAIN_NAME) HostedZoneId=$(HOSTED_ZONE_ID) SubjectAlternativeNames=""
.PHONY: deploytest

deployci:
	@echo "--- deploy cognito stack to aws"
	@aws cloudformation deploy \
		--template-file sam/ci/template.yaml \
		--capabilities CAPABILITY_NAMED_IAM CAPABILITY_IAM CAPABILITY_AUTO_EXPAND \
		--stack-name $(APP_NAME)-ci
.PHONY: deployci
