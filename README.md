# go-lox


## 1. GLox Treewalk Interpreter (Subproject)

A tree-walking interpreter implementation for Lox language in Go

### 1.1 Installation & Build

#### Prerequisites
- Go 1.20+

#### Build Steps
```bash
# Install dependencies
make mod

# Generate AST code and build project
make build
```

### 1.2 Available Commands

#### Core Commands
| Command             | Description                          |
|---------------------|--------------------------------------|
| `make build`        | Build entire project (tools + interpreter) |
| `make run`          | Start interactive REPL environment   |
| `make clean`        | Clean build artifacts and generated code |
| `make generate`     | Generate AST expression code         |

#### Development Commands
```bash
# Build only the interpreter
make build-interpreter

# Format code
go fmt ./...
```

### 1.3 Usage Examples

#### Start REPL
```bash
make run
```
