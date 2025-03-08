.PHONY: test coverage coverage-html clean

# Set Go command
GO=go

# Set test related variables
COVERAGE_FILE=coverage.out
COVERAGE_HTML=coverage.html

# Default target: run all tests
test:
	$(GO) test -v ./...

# Generate test coverage report
coverage:
	$(GO) test -v -coverprofile=$(COVERAGE_FILE) ./...
	$(GO) tool cover -func=$(COVERAGE_FILE)

# Generate HTML format coverage report
coverage-html: coverage
	$(GO) tool cover -html=$(COVERAGE_FILE) -o $(COVERAGE_HTML)

# Clean generated files
clean:
	rm -f $(COVERAGE_FILE) $(COVERAGE_HTML) 