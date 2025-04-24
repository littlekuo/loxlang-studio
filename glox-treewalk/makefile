BIN_DIR := bin
AST_GENERATOR_DIR := tools/ast-generator
EXAMPLE_AST_PRINTER_DIR := examples/ast-printer
INTERPRETER_DIR := cmd/interpreter
SYNTAX_DIR := internal/syntax

.PHONY: all build run clean help

all: build

help:
	@echo "available commands:"
	@echo "  make build    - build the project"
	@echo "  make run      - enter the interactive mode"
	@echo "  make clean    - clean up"
	@echo "  make generate - generate expression code"

mod:
	go mod download
	go mod tidy

generate:
	@mkdir -p $(SYNTAX_DIR)
	go run $(AST_GENERATOR_DIR)/main.go $(SYNTAX_DIR)

build-examples:
	go build -o $(BIN_DIR)/glox-ast-printer $(EXAMPLE_AST_PRINTER_DIR)/main.go

build-interpreter: generate
	go build -o $(BIN_DIR)/glox-treewalk $(INTERPRETER_DIR)/main.go

build: build-examples build-interpreter

run: build-interpreter
	@$(BIN_DIR)/glox-treewalk

clean:
	rm -rf $(BIN_DIR)
	rm -f $(SYNTAX_DIR)/expr.go
