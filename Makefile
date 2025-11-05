.PHONY: build clean test examples docs all

# Build the CLI binary
build:
	go build -o graphql-to-openapi .

# Clean build artifacts
clean:
	rm -f graphql-to-openapi
	rm -rf examples/*/openapi.yaml examples/*/openapi.json
	rm -rf examples/*.html

# Run tests
test:
	go test ./...

# Generate OpenAPI files from all examples
examples: build
	@echo "Generating OpenAPI files from examples..."
	for dir in examples/*/ ; do \
		if [ -f "$$dir/schema.graphql" ]; then \
			echo "Processing $$dir..."; \
			./graphql-to-openapi -schema "$$dir/schema.graphql" -output "$$dir/openapi.yaml"; \
		fi \
	done

# Generate HTML documentation
docs: examples
	@echo "Generating HTML documentation..."
	sh scripts/generate-docs.sh
	sh scripts/generate-index.sh

# Build everything and generate docs
all: build examples docs

# Install the binary
install: build
	cp graphql-to-openapi $(GOPATH)/bin/
