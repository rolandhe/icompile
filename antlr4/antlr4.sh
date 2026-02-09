#!/bin/sh

# Run from antlr4 directory: ./antlr4.sh
# Or specify project root: ./antlr4.sh /path/to/icomplie

SCRIPT_DIR=$(cd "$(dirname "$0")" && pwd)
PROJECT_ROOT=${1:-$(dirname "$SCRIPT_DIR")}
PARSER_DIR=$PROJECT_ROOT/internal/parser/

mkdir -p "$SCRIPT_DIR/go" "$PARSER_DIR"
antlr -Dlanguage=Go "$SCRIPT_DIR/Service.g4" -o "$SCRIPT_DIR/go/"
cp "$SCRIPT_DIR/go/"*.go "$PARSER_DIR"
rm -fr "$SCRIPT_DIR/go"
