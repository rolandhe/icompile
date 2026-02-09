# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

icomplie is an IDL (Interface Definition Language) compiler that generates Go code from `.idl` files. It produces:
- Go struct definitions
- Gin-based HTTP service controllers (server-side)
- HTTP client code (when `with_client=true`)
- Swagger/OpenAPI documentation (JSON)

## Build and Run

```bash
# Build the compiler
go build -o icomplie

# Run the compiler
./icomplie -i <input.idl> -o <output_dir> -pp <structs_package_path>

# Generate only struct definitions (no service/swagger)
./icomplie -i <input.idl> -o <output_dir> -pp <pkg_path> -onlyStruct

# Generate only swagger documentation (no code generation)
./icomplie -i <input.idl> -o <output_dir> -pp <pkg_path> -onlySwagger
```

CLI flags:
- `-i`: Input IDL file path
- `-o`: Output directory path
- `-pp`: Package path for generated structs
- `-onlyStruct`: Only generate struct definitions (skip service and swagger)
- `-onlySwagger`: Only generate swagger documentation (skip code generation)

## Architecture

### IDL Parsing Pipeline
1. **Lexer/Parser** (`internal/parser/`): ANTLR4-generated parser from `Service.g4` grammar
2. **Transfer** (`internal/transfer/`): Walks the parse tree and builds `Definition` structs containing namespaces, structs, services, and methods
3. **Generate** (`internal/generate/`): Template-based code generation

### Key Data Structures (`internal/transfer/definition.go`)
- `Definition`: Root container holding namespace, imports, structs, and services
- `StructDefine`: Struct with fields and optional inheritance (`extends`)
- `ServiceDefine`: HTTP service with GET/POST methods
- `PostMethod`/`GetMethod`: HTTP endpoints with params and return types

### Code Generation (`internal/generate/`)
- `go_generator.go`: Entry point coordinating generation
- `go_generator_structs.go`: Generates Go struct files
- `go_generator_server.go`: Generates Gin controller interfaces and implementations
- `go_generator_client.go`: Generates HTTP client code
- `swagger.go`: Generates OpenAPI 2.0 JSON documentation

## IDL Syntax

```idl
namespace go <package_name>
go_import "<go_package_path>" "<relative_idl_path>"
with_client=true|false

struct <Name> [extends [POINT] struct <ParentName>] {
    [required|optional] <type> <field_name>(annotations...),
}

service <Name> URL="<base_path>" {
    ("<description>")
    POST|GET URL="<path>" [not_login] <return_type> <method_name>(<params>),
}
```

Supported types: `bool`, `byte`, `i8`, `i16`, `i32`, `i64`, `float`, `double`, `string`, `list<T>`, `map<K,V>`, `struct <Name>`

Return type modifiers: `PAGEABLE` (for paginated lists), `void`

## Generated Code Structure

For input `example/order/order.idl` with `namespace go order`:
```
<output_dir>/order/
├── structs/
│   └── order_structs.go      # Struct definitions
├── order_controller.go       # Service interface + Gin bindings
└── order_controller_impl.go  # Implementation stubs (not overwritten if exists)
```

## Parser Regeneration

The parser is generated from an ANTLR4 grammar. If modifying the grammar:
1. Edit `Service.g4`
2. Regenerate with ANTLR4: `antlr4 -Dlanguage=Go -o internal/parser Service.g4`
