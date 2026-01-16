# cmdconfig

A Go library for parsing shell-style configuration files with maximum flexibility and human readability.

## Overview

**cmdconfig** is a command-structured data format that brings the familiarity of command-line syntax to configuration files. Think of it as "YAML for people who live in terminals" or "nginx.conf syntax, but general-purpose."

```bash
# Natural shell-like syntax
server web01 {
    host 192.168.1.10
    port 8080 9443
    ssl_cert "/path/to/cert.pem"
    upstream backend1 backend2 backend3

    location /api {
        proxy_pass http://backend
        timeout 30s
    }
}

user "John Doe" {
    email john@example.com
    roles admin developer
    config {
        theme dark
        notifications enabled
    }
}
```

## Features

### Core Capabilities
- **Shell-like parsing** with familiar escaping and quoting rules
- **Nested brace blocks** for hierarchical data structures
- **Flexible validation** - implement any validation logic in code
- **Custom processing** - case sensitivity, value constraints, transformations
- **Streaming friendly** - process configurations line-by-line
- **Precise error reporting** with line/column information
- **Round-trip encoding** - parse and format back to readable strings

### Advanced Features
- **Mixed quoting** - seamlessly combine single quotes, double quotes, and barewords
- **Backslash escaping** - full shell-style escape sequences
- **Automatic dedenting** - intelligent whitespace removal from nested blocks
- **Position tracking** - know exactly where parsing errors occur
- **Zero dependencies** - pure Go standard library

## Installation

```bash
go get github.com/yourusername/cmdconfig
```

## Quick Start

### Basic Parsing

```go
package main

import (
    "fmt"
    "io"
    "github.com/yourusername/cmdconfig"
)

func main() {
    config := `
    name "My App"
    port 8080
    database {
        host localhost
        user myapp
        password "secret123"
    }
    `

    scanner := cmdconfig.NewScanner([]byte(config))
    for {
        args, body, err := scanner.Next()
        if err == io.EOF {
            break
        }
        if err != nil {
            fmt.Printf("Parse error: %v\n", err)
            return
        }

        fmt.Printf("Command: %v, Body: %q\n", args, body)
    }
}
```

### Using with Go's TextUnmarshaler

```go
type Config struct {
    Name     string
    Port     int
    Database DatabaseConfig
}

func (c *Config) UnmarshalText(text []byte) error {
    scanner := cmdconfig.NewScanner(text)
    for {
        args, body, err := scanner.Next()
        if err == io.EOF {
            return nil
        }
        if err != nil {
            return fmt.Errorf("parse error: %w", err)
        }
        if len(args) == 0 {
            continue
        }

        pos := scanner.CurrentPos()
        switch args[0] {
        case "name":
            if len(args) < 2 {
                return fmt.Errorf("name requires argument at %s", pos)
            }
            c.Name = args[1]

        case "port":
            if len(args) < 2 {
                return fmt.Errorf("port requires argument at %s", pos)
            }
            port, err := strconv.Atoi(args[1])
            if err != nil {
                return fmt.Errorf("invalid port %q at %s", args[1], pos)
            }
            if port < 1 || port > 65535 {
                return fmt.Errorf("port %d out of range at %s", port, pos)
            }
            c.Port = port

        case "database":
            if err := c.Database.UnmarshalText([]byte(body)); err != nil {
                return fmt.Errorf("database config error at %s: %w", pos, err)
            }

        default:
            return fmt.Errorf("unknown directive %q at %s", args[0], pos)
        }
    }
}
```

## Why cmdconfig?

### ✅ Advantages

**🎯 Maximum Flexibility**
- Implement any validation logic you need
- Case-sensitive or insensitive parsing
- Custom value transformations and constraints
- Domain-specific processing rules

**👥 Human-Friendly**
- Familiar command-line syntax
- Natural for DevOps and system administrators
- Excellent for text-heavy configurations
- Minimal escaping requirements

**⚡ Performance**
- Streaming/incremental parsing
- No reflection overhead
- Direct control over memory allocation
- Line-by-line processing capability

**🔧 Powerful Features**
- Precise error reporting with line/column info
- Shell-style escaping and quoting
- Nested block structures
- Round-trip formatting support

### ⚠️ Trade-offs

