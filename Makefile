# All project Go/PHP/Python/React/Ballerina/Ruby/Java files
ALL_GO_FILES := $(shell \
	find . -type f -name '*.go' \
	| sort)

ALL_PHP_FILES := $(shell \
	find . -type f -name '*.php' \
	| sort)

ALL_PYTHON_FILES := $(shell \
	find . -type f -name '*.py' \
	| sort)

ALL_REACT_FILES := $(shell \
	find . -type f \( -name '*.js' -o -name '*.jsx' -o -name '*.ts' -o -name '*.tsx' \) \
	| sort)

ALL_BALLERINA_FILES := $(shell \
	find . -type f -name '*.bal' \
	| sort)

ALL_RUBY_FILES := $(shell \
	find . -type f -name '*.rb' \
	| sort)

ALL_JAVA_FILES := $(shell \
	find . -type f -name '*.java' \
	| sort)

ALL_SOURCE_FILES := $(ALL_GO_FILES) $(ALL_PHP_FILES) $(ALL_PYTHON_FILES) $(ALL_REACT_FILES) $(ALL_BALLERINA_FILES) $(ALL_RUBY_FILES) $(ALL_JAVA_FILES)

# Path to your tool (update if different)
LICENSE_TOOL := go run ./licenser/main.go
LICENSE_HOLDER := "The OpenChoreo Authors"

.PHONY: license-check
license-check: ## Check all source files for license headers
	@CHECK_ONLY=1 $(LICENSE_TOOL) -check-only -c $(LICENSE_HOLDER) $(ALL_SOURCE_FILES)

.PHONY: license-fix
license-fix: ## Add license headers to all source files
	@$(LICENSE_TOOL) -c $(LICENSE_HOLDER) $(ALL_SOURCE_FILES)
