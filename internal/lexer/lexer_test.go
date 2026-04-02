package lexer

import (
	"testing"

	"github.com/thetangentline/interpreter/internal/token"
)

// tokenize drains the lexer and returns all tokens including the final EOF.
func tokenize(input string) []token.Token {
	l := New(input)
	var tokens []token.Token
	for {
		tok := l.NextToken()
		tokens = append(tokens, tok)
		if tok.Type == token.EOF {
			break
		}
	}
	return tokens
}

// assertTokens compares got vs want slice, reporting every mismatch.
func assertTokens(t *testing.T, got, want []token.Token) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("token count: got %d, want %d\n got:  %v\n want: %v", len(got), len(want), got, want)
	}
	for i := range want {
		if got[i].Type != want[i].Type {
			t.Errorf("token[%d] type:    got %q, want %q (literal %q)", i, got[i].Type, want[i].Type, got[i].Literal)
		}
		if got[i].Literal != want[i].Literal {
			t.Errorf("token[%d] literal: got %q, want %q (type %q)", i, got[i].Literal, want[i].Literal, got[i].Type)
		}
	}
}

// ---------------------------------------------------------------------------
// Operators — every single- and two-character operator in isolation
// ---------------------------------------------------------------------------

func TestSingleCharOperators(t *testing.T) {
	tests := []struct {
		input   string
		tokType token.TokenType
		literal string
	}{
		{"=", token.ASSIGN, "="},
		{"+", token.PLUS, "+"},
		{"-", token.MINUS, "-"},
		{"*", token.ASTERISK, "*"},
		{"/", token.SLASH, "/"},
		{"!", token.BANG, "!"},
		{"<", token.LT, "<"},
		{">", token.GT, ">"},
		{",", token.COMMA, ","},
		{";", token.SEMICOLON, ";"},
		{"(", token.LPAREN, "("},
		{")", token.RPAREN, ")"},
		{"{", token.LBRACE, "{"},
		{"}", token.RBRACE, "}"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := tokenize(tt.input)
			want := []token.Token{
				{Type: tt.tokType, Literal: tt.literal},
				{Type: token.EOF, Literal: ""},
			}
			assertTokens(t, got, want)
		})
	}
}

func TestTwoCharOperators(t *testing.T) {
	tests := []struct {
		input   string
		tokType token.TokenType
		literal string
	}{
		{"==", token.EQ, "=="},
		{"!=", token.NOT_EQ, "!="},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := tokenize(tt.input)
			want := []token.Token{
				{Type: tt.tokType, Literal: tt.literal},
				{Type: token.EOF, Literal: ""},
			}
			assertTokens(t, got, want)
		})
	}
}

// Ensure = and == are not confused when adjacent.
func TestAssignVsEquals(t *testing.T) {
	got := tokenize("= ==")
	want := []token.Token{
		{Type: token.ASSIGN, Literal: "="},
		{Type: token.EQ, Literal: "=="},
		{Type: token.EOF, Literal: ""},
	}
	assertTokens(t, got, want)
}

// Ensure ! and != are not confused when adjacent.
func TestBangVsNotEq(t *testing.T) {
	got := tokenize("! !=")
	want := []token.Token{
		{Type: token.BANG, Literal: "!"},
		{Type: token.NOT_EQ, Literal: "!="},
		{Type: token.EOF, Literal: ""},
	}
	assertTokens(t, got, want)
}

// ---------------------------------------------------------------------------
// Keywords — every reserved word
// ---------------------------------------------------------------------------

func TestKeywords(t *testing.T) {
	tests := []struct {
		input   string
		tokType token.TokenType
	}{
		{"let", token.LET},
		{"fn", token.FUNCTION},
		{"if", token.IF},
		{"else", token.ELSE},
		{"return", token.RETURN},
		{"true", token.TRUE},
		{"false", token.FALSE},
		{"for", token.FOR},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := tokenize(tt.input)
			want := []token.Token{
				{Type: tt.tokType, Literal: tt.input},
				{Type: token.EOF, Literal: ""},
			}
			assertTokens(t, got, want)
		})
	}
}

