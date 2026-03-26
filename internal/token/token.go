package token

type TokenType string

type Token struct {
	Type    TokenType
	Literal string
}

const (
	// Identifiers and literals
	IDENT  TokenType = "IDENT"
	INT    TokenType = "INT"
	STRING TokenType = "STRING"

	// Operators
	ASSIGN   TokenType = "ASSIGN"
	PLUS     TokenType = "PLUS"
	MINUS    TokenType = "MINUS"
	ASTERISK TokenType = "ASTERISK"
	SLASH    TokenType = "SLASH"
	BANG     TokenType = "BANG"
	EQ       TokenType = "EQ"
	NOT_EQ   TokenType = "NOT_EQ"
	LT       TokenType = "LT"
	GT       TokenType = "GT"

	// Delimiters
	COMMA     TokenType = "COMMA"
	SEMICOLON TokenType = "SEMICOLON"
	LPAREN    TokenType = "LPAREN"
	RPAREN    TokenType = "RPAREN"
	LBRACE    TokenType = "LBRACE"
	RBRACE    TokenType = "RBRACE"

	// Keywords
	LET      TokenType = "LET"
	FUNCTION TokenType = "FUNCTION"
	IF       TokenType = "IF"
	ELSE     TokenType = "ELSE"
	RETURN   TokenType = "RETURN"
	TRUE     TokenType = "TRUE"
	FALSE    TokenType = "FALSE"
	FOR      TokenType = "FOR"
)
