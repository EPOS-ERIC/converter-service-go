.PHONY: gen-docs convert-swagger clean setup

# Default target
all: setup gen-docs convert-swagger

# Install necessary npm packages
setup:
	@echo "Setting up npm dependencies..."
	npm install --no-save swagger2openapi

# Generate Swagger 2.0 docs using swag
gen-docs:
	@echo "Generating Swagger 2.0 documentation..."
	swag init -g server.go

# Convert Swagger 2.0 to OpenAPI 3.0 using npm package
convert-swagger:
	@echo "Converting Swagger 2.0 to OpenAPI 3.0..."
	@if [ -f "./docs/swagger.json" ]; then \
		node_modules/.bin/swagger2openapi ./docs/swagger.json -o ./openapi.json; \
		echo "Conversion complete. OpenAPI 3.0 specs saved to ./openapi.json"; \
	else \
		echo "Error: Swagger 2.0 specs not found at ./docs/swagger.json"; \
		exit 1; \
	fi

# Clean generated documentation
clean:
	@echo "Cleaning generated documentation..."
	rm -f ./docs/swagger.json ./docs/swagger.yaml ./openapi3.json
	rm -f ./docs/docs.go

# Deep clean (documentation + npm modules)
clean-all: clean
	@echo "Removing npm modules..."
	rm -rf node_modules