// A keyword that is a prefix of an identifier must not be mis-classified.
func TestKeywordPrefixNotMisclassified(t *testing.T) {
	tests := []struct{ input, literal string }{
		{"letter", "letter"},   // starts with "let"
		{"ifTrue", "ifTrue"},   // starts with "if"
		{"forLoop", "forLoop"}, // starts with "for"
		{"returned", "returned"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := tokenize(tt.input)
			want := []token.Token{
				{Type: token.IDENT, Literal: tt.literal},
				{Type: token.EOF, Literal: ""},
			}
			assertTokens(t, got, want)
		})
	}
}

// ---------------------------------------------------------------------------
// Identifiers
// ---------------------------------------------------------------------------

func TestIdentifiers(t *testing.T) {
	tests := []string{
		"foo",
		"bar",
		"fooBar",
		"my_var",
		"_private",
		"x",
		"FOO",
	}

	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			got := tokenize(input)
			want := []token.Token{
				{Type: token.IDENT, Literal: input},
				{Type: token.EOF, Literal: ""},
			}
			assertTokens(t, got, want)
		})
	}
}

// ---------------------------------------------------------------------------
// Integer literals
// ---------------------------------------------------------------------------

func TestIntegers(t *testing.T) {
	tests := []string{"0", "5", "42", "100", "99999"}

	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			got := tokenize(input)
			want := []token.Token{
				{Type: token.INT, Literal: input},
				{Type: token.EOF, Literal: ""},
			}
			assertTokens(t, got, want)
		})
	}
}

// ---------------------------------------------------------------------------
// String literals
// ---------------------------------------------------------------------------

func TestStringLiterals(t *testing.T) {
	tests := []struct {
		input   string
		literal string
	}{
		{`"hello"`, "hello"},
		{`"hello world"`, "hello world"},
		{`""`, ""},
		{`"123"`, "123"},
		{`"!@#"`, "!@#"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := tokenize(tt.input)
			want := []token.Token{
				{Type: token.STRING, Literal: tt.literal},
				{Type: token.EOF, Literal: ""},
			}
			assertTokens(t, got, want)
		})
	}
}

// ---------------------------------------------------------------------------
// Whitespace handling
// ---------------------------------------------------------------------------

func TestWhitespaceSkipping(t *testing.T) {
	// All whitespace variants between tokens should be silently skipped.
	got := tokenize("  \t  5  \n  +  \r\n  3  ")
	want := []token.Token{
		{Type: token.INT, Literal: "5"},
		{Type: token.PLUS, Literal: "+"},
		{Type: token.INT, Literal: "3"},
		{Type: token.EOF, Literal: ""},
	}
	assertTokens(t, got, want)
}

func TestOnlyWhitespace(t *testing.T) {
	got := tokenize("   \t\n\r\n   ")
	want := []token.Token{
		{Type: token.EOF, Literal: ""},
	}
	assertTokens(t, got, want)
}

// ---------------------------------------------------------------------------
// ILLEGAL token
// ---------------------------------------------------------------------------

