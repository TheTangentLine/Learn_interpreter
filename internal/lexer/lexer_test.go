package lexer

import (
	"testing"

	"github.com/thetangentline/interpreter/internal/token"
)

func TestNextToken(t *testing.T) {
	input := `let x = 5 + 10;
let add = fn(a, b) { a + b };
if (x == 10) { return true; } else { return false; }
"hello"`

	expected := []token.Token{
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

		{Type: token.STRING, Literal: "hello"},

		{Type: token.EOF, Literal: ""},
	}

	l := New(input)

	for i, want := range expected {
		got := l.NextToken()
		if got.Type != want.Type {
			t.Errorf("test[%d] wrong token type: want=%q, got=%q (literal=%q)",
				i, want.Type, got.Type, got.Literal)
		}
		if got.Literal != want.Literal {
			t.Errorf("test[%d] wrong literal: want=%q, got=%q (type=%q)",
				i, want.Literal, got.Literal, got.Type)
		}
	}
}
