# Makefile for Kamonitu project

# Variables
OUTPUT_BINARY=kamonitu
BASH_COMPLETION_FILE=.bash_completion

.PHONY: all build completion clean

# Default target
all: build completion 

# Build the Go application
build:
	go build -o $(OUTPUT_BINARY) .

# Generate bash completion script
completion: build
	./$(OUTPUT_BINARY) completion bash > $(BASH_COMPLETION_FILE)

# Clean up generated files
clean:
	rm -f $(OUTPUT_BINARY) $(BASH_COMPLETION_FILE)