func TestIllegalCharacters(t *testing.T) {
	illegals := []string{"@", "#", "$", "^", "&", "~", "`"}

	for _, ch := range illegals {
		t.Run(ch, func(t *testing.T) {
			got := tokenize(ch)
			if len(got) == 0 || got[0].Type != token.ILLEGAL {
				t.Errorf("expected ILLEGAL, got %v", got)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Empty input
// ---------------------------------------------------------------------------

func TestEmptyInput(t *testing.T) {
	got := tokenize("")
	want := []token.Token{{Type: token.EOF, Literal: ""}}
	assertTokens(t, got, want)
}

// ---------------------------------------------------------------------------
// EOF stability — calling NextToken past EOF must keep returning EOF
// ---------------------------------------------------------------------------

func TestEOFStability(t *testing.T) {
	l := New("")
	for i := 0; i < 5; i++ {
		tok := l.NextToken()
		if tok.Type != token.EOF {
			t.Fatalf("call %d: expected EOF, got %q (%q)", i, tok.Type, tok.Literal)
		}
	}
}

// ---------------------------------------------------------------------------
// Integration — a realistic program fragment covering all token types
// ---------------------------------------------------------------------------

func TestIntegration(t *testing.T) {
	input := `let x = 5 + 10;
let add = fn(a, b) { a + b };
if (x == 10) { return true; } else { return false; }
for (let i = 0; i < 10; i) { i }
!x != x
"hello"
@`

	want := []token.Token{
		{Type: token.LET, Literal: "let"},
		{Type: token.IDENT, Literal: "x"},
		{Type: token.ASSIGN, Literal: "="},
		{Type: token.INT, Literal: "5"},
		{Type: token.PLUS, Literal: "+"},
		{Type: token.INT, Literal: "10"},
		{Type: token.SEMICOLON, Literal: ";"},

		{Type: token.LET, Literal: "let"},
		{Type: token.IDENT, Literal: "add"},
		{Type: token.ASSIGN, Literal: "="},
		{Type: token.FUNCTION, Literal: "fn"},
		{Type: token.LPAREN, Literal: "("},
		{Type: token.IDENT, Literal: "a"},
		{Type: token.COMMA, Literal: ","},
		{Type: token.IDENT, Literal: "b"},
		{Type: token.RPAREN, Literal: ")"},
		{Type: token.LBRACE, Literal: "{"},
		{Type: token.IDENT, Literal: "a"},
		{Type: token.PLUS, Literal: "+"},
		{Type: token.IDENT, Literal: "b"},
		{Type: token.RBRACE, Literal: "}"},
		{Type: token.SEMICOLON, Literal: ";"},

		{Type: token.IF, Literal: "if"},
		{Type: token.LPAREN, Literal: "("},
		{Type: token.IDENT, Literal: "x"},
		{Type: token.EQ, Literal: "=="},
		{Type: token.INT, Literal: "10"},
		{Type: token.RPAREN, Literal: ")"},
		{Type: token.LBRACE, Literal: "{"},
		{Type: token.RETURN, Literal: "return"},
		{Type: token.TRUE, Literal: "true"},
		{Type: token.SEMICOLON, Literal: ";"},
		{Type: token.RBRACE, Literal: "}"},
		{Type: token.ELSE, Literal: "else"},
		{Type: token.LBRACE, Literal: "{"},
		{Type: token.RETURN, Literal: "return"},
		{Type: token.FALSE, Literal: "false"},
		{Type: token.SEMICOLON, Literal: ";"},
		{Type: token.RBRACE, Literal: "}"},

		{Type: token.FOR, Literal: "for"},
		{Type: token.LPAREN, Literal: "("},
		{Type: token.LET, Literal: "let"},
		{Type: token.IDENT, Literal: "i"},
		{Type: token.ASSIGN, Literal: "="},
		{Type: token.INT, Literal: "0"},
		{Type: token.SEMICOLON, Literal: ";"},
		{Type: token.IDENT, Literal: "i"},
		{Type: token.LT, Literal: "<"},
		{Type: token.INT, Literal: "10"},
		{Type: token.SEMICOLON, Literal: ";"},
		{Type: token.IDENT, Literal: "i"},
		{Type: token.RPAREN, Literal: ")"},
		{Type: token.LBRACE, Literal: "{"},
		{Type: token.IDENT, Literal: "i"},
		{Type: token.RBRACE, Literal: "}"},

		{Type: token.BANG, Literal: "!"},
		{Type: token.IDENT, Literal: "x"},
		{Type: token.NOT_EQ, Literal: "!="},
		{Type: token.IDENT, Literal: "x"},

		{Type: token.STRING, Literal: "hello"},

		{Type: token.ILLEGAL, Literal: "@"},

		{Type: token.EOF, Literal: ""},
	}

	assertTokens(t, tokenize(input), want)
}