**Manual Implementation Required**
- No automatic struct marshaling (like JSON/YAML)
- Must implement UnmarshalText for each type
- Validation logic written in code, not declarative

**Smaller Ecosystem**
- Not as widely supported as JSON/YAML/TOML
- Fewer third-party tools and integrations
- Custom format means custom tooling

## Format Comparison

| Feature | cmdconfig | JSON | YAML | TOML |
|---------|-----------|------|------|------|
| **Human editing** | ⭐⭐⭐⭐⭐ | ⭐⭐ | ⭐⭐⭐⭐ | ⭐⭐⭐⭐ |
| **Text content** | ⭐⭐⭐⭐⭐ | ⭐⭐ | ⭐⭐⭐ | ⭐⭐ |
| **Custom validation** | ⭐⭐⭐⭐⭐ | ⭐⭐ | ⭐⭐ | ⭐⭐ |
| **Processing flexibility** | ⭐⭐⭐⭐⭐ | ⭐⭐ | ⭐⭐ | ⭐⭐ |
| **Streaming/incremental** | ⭐⭐⭐⭐⭐ | ⭐ | ⭐ | ⭐ |
| **Auto serialization** | ⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ |
| **Deep nesting** | ⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐ |
| **Ecosystem support** | ⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐ | ⭐⭐⭐ |

## Use Cases

### Perfect For:
- **Infrastructure configuration** (nginx-style configs)
- **Multi-environment configs** requiring frequent human editing
- **Content with embedded metadata** (documentation systems)
- **Domain-specific languages** for technical users
- **Configurations with complex validation rules**
- **Template-heavy content** with configuration blocks

### Consider Alternatives For:
- **Machine-generated configurations** → Use JSON
- **Simple key-value configs** → Use TOML
- **Complex nested data structures** → Use YAML/JSON
- **Wide ecosystem compatibility** → Use JSON
- **Automatic serialization needs** → Use JSON/YAML/TOML

## Syntax Reference

### Basic Commands
```bash
# Simple key-value
name John
port 8080
enabled true

# Multiple arguments
servers web1 web2 web3
```

### Quoting and Escaping
```bash
# Barewords (no quotes needed)
path /usr/local/bin
host example.com

# Quoted strings
message "Hello, world!"
query 'SELECT * FROM users'
command `git status`

# Escaping
path "/path with spaces/file"
json_data "{\\"key\\": \\"value\\"}"
multiline "line1\\nline2"
```

### Nested Blocks
```bash
server web01 {
    host 192.168.1.10
    port 8080

    location /api {
        proxy_pass backend
        timeout 30s
    }
}
```

### Advanced Features
```bash
# Line continuation
long_command arg1 arg2 \
    arg3 arg4

# Mixed quoting in arguments
user name="John Doe" email='john@example.com' active

# Brace escaping in arguments (when needed)
regex "pattern with \\{ and \\}"
```

## API Reference

### Core Functions

```go
// Create a new scanner
scanner := NewScanner([]byte(input))

// Parse next command
args, body, err := scanner.Next()

// Get current position (for error reporting)
pos := scanner.CurrentPos()

// Format back to string
output := Format(args, body)
output := FormatIndent(args, body, "  ") // with indentation
```

### Error Handling

```go
// ScanError provides position information
type ScanError struct {
    Pos Position
    Msg string
}

// Position tracks location in input
type Position struct {
    Line   int // 1-based
    Column int // 1-based
    Offset int // 0-based byte offset
}
```

## Testing

```bash
go test                    # Run all tests
go test -v                 # Verbose output
go test -cover             # Show coverage (95.7%)
go test -coverprofile=coverage.out
go tool cover -html=coverage.out    # View coverage report
```

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes with tests
4. Ensure tests pass (`go test`)
5. Commit your changes (`git commit -am 'Add amazing feature'`)
6. Push to the branch (`git push origin feature/amazing-feature`)
7. Open a Pull Request

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Related Projects

- [nginx configuration format](https://nginx.org/en/docs/beginners_guide.html#conf_structure) - Similar syntax inspiration
- [HCL](https://github.com/hashicorp/hcl) - HashiCorp Configuration Language
- [TOML](https://toml.io/) - Tom's Obvious, Minimal Language
- [YAML](https://yaml.org/) - YAML Ain't Markup Language

---

**cmdconfig** - *Configuration files that feel as natural as your command line.*