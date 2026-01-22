package cmdconfig

import (
	"reflect"
	"testing"
)

func TestParse(t *testing.T) {
	type test struct {
		input string
		args  []string
		body  string
	}

	tests := []test{
		{
			input: "name John Brown",
			args:  []string{"name", "John", "Brown"},
		},
		{
			input: "name 'John' \"Brown\"",
			args:  []string{"name", "John", "Brown"},
		},
		{
			input: "name John Brown\n\n\n",
			args:  []string{"name", "John", "Brown"},
		},
		{
			input: "   name   John   Brown   \n\n\n",
			args:  []string{"name", "John", "Brown"},
		},
		{
			// check to make sure flags are ok
			input: "name -first John -last Brown",
			args:  []string{"name", "-first", "John", "-last", "Brown"},
		},
		{
			// check to make sure key=val are ok
			input: "name first=John last=Brown",
			args:  []string{"name", "first=John", "last=Brown"},
		},
		{
			// check arg quotes
			input: `func first 'second' "third"`,
			args:  []string{"func", "first", "second", "third"},
		},
		{
			// check to make sure flags are ok
			input: "name -first 'john'",
			args:  []string{"name", "-first", "john"},
		},
		{
			// check arg quotes
			input: `name first="mary ann" last="brown"`,
			args:  []string{"name", "first=mary ann", "last=brown"},
		},
		{
			// check arg quotes
			input: `name -first="mary ann" last="brown"`,
			args:  []string{"name", "-first=mary ann", "last=brown"},
		},
		{
			// check arg quotes
			input: `name -first "mary ann"`,
			args:  []string{"name", "-first", "mary ann"},
		},
		{
			// check arg quotes
			input: `name -first 'mary ann'   `,
			args:  []string{"name", "-first", "mary ann"},
		},
		{
			// check arg quotes
			input: `name first='mary ann' last='brown'`,
			args:  []string{"name", "first=mary ann", "last=brown"},
		},
		{
			// check arg quotes
			input: `name "first"='mary ann' "last"='brown'`,
			args:  []string{"name", "first=mary ann", "last=brown"},
		},
		{
			input: "name John Brown {\napple {banana}}",
			args:  []string{"name", "John", "Brown"},
			body:  "\napple {banana}",
		},
		{
			// multiline in double quotes
			input: "acmd \"a\nb\" c",
			args:  []string{"acmd", "a\nb", "c"},
		},
		{
			// multiline in single quotes
			input: "acmd 'a\nb' c",
			args:  []string{"acmd", "a\nb", "c"},
		},
		{
			// backslash escaping in barewords
			input: "echo a\\ b",
			args:  []string{"echo", "a b"},
		},
		{
			// backslash escaping in barewords - newline
			input: "echo a\\nb",
			args:  []string{"echo", "a\nb"},
		},
		{
			// backslash-newline continuation in barewords
			input: "echo a\\\nb",
			args:  []string{"echo", "ab"},
		},
		{
			// backslash escaping in double quotes
			input: "echo \"a\\nb\"",
			args:  []string{"echo", "a\nb"},
		},
		{
			// backslash-newline continuation in double quotes
			input: "echo \"a\\\nb\"",
			args:  []string{"echo", "ab"},
		},
		{
			// escaped quotes in double quotes
			input: "echo \"say \\\"hello\\\"\"",
			args:  []string{"echo", "say \"hello\""},
		},
		{
			// single quotes remain literal (no escaping)
			input: "echo 'a\\nb'",
			args:  []string{"echo", "a\\nb"},
		},
		{
			// backslash literal in single quotes
			input: "echo 'a\\\\b'",
			args:  []string{"echo", "a\\\\b"},
		},
		{
			// backtick strings (basic)
			input: "echo `hello`",
			args:  []string{"echo", "hello"},
		},
		{
			// backtick strings with spaces
			input: "echo `hello world`",
			args:  []string{"echo", "hello world"},
		},
		{
			// backtick strings with newlines
			input: "echo `hello\nworld`",
			args:  []string{"echo", "hello\nworld"},
		},
		{
			// escaped backslash in barewords
			input: "echo a\\\\b",
			args:  []string{"echo", "a\\b"},
		},
		{
			// escaped tab in barewords
			input: "echo a\\tb",
			args:  []string{"echo", "a\tb"},
		},
		{
			// escaped carriage return in barewords
			input: "echo a\\rb",
			args:  []string{"echo", "a\rb"},
		},
		{
			// escaped tab in double quotes
			input: "echo \"a\\tb\"",
			args:  []string{"echo", "a\tb"},
		},
		{
			// escaped carriage return in double quotes
			input: "echo \"a\\rb\"",
			args:  []string{"echo", "a\rb"},
		},
		{
			// escaped backslash in double quotes
			input: "echo \"a\\\\b\"",
			args:  []string{"echo", "a\\b"},
		},
		{
			// escaped single quote in barewords
			input: "echo a\\'b",
			args:  []string{"echo", "a'b"},
		},
		{
			// mixed quotes and escaping
			input: "cmd \"a\\\"b\" 'c\\d' e\\f",
			args:  []string{"cmd", "a\"b", "c\\d", "ef"},
		},
		{
			// brace with escaped braces
			input: "cmd { config \\{ key: value \\} }",
			args:  []string{"cmd"},
			body:  " config { key: value } ",
		},
		{
			// brace with escaped backslash
			input: "cmd { path: C:\\\\Program Files }",
			args:  []string{"cmd"},
			body:  " path: C:\\Program Files ",
		},
		{
			// brace with unescaped sequences (preserved for downstream parser)
			input: "cmd { name: \"John\\nDoe\" }",
			args:  []string{"cmd"},
			body:  " name: \"John\\nDoe\" ",
		},
		{
			// nested braces with escaping
			input: "cmd { outer { inner \\} still inner } }",
			args:  []string{"cmd"},
			body:  " outer { inner } still inner } ",
		},
	}

	for i, tc := range tests {
		s := NewScanner([]byte(tc.input))
		args, body, err := s.Next()
		if err != nil {
			t.Fatalf("case %d, got error %v", i, err)
		}
		if !reflect.DeepEqual(args, tc.args) {
			t.Fatalf("case %d, expected args %v got %v", i, tc.args, args)
		}
		if tc.body != body {
			t.Fatalf("case %d, expected body %q got %q", i, tc.body, body)
		}
	}
}

func TestParseErrors(t *testing.T) {
	type errorTest struct {
		input         string
		errorContains string
	}

	tests := []errorTest{
		{
			input:         "'unclosed single quote",
			errorContains: "got EOF in single quote",
		},
		{
			input:         "\"unclosed double quote",
			errorContains: "got EOF in double quote",
		},
		{
			input:         "`unclosed backtick",
			errorContains: "got EOF in back quote",
		},
		{
			input:         "{unclosed brace",
			errorContains: "got EOF in opening brace",
		},
		{
			input:         "trailing\\",
			errorContains: "got EOF after backslash",
		},
		{
			input:         "{unclosed with escape\\",
			errorContains: "got EOF after backslash",
		},
	}

	for i, tc := range tests {
		s := NewScanner([]byte(tc.input))
		_, _, err := s.Next()
		if err == nil {
			t.Fatalf("case %d, expected error containing %q but got nil", i, tc.errorContains)
		}
		if !contains(err.Error(), tc.errorContains) {
			t.Fatalf("case %d, expected error containing %q but got %q", i, tc.errorContains, err.Error())
		}
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || (len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsSubstr(s, substr))))
}

func containsSubstr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestEdgeCases(t *testing.T) {
	type test struct {
		input string
		args  []string
		body  string
	}

	tests := []test{
		{
			// empty input
			input: "",
			args:  nil,
		},
		{
			// only whitespace
			input: "   \t  \n  ",
			args:  nil,
		},
		{
			// only newlines
			input: "\n\n\n",
			args:  nil,
		},
		{
			// escaped unknown character
			input: "echo \\z",
			args:  []string{"echo", "z"},
		},
		{
			// multiple backslash-newlines
			input: "echo a\\\n\\\nb",
			args:  []string{"echo", "ab"},
		},
		{
			// empty braces
			input: "cmd {}",
			args:  []string{"cmd"},
			body:  "",
		},
		{
			// nested braces
			input: "cmd {outer {inner} more}",
			args:  []string{"cmd"},
			body:  "outer {inner} more",
		},
		{
			// bareword ending at EOF
			input: "single",
			args:  []string{"single"},
		},
	}

	for i, tc := range tests {
		s := NewScanner([]byte(tc.input))
		args, body, err := s.Next()
		if err != nil && err.Error() != "EOF" {
			t.Fatalf("case %d, got error %v", i, err)
		}
		if !equalStringSlices(args, tc.args) {
			t.Fatalf("case %d, expected args %v got %v", i, tc.args, args)
		}
		if tc.body != body {
			t.Fatalf("case %d, expected body %q got %q", i, tc.body, body)
		}
	}
}

func equalStringSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestFormat(t *testing.T) {
	type formatTest struct {
		args     []string
		body     string
		expected string
	}

	tests := []formatTest{
		{
			// Simple barewords, no body
			args:     []string{"deploy", "app"},
			body:     "",
			expected: "deploy app",
		},
		{
			// Barewords with body
			args:     []string{"cmd", "arg"},
			body:     "config content",
			expected: "cmd arg {\nconfig content\n}",
		},
		{
			// Arguments needing quotes
			args:     []string{"cmd", "with spaces", "and\"quotes"},
			body:     "",
			expected: "cmd \"with spaces\" \"and\\\"quotes\"",
		},
		{
			// Body with braces (preserved, not escaped)
			args:     []string{"template"},
			body:     "function() { return {key: value}; }",
			expected: "template {\nfunction() { return {key: value}; }\n}",
		},
		{
			// Body with backslashes (preserved as-is)
			args:     []string{"path"},
			body:     "C:\\Program Files\\App",
			expected: "path {\nC:\\Program Files\\App\n}",
		},
		{
			// Arguments with braces need escaping
			args:     []string{"cmd", "arg{with}braces"},
			body:     "",
			expected: "cmd \"arg\\{with\\}braces\"",
		},
		{
			// Empty args
			args:     []string{},
			body:     "just body",
			expected: " {\njust body\n}",
		},
		{
			// Mixed complex case with nested structure
			args:     []string{"deploy", "my-app", "env=prod"},
			body:     "{\n  config: {\n    port: 8080\n  }\n}",
			expected: "deploy my-app env=prod {\n{\n  config: {\n    port: 8080\n  }\n}\n}",
		},
	}

	for i, tc := range tests {
		result := Format(tc.args, tc.body)
		if result != tc.expected {
			t.Errorf("case %d, expected:\n%q\ngot:\n%q", i, tc.expected, result)
		}
	}
}

func TestRoundTrip(t *testing.T) {
	type roundTripTest struct {
		name string
		args []string
		body string
	}

	tests := []roundTripTest{
		{
			name: "simple command",
			args: []string{"echo", "hello"},
			body: "",
		},
		{
			name: "command with body",
			args: []string{"deploy", "app"},
			body: "config:\n  port: 8080",
		},
		{
			name: "complex arguments",
			args: []string{"cmd", "arg with spaces", "key=value"},
			body: "",
		},
		{
			name: "body with braces",
			args: []string{"template"},
			body: "{ nested: { data } }",
		},
		{
			name: "body with backslashes",
			args: []string{"path"},
			body: "C:\\Windows\\System32",
		},
		{
			name: "args with braces",
			args: []string{"complex", "arg{with}braces"},
			body: "",
		},
		{
			name: "mixed escaping",
			args: []string{"complex", "arg\"with'quotes"},
			body: "content with \\backslashes\\ and {braces}",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Format -> Parse (round trip)
			formatted := Format(tc.args, tc.body)

			s := NewScanner([]byte(formatted))
			parsedArgs, parsedBody, err := s.Next()

			if err != nil {
				t.Fatalf("round trip failed to parse: %v\nFormatted: %q", err, formatted)
			}

			// Check arguments match
			if !equalStringSlices(parsedArgs, tc.args) {
				t.Errorf("arguments don't match.\nOriginal: %v\nParsed: %v\nFormatted: %q",
					tc.args, parsedArgs, formatted)
			}

			// Check body matches (accounting for formatting newlines)
			expectedBody := tc.body
			if tc.body != "" {
				expectedBody = "\n" + tc.body + "\n"
			}
			if parsedBody != expectedBody {
				t.Errorf("body doesn't match.\nOriginal: %q\nExpected: %q\nParsed: %q\nFormatted: %q",
					tc.body, expectedBody, parsedBody, formatted)
			}
		})
	}
}

func TestFormatIndent(t *testing.T) {
	type indentTest struct {
		args     []string
		body     string
		indent   string
		expected string
	}

	tests := []indentTest{
		{
			// No body, no indentation needed
			args:     []string{"deploy", "app"},
			body:     "",
			indent:   "  ",
			expected: "deploy app",
		},
		{
			// Simple body with 2-space indentation
			args:     []string{"config", "set"},
			body:     "port: 8080\nhost: localhost",
			indent:   "  ",
			expected: "config set {\n  port: 8080\n  host: localhost\n}",
		},
		{
			// Nested structure with tab indentation
			args:     []string{"deploy"},
			body:     "app: myapp\nconfig:\n  port: 8080\n  env: prod",
			indent:   "\t",
			expected: "deploy {\n\tapp: myapp\n\tconfig:\n\t  port: 8080\n\t  env: prod\n}",
		},
		{
			// Empty indent (same as Format)
			args:     []string{"cmd"},
			body:     "line1\nline2",
			indent:   "",
			expected: "cmd {\nline1\nline2\n}",
		},
		{
			// Complex nesting with indentation
			args:     []string{"service", "definition"},
			body:     "name: web\nports:\n- 8080:80\nvolumes:\n- ./app:/app",
			indent:   "    ", // 4 spaces
			expected: "service definition {\n    name: web\n    ports:\n    - 8080:80\n    volumes:\n    - ./app:/app\n}",
		},
		{
			// Single line body
			args:     []string{"echo"},
			body:     "hello world",
			indent:   "  ",
			expected: "echo {\n  hello world\n}",
		},
		{
			// Body with empty lines
			args:     []string{"config"},
			body:     "line1\n\nline3",
			indent:   "  ",
			expected: "config {\n  line1\n  \n  line3\n}",
		},
	}

	for i, tc := range tests {
		result := FormatIndent(tc.args, tc.body, tc.indent)
		if result != tc.expected {
			t.Errorf("case %d, expected:\n%q\ngot:\n%q", i, tc.expected, result)
		}
	}
}

func TestFormatBackwardsCompatibility(t *testing.T) {
	// Test that Format() produces the same output as FormatIndent with empty indent
	testCases := []struct {
		args []string
		body string
	}{
		{[]string{"cmd"}, ""},
		{[]string{"deploy", "app"}, "config: value"},
		{[]string{"test", "with spaces"}, "multi\nline\nbody"},
	}

	for i, tc := range testCases {
		formatResult := Format(tc.args, tc.body)
		indentResult := FormatIndent(tc.args, tc.body, "")

		if formatResult != indentResult {
			t.Errorf("case %d, Format() and FormatIndent() differ:\nFormat(): %q\nFormatIndent(): %q",
				i, formatResult, indentResult)
		}
	}
}

func TestDedent(t *testing.T) {
	type dedentTest struct {
		name     string
		input    string
		expected string
	}

	tests := []dedentTest{
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "single line",
			input:    "hello",
			expected: "hello",
		},
		{
			name:     "single line with leading spaces",
			input:    "  hello",
			expected: "  hello",
		},
		{
			name:     "uniform 2-space indentation",
			input:    "  line1\n  line2\n  line3",
			expected: "line1\nline2\nline3",
		},
		{
			name:     "uniform 4-space indentation",
			input:    "    line1\n    line2",
			expected: "line1\nline2",
		},
		{
			name:     "uniform tab indentation",
			input:    "\tline1\n\tline2",
			expected: "line1\nline2",
		},
		{
			name:     "mixed indentation - common prefix removed",
			input:    "  line1\n    line2",
			expected: "line1\n  line2",
		},
		{
			name:     "partial common indentation",
			input:    "    line1\n  line2",
			expected: "  line1\nline2",
		},
		{
			name:     "with empty lines",
			input:    "  line1\n\n  line3",
			expected: "line1\n\nline3",
		},
		{
			name:     "only empty lines",
			input:    "\n\n\n",
			expected: "\n\n\n",
		},
		{
			name:     "mixed spaces and tabs - no common prefix",
			input:    "  line1\n\tline2",
			expected: "  line1\n\tline2",
		},
		{
			name:     "zero indentation line breaks pattern",
			input:    "  line1\nline2\n  line3",
			expected: "  line1\nline2\n  line3",
		},
		{
			name:     "complex nested structure",
			input:    "  config:\n    database:\n      host: localhost\n    cache:\n      enabled: true",
			expected: "config:\n  database:\n    host: localhost\n  cache:\n    enabled: true",
		},
		{
			name:     "trailing whitespace preserved",
			input:    "  line1  \n  line2\t",
			expected: "line1  \nline2\t",
		},
		{
			name:     "single space common prefix",
			input:    " a\n b\n c",
			expected: "a\nb\nc",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := dedent(tc.input)
			if result != tc.expected {
				t.Errorf("dedent failed.\nInput: %q\nExpected: %q\nGot: %q",
					tc.input, tc.expected, result)
			}
		})
	}
}

func TestNestedScannerPositions(t *testing.T) {
	// Test that NewFromScanner preserves position information correctly
	input := "cmd1 arg1\ncmd2 { nested content }\ncmd3 arg3"
	parentScanner := NewScanner([]byte(input))

	// Parse first command
	args1, _, err := parentScanner.Next()
	if err != nil {
		t.Fatalf("Failed to parse first command: %v", err)
	}
	if !reflect.DeepEqual(args1, []string{"cmd1", "arg1"}) {
		t.Errorf("Expected [cmd1 arg1], got %v", args1)
	}

	// Parse second command with body
	args2, body2, err := parentScanner.Next()
	if err != nil {
		t.Fatalf("Failed to parse second command: %v", err)
	}
	if !reflect.DeepEqual(args2, []string{"cmd2"}) {
		t.Errorf("Expected [cmd2], got %v", args2)
	}

	// Create nested scanner with position inheritance
	nestedScanner := NewFromScanner(parentScanner, []byte(body2))

	// The nested scanner should start with line 2 (where the brace block was)
	pos := nestedScanner.currentPos()
	if pos.Line != 2 {
		t.Errorf("Expected nested scanner to start at line 2, got line %d", pos.Line)
	}

	// Parse nested content
	nestedArgs, _, err := nestedScanner.Next()
	if err != nil {
		t.Fatalf("Failed to parse nested content: %v", err)
	}
	if !reflect.DeepEqual(nestedArgs, []string{"nested", "content"}) {
		t.Errorf("Expected [nested content], got %v", nestedArgs)
	}
}

func TestErrorLocations(t *testing.T) {
	type errorLocationTest struct {
		input          string
		expectedLine   int
		expectedColumn int
		expectedOffset int
		errorContains  string
	}

	tests := []errorLocationTest{
		{
			// Error on line 1, column 1
			input:          "'unclosed",
			expectedLine:   1,
			expectedColumn: 10, // At EOF position after 'unclosed'
			expectedOffset: 9,
			errorContains:  "got EOF in single quote at line 1, column 10",
		},
		{
			// Error on line 2 (single quote starts on line 2)
			input:          "\n'unclosed",
			expectedLine:   2,
			expectedColumn: 10, // After 'unclosed' (9 chars + initial quote position)
			expectedOffset: 10, // newline + 9 chars
			errorContains:  "got EOF in single quote at line 2, column 10",
		},
		{
			// Error spans multiple lines within single quote
			input:          "'start\nunclosed",
			expectedLine:   2,
			expectedColumn: 9,  // After 'unclosed'
			expectedOffset: 15, // 'start\nunclosed = 15 chars total
			errorContains:  "got EOF in single quote at line 2, column 9",
		},
		{
			// Backslash at EOF
			input:          "trailing\\",
			expectedLine:   1,
			expectedColumn: 10, // At EOF after backslash
			expectedOffset: 9,
			errorContains:  "got EOF after backslash at line 1, column 10",
		},
		{
			// Multi-line with brace error
			input:          "{\nline2\nline3",
			expectedLine:   3,
			expectedColumn: 6,  // At EOF position
			expectedOffset: 13, // { + \n + line2 + \n + line3 = 13 chars
			errorContains:  "got EOF in opening brace at line 3, column 6",
		},
	}

	for i, tc := range tests {
		s := NewScanner([]byte(tc.input))
		_, _, err := s.Next()

		if err == nil {
			t.Fatalf("case %d, expected error but got nil", i)
		}

		// Check if it's our custom error type
		scanErr, ok := err.(*ScanError)
		if !ok {
			t.Fatalf("case %d, expected *ScanError but got %T: %v", i, err, err)
		}

		// Check position information
		if scanErr.Pos.Line != tc.expectedLine {
			t.Errorf("case %d, expected line %d but got %d", i, tc.expectedLine, scanErr.Pos.Line)
		}
		if scanErr.Pos.Column != tc.expectedColumn {
			t.Errorf("case %d, expected column %d but got %d", i, tc.expectedColumn, scanErr.Pos.Column)
		}
		if scanErr.Pos.Offset != tc.expectedOffset {
			t.Errorf("case %d, expected offset %d but got %d", i, tc.expectedOffset, scanErr.Pos.Offset)
		}

		// Check error message format
		errStr := err.Error()
		if !contains(errStr, tc.errorContains) {
			t.Errorf("case %d, expected error containing %q but got %q", i, tc.errorContains, errStr)
		}
	}
}

func TestPositionTracking(t *testing.T) {
	// Test that position tracking works correctly during normal parsing
	input := "line1\nline2 arg\nline3"
	s := NewScanner([]byte(input))

	// Parse first command: "line1"
	args1, _, err1 := s.Next()
	if err1 != nil {
		t.Fatalf("unexpected error: %v", err1)
	}
	if !equalStringSlices(args1, []string{"line1"}) {
		t.Fatalf("expected [line1] but got %v", args1)
	}

	// Parse second command: "line2 arg"
	args2, _, err2 := s.Next()
	if err2 != nil {
		t.Fatalf("unexpected error: %v", err2)
	}
	if !equalStringSlices(args2, []string{"line2", "arg"}) {
		t.Fatalf("expected [line2 arg] but got %v", args2)
	}

	// Parse third command: "line3"
	args3, _, err3 := s.Next()
	if err3 != nil {
		t.Fatalf("unexpected error: %v", err3)
	}
	if !equalStringSlices(args3, []string{"line3"}) {
		t.Fatalf("expected [line3] but got %v", args3)
	}
}
