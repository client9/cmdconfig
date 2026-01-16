# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Go-based command configuration parser library inspired by shell-like command line syntax. The library provides robust parsing and formatting capabilities with comprehensive error handling.

**Key Files:**
- `parser.go`: Core parsing logic with Scanner struct and formatting functions
- `parser_test.go`: Comprehensive test suite with 95.7% coverage
- `go.mod`: Go module configuration (Go 1.25.1)

## Features

### Parsing Capabilities
- **Barewords**: Unquoted strings with shell-like escaping
- **Quoted strings**: Single quotes (`'`), double quotes (`"`), and backticks (`` ` ``)
- **Backslash escaping**: Full shell-like escape sequences (`\n`, `\t`, `\"`, `\\`, etc.)
- **Brace blocks**: Nested `{ }` blocks for command bodies with balanced parsing
- **Mixed quoting**: Seamless transitions between quote types within arguments
- **Position tracking**: Precise line/column error reporting

### Formatting Capabilities
- **Round-trip encoding**: `Format(args, body)` recreates parseable strings
- **Indented formatting**: `FormatIndent(args, body, indent)` for readable output
- **Smart quoting**: Automatic bareword detection with minimal escaping
- **Brace escaping**: Proper handling of `{` and `}` in arguments

### Advanced Features
- **Heuristic dedenting**: Automatic removal of common leading whitespace from brace content
- **Professional error reporting**: Structured errors with `Position` and `ScanError` types
- **Line continuation**: Backslash-newline support for multiline arguments
- **Minimal brace escaping**: Only `{`, `}`, and `\` are escaped within braces

## Architecture

### Core Components

**Scanner struct** (`parser.go:61-66`):
- Maintains parsing state with byte position and line/column tracking
- Provides character-by-character scanning with position updates

**Position struct** (`parser.go:40-44`):
- Tracks line (1-based), column (1-based), and byte offset (0-based)
- Used for precise error location reporting

**Key Methods**:
- `Next()` (`parser.go:325`): Main parsing entry point returning `([]string, string, error)`
- `parseBareword()` (`parser.go:105`): Handles unquoted arguments with escaping
- `parseBrace()` (`parser.go:280`): Processes nested brace blocks with dedenting
- `Format()` (`parser.go:552`): Encodes arguments and body back to string
- `FormatIndent()` (`parser.go:431`): Indented formatting for readable output
- `dedent()` (`parser.go:469`): Removes common leading whitespace

### Error Handling

**ScanError struct** (`parser.go:52-55`):
- Combines error message with precise position information
- Provides human-readable error formatting

## Development Commands

### Testing
```bash
go test                    # Run all tests
go test -v                 # Verbose test output
go test -cover             # Show coverage (95.7%)
go test -coverprofile=coverage.out  # Generate coverage profile
go tool cover -html=coverage.out    # View HTML coverage report
```

### Building
```bash
go build                   # Build the project
go mod tidy               # Clean up dependencies
```

## Usage Examples

### Basic Parsing
```go
scanner := NewScanner([]byte("cmd arg1 arg2 { body content }"))
args, body, err := scanner.Next()
// args: ["cmd", "arg1", "arg2"]
// body: "body content"
```

### Formatting
```go
// Basic formatting
result := Format([]string{"cmd", "arg with spaces"}, "body")
// Output: cmd "arg with spaces" { body }

// Indented formatting
result := FormatIndent([]string{"config"}, "  line1\n  line2", "  ")
// Output: config {
//           line1
//           line2
//         }
```

### Error Handling
```go
scanner := NewScanner([]byte("cmd 'unterminated"))
args, body, err := scanner.Next()
if err != nil {
    fmt.Printf("Parse error: %v\n", err)
    // Output: Parse error: got EOF in single quote at line 1, column 5
}
```

## Supported Syntax

### Shell-like Features
- **Escaping**: `cmd arg\ with\ spaces "quoted arg" 'single quoted'`
- **Line continuation**: `cmd long\\\n    argument`
- **Mixed quoting**: `name="John Doe" age='25'`
- **Backquotes**: `` `command substitution` ``

### Brace Blocks
- **Simple**: `cmd arg { body }`
- **Nested**: `outer { inner { content } more }`
- **Escaped braces**: `cmd { content with \\{ and \\} }`
- **Auto-dedenting**: Removes common indentation from body content

### Test Coverage
The test suite includes 15+ test categories covering:
- Basic parsing scenarios
- Error conditions with location tracking
- Round-trip encoding/decoding
- Edge cases and malformed input
- Format function variations
- Comprehensive dedent testing

Current coverage: **95.7% of statements**

## Implementation Notes

- Position tracking updates correctly for newlines and character advancement
- Dedent algorithm only removes whitespace common to ALL non-empty lines
- Brace escaping is minimal - preserves other escaping for downstream parsers
- Format functions follow Go's `json.Marshal`/`json.MarshalIndent` pattern
- Error messages provide actionable information with precise locations