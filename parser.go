package cmdconfig

import (
	"fmt"
	"io"
	"strconv"
	"strings"
)

// really anything thats not a whitespace, quote or brace
func isBareword(b byte) bool {
	if b < 32 {
		return false
	}
	if isSpace(b) {
		return false
	}
	if isQuote1(b) {
		return false
	}
	if isQuote2(b) {
		return false
	}
	if isLeftBrace(b) {
		return false
	}
	if isBackQuote(b) {
		return false
	}
	return true
}
func isQuote1(b byte) bool    { return b == '\'' }
func isQuote2(b byte) bool    { return b == '"' }
func isSpace(b byte) bool     { return b == ' ' || b == '\t' }
func isLeftBrace(b byte) bool { return b == '{' }
func isNewLine(b byte) bool   { return b == '\n' }
func isBackQuote(b byte) bool { return b == '`' }

// Position represents a location in the input
type Position struct {
	Line   int // 1-based line number
	Column int // 1-based column number
	Offset int // 0-based byte offset
}

// String returns a human-readable position
func (p Position) String() string {
	return fmt.Sprintf("line %d, column %d", p.Line, p.Column)
}

// ScanError represents a parsing error with location information
type ScanError struct {
	Pos Position
	Msg string
}

func (e *ScanError) Error() string {
	return fmt.Sprintf("%s at %s", e.Msg, e.Pos)
}

type Scanner struct {
	s          []byte
	pos        int
	line       int // 1-based line number
	column     int // 1-based column number
	baseOffset int // base offset for nested scanners
}

func NewScanner(in []byte) *Scanner {
	return &Scanner{
		s:          in,
		pos:        0,
		line:       1,
		column:     1,
		baseOffset: 0,
	}
}

// NewFromScanner creates a new Scanner with position information inherited from parent
// The new scanner starts at line/column 1 but tracks its offset relative to the parent's position
func NewFromScanner(parent *Scanner, in []byte) *Scanner {
	parentPos := parent.currentPos()
	return &Scanner{
		s:          in,
		pos:        0,
		line:       parentPos.Line,   // Start from parent's current line
		column:     1,                // Reset column since we're parsing new content
		baseOffset: parentPos.Offset, // Track where this content starts in the original
	}
}

// currentPos returns the current position
func (s *Scanner) currentPos() Position {
	return Position{
		Line:   s.line,
		Column: s.column,
		Offset: s.baseOffset + s.pos,
	}
}

// advance moves the scanner position forward by one character
// and updates line/column tracking
func (s *Scanner) advance() {
	if s.pos < len(s.s) && s.s[s.pos] == '\n' {
		s.line++
		s.column = 1
	} else {
		s.column++
	}
	s.pos++
}

// errorAt creates a ScanError at the current position
func (s *Scanner) errorAt(msg string) error {
	return &ScanError{
		Pos: s.currentPos(),
		Msg: msg,
	}
}
func (s *Scanner) parseBareword() (string, error) {
	out := ""
	i := s.pos
	for s.pos < len(s.s) {
		b := s.s[s.pos]
		switch {
		case isSpace(b):
			out += string(s.s[i:s.pos])
			return out, nil
		case isQuote1(b):
			out += string(s.s[i:s.pos])
			inner, err := s.parseQuote1()
			if err != nil {
				return out, err
			}
			i = s.pos
			out += inner
		case isQuote2(b):
			out += string(s.s[i:s.pos])
			inner, err := s.parseQuote2()
			if err != nil {
				return out, err
			}
			i = s.pos
			out += inner
		case b == '\n':
			return out + string(s.s[i:s.pos]), nil
		case b == '\\':
			// Handle backslash escaping in barewords
			out += string(s.s[i:s.pos])
			escaped, err := s.parseBackslashEscape()
			if err != nil {
				return out, err
			}
			out += escaped
			i = s.pos
		default:
			s.advance()
		}
	}
	if out == "" {
		// bareword till EOF
		return string(s.s[i:]), nil
	}
	return out + string(s.s[i:]), nil
}
func (s *Scanner) parseBackQuote() (string, error) {
	s.advance()
	// first char after initial quote1
	i := s.pos
	for s.pos < len(s.s) {
		b := s.s[s.pos]
		if b == '`' {
			out := string(s.s[i:s.pos])
			s.advance()
			return out, nil
		}
		s.advance()
	}
	return "", s.errorAt("got EOF in back quote")
}
func (s *Scanner) parseQuote1() (string, error) {
	s.advance()
	// first char after initial quote1
	i := s.pos
	for s.pos < len(s.s) {
		b := s.s[s.pos]
		switch b {
		case '\'':
			out := string(s.s[i:s.pos])
			s.advance()
			return out, nil
		default:
			s.advance()
		}
	}
	return "", s.errorAt("got EOF in single quote")
}
func (s *Scanner) parseQuote2() (string, error) {
	s.advance()
	// first char after initial quote1
	i := s.pos
	out := ""
	for s.pos < len(s.s) {
		b := s.s[s.pos]
		switch b {
		case '"':
			out += string(s.s[i:s.pos])
			s.advance()
			return out, nil
		case '\\':
			// Handle backslash escaping in double quotes
			out += string(s.s[i:s.pos])
			escaped, err := s.parseBackslashEscape()
			if err != nil {
				return out, err
			}
			out += escaped
			i = s.pos
		default:
			s.advance()
		}
	}
	return "", s.errorAt("got EOF in double quote")
}

// parseBackslashEscape handles backslash escaping for barewords and double quotes
func (s *Scanner) parseBackslashEscape() (string, error) {
	if s.pos >= len(s.s) {
		return "", s.errorAt("got EOF after backslash")
	}

	// Skip the backslash
	s.advance()

	if s.pos >= len(s.s) {
		return "", s.errorAt("got EOF after backslash")
	}

	b := s.s[s.pos]
	s.advance()

	switch b {
	case 'n':
		return "\n", nil
	case 'r':
		return "\r", nil
	case 't':
		return "\t", nil
	case '\\':
		return "\\", nil
	case '"':
		return "\"", nil
	case '\'':
		return "'", nil
	case '\n':
		// Backslash-newline: line continuation (consume the newline, return nothing)
		return "", nil
	default:
		// For any other character, just escape it literally
		return string(b), nil
	}
}

// parseBraceEscape handles minimal escaping for brace content (only braces and backslashes)
func (s *Scanner) parseBraceEscape() (string, error) {
	if s.pos >= len(s.s) {
		return "", s.errorAt("got EOF after backslash")
	}

	// Skip the backslash
	s.advance()

	if s.pos >= len(s.s) {
		return "", s.errorAt("got EOF after backslash")
	}

	b := s.s[s.pos]
	s.advance()

	switch b {
	case '{':
		return "{", nil
	case '}':
		return "}", nil
	case '\\':
		return "\\", nil
	default:
		// For any other character, include the backslash literally
		// This preserves other escaping for the downstream parser
		return "\\" + string(b), nil
	}
}

func (s *Scanner) parseBrace() (string, error) {
	// skip opening brace
	s.advance()
	// first char after opening '{'
	i := s.pos
	out := ""
	stack := 1

	for s.pos < len(s.s) {
		b := s.s[s.pos]
		switch b {
		case '\\':
			// Handle minimal backslash escaping in braces
			out += string(s.s[i:s.pos])
			escaped, err := s.parseBraceEscape()
			if err != nil {
				return out, err
			}
			out += escaped
			i = s.pos
		case '{':
			stack += 1
			s.advance()
		case '}':
			stack -= 1
			if stack == 0 {
				out += string(s.s[i:s.pos])
				s.advance()
				// Apply dedent to remove common leading whitespace
				return dedent(out), nil
			}
			s.advance()
		default:
			s.advance()
		}
	}
	return "", s.errorAt("got EOF in opening brace")
}

// Next returns the arguments and the optional body, along with an error if any.
// ex: foo bar { the body }
//
//	--> []string{"foo", "bar"}, "the body"
//
// ex: foo bar
//
//	--> []stirng{"foo", "bar"}, ""
func (s *Scanner) Next() ([]string, string, error) {
	args := []string{}
	body := ""

	arg := ""
	var err error

	for s.pos < len(s.s) {
		b := s.s[s.pos]
		switch {
		case isSpace(b):
			s.advance()
		case isBareword(b) || isQuote1(b) || isQuote2(b) || b == '\\':
			arg, err = s.parseBareword()
			if err != nil {
				return args, body, err
			}
			args = append(args, arg)
		case isBackQuote(b):
			arg, err = s.parseBackQuote()
			if err != nil {
				return args, body, err
			}
			args = append(args, arg)
		case isLeftBrace(b):
			body, err = s.parseBrace()
			return args, body, err
		case isNewLine(b):
			s.advance()
			if len(args) > 0 {
				return args, body, nil
			}
		}
		// TODO: # comments
	}

	// nothing to do.. end of file
	if len(args) == 0 {
		return nil, "", io.EOF
	}
	return args, "", nil
}

// isBarewordString checks if a string can be represented as a bareword (no quotes needed)
func isBarewordString(s string) bool {
	if s == "" {
		return false // empty strings need quotes
	}

	for i := 0; i < len(s); i++ {
		b := s[i]
		// Use same logic as isBareword() but also exclude backslash
		if !isBareword(b) || b == '\\' {
			return false
		}
	}
	return true
}

// quoteArg quotes an argument, handling braces that strconv.Quote doesn't escape
func quoteArg(s string) string {
	// Check if it contains braces that need escaping
	needsBraceEscape := strings.ContainsAny(s, "{}")

	if needsBraceEscape {
		// Manual quoting with brace escaping
		result := strings.Builder{}
		result.WriteByte('"')

		for i := 0; i < len(s); i++ {
			b := s[i]
			switch b {
			case '"':
				result.WriteString("\\\"")
			case '\\':
				result.WriteString("\\\\")
			case '{':
				result.WriteString("\\{")
			case '}':
				result.WriteString("\\}")
			case '\n':
				result.WriteString("\\n")
			case '\r':
				result.WriteString("\\r")
			case '\t':
				result.WriteString("\\t")
			default:
				if b < 32 {
					result.WriteString(fmt.Sprintf("\\x%02x", b))
				} else {
					result.WriteByte(b)
				}
			}
		}

		result.WriteByte('"')
		return result.String()
	}

	// Use standard quoting for everything else
	return strconv.Quote(s)
}

// FormatIndent takes parsed arguments and body and returns a formatted command string
// with each line of the body indented by the given prefix string
func FormatIndent(args []string, body string, indent string) string {
	var parts []string

	// Format arguments
	for _, arg := range args {
		if isBarewordString(arg) {
			parts = append(parts, arg)
		} else {
			parts = append(parts, quoteArg(arg))
		}
	}

	result := strings.Join(parts, " ")

	// Add body if present
	if body != "" {
		result += " {\n"

		if indent != "" {
			// Split body into lines and indent each one
			lines := strings.Split(body, "\n")
			for _, line := range lines {
				result += indent + line + "\n"
			}
		} else {
			// No indentation
			result += body + "\n"
		}

		result += "}"
	}

	return result
}

// dedent removes common leading whitespace from all non-empty lines
// This implements a heuristic approach: only dedent if ALL non-empty lines
// share the same leading whitespace prefix
func dedent(s string) string {
	if s == "" {
		return s
	}

	lines := strings.Split(s, "\n")
	if len(lines) <= 1 {
		return s
	}

	// Find non-empty lines and their leading whitespace
	var nonEmptyLines []string
	var leadingWhitespace []string

	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			nonEmptyLines = append(nonEmptyLines, line)

			// Extract leading whitespace
			leadingWS := ""
			for _, char := range line {
				if char == ' ' || char == '\t' {
					leadingWS += string(char)
				} else {
					break
				}
			}
			leadingWhitespace = append(leadingWhitespace, leadingWS)
		}
	}

	// If no non-empty lines, return as-is
	if len(nonEmptyLines) == 0 {
		return s
	}

	// Find the shortest common prefix among all leading whitespace
	commonPrefix := leadingWhitespace[0]
	for _, ws := range leadingWhitespace[1:] {
		// Find common prefix between commonPrefix and ws
		minLen := len(commonPrefix)
		if len(ws) < minLen {
			minLen = len(ws)
		}

		newCommon := ""
		for i := 0; i < minLen; i++ {
			if commonPrefix[i] == ws[i] {
				newCommon += string(commonPrefix[i])
			} else {
				break
			}
		}
		commonPrefix = newCommon
	}

	// If no common prefix, return as-is
	if commonPrefix == "" {
		return s
	}

	// Remove common prefix from all lines
	result := make([]string, len(lines))
	prefixLen := len(commonPrefix)

	for i, line := range lines {
		if strings.TrimSpace(line) == "" {
			// Keep empty lines as-is
			result[i] = line
		} else if len(line) >= prefixLen && line[:prefixLen] == commonPrefix {
			// Remove common prefix
			result[i] = line[prefixLen:]
		} else {
			// This shouldn't happen if our logic is correct, but be safe
			result[i] = line
		}
	}

	return strings.Join(result, "\n")
}

// Format takes parsed arguments and body and returns a formatted command string
// This is equivalent to FormatIndent(args, body, "")
func Format(args []string, body string) string {
	return FormatIndent(args, body, "")
}
