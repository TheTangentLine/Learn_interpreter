# Lexer

## Role

The lexer (also called a *scanner* or *tokenizer*) is the first stage of the interpreter pipeline. It takes **raw source code as a string** and produces a **stream of tokens**. Everything downstream -- the parser, the evaluator -- never sees raw text. They only work with tokens.

## The Lexer Struct

The lexer keeps a small amount of state as it moves through the input:

```go
type Lexer struct {
    input        string
    position     int  // points to the character we already read (current char)
    readPosition int  // points to the next character to read (always position + 1)
    ch           byte // the current character under examination
}
```

- **`input`** -- The entire source code as a single string.
- **`position`** -- Index of the character we are currently looking at.
- **`readPosition`** -- Index of the *next* character. This is how we "peek ahead" (e.g. to check if `=` is followed by another `=`).
- **`ch`** -- The actual byte at `input[position]`. When we have consumed the entire input, `ch` is set to `0` (null byte), which signals end-of-file.

### Initialization

When we create a new lexer, we immediately read the first character so that `ch`, `position`, and `readPosition` are all ready before anyone calls `NextToken()`:

```go
func New(input string) *Lexer {
    l := &Lexer{input: input}
    l.readChar() // sets position=0, readPosition=1, ch=input[0]
    return l
}
```

### readChar

This is the most fundamental helper. It advances the lexer by one character:

```go
func (l *Lexer) readChar() {
    if l.readPosition >= len(l.input) {
        l.ch = 0 // ASCII NUL signals EOF
    } else {
        l.ch = l.input[l.readPosition]
    }
    l.position = l.readPosition
    l.readPosition++
}
```

### peekChar

Sometimes we need to look at the next character *without* consuming it (e.g. to distinguish `=` from `==`):

```go
func (l *Lexer) peekChar() byte {
    if l.readPosition >= len(l.input) {
        return 0
    }
    return l.input[l.readPosition]
}
```

## The NextToken Algorithm

`NextToken()` is the lexer's public API. Every time it is called, it returns the next token from the input. The parser calls it repeatedly until it receives an `EOF` token.

Here is the algorithm in pseudocode:

```
function NextToken():
    1. Skip whitespace (spaces, tabs, newlines, carriage returns).
    2. Look at the current character (ch):
       a. If it is a single-char operator or delimiter (+, -, *, /, <, >, comma, etc.)
          → create the corresponding token, advance one character.
       b. If it is '=' or '!':
          → peek at the next character.
            - If next is '=', create EQ or NOT_EQ token, advance two characters.
            - Otherwise, create ASSIGN or BANG token, advance one character.
       c. If it is a letter (a-z, A-Z, _):
          → read the full identifier (all consecutive letters/digits/underscores).
          → look it up in the keyword table.
          → return a keyword token or an IDENT token.
       d. If it is a digit (0-9):
          → read the full number (all consecutive digits).
          → return an INT token.
       e. If it is a quote ("):
          → read characters until the closing quote.
          → return a STRING token.
       f. If it is the null byte (0):
          → return an EOF token.
       g. Otherwise:
          → return an ILLEGAL token.
    3. Advance the character pointer (readChar) if not already advanced.
    4. Return the token.
```

### Go Implementation Sketch

```go
func (l *Lexer) NextToken() token.Token {
    var tok token.Token
    l.skipWhitespace()

    switch l.ch {
    case '=':
        if l.peekChar() == '=' {
            l.readChar()
            tok = token.Token{Type: token.EQ, Literal: "=="}
        } else {
            tok = newToken(token.ASSIGN, l.ch)
        }
    case '!':
        if l.peekChar() == '=' {
            l.readChar()
            tok = token.Token{Type: token.NOT_EQ, Literal: "!="}
        } else {
            tok = newToken(token.BANG, l.ch)
        }
    case '+':
        tok = newToken(token.PLUS, l.ch)
    case '-':
        tok = newToken(token.MINUS, l.ch)
    case '*':
        tok = newToken(token.ASTERISK, l.ch)
    case '/':
        tok = newToken(token.SLASH, l.ch)
    case '<':
        tok = newToken(token.LT, l.ch)
    case '>':
        tok = newToken(token.GT, l.ch)
    case ',':
        tok = newToken(token.COMMA, l.ch)
    case ';':
        tok = newToken(token.SEMICOLON, l.ch)
    case '(':
        tok = newToken(token.LPAREN, l.ch)
    case ')':
        tok = newToken(token.RPAREN, l.ch)
    case '{':
        tok = newToken(token.LBRACE, l.ch)
    case '}':
        tok = newToken(token.RBRACE, l.ch)
    case '"':
        tok.Type = token.STRING
        tok.Literal = l.readString()
    case 0:
        tok = token.Token{Type: token.EOF, Literal: ""}
    default:
        if isLetter(l.ch) {
            tok.Literal = l.readIdentifier()
            tok.Type = token.LookupIdent(tok.Literal)
            return tok // readIdentifier already advanced past the identifier
        } else if isDigit(l.ch) {
            tok.Type = token.INT
            tok.Literal = l.readNumber()
            return tok // readNumber already advanced past the number
        } else {
            tok = newToken(token.ILLEGAL, l.ch)
        }
    }

    l.readChar()
    return tok
}
```

### Helper: readIdentifier / readNumber

Both work the same way -- they record the starting position, then keep calling `readChar()` as long as the character satisfies a predicate (letter or digit). When the predicate fails, they return the slice of `input` from start to current position:

```go
func (l *Lexer) readIdentifier() string {
    start := l.position
    for isLetter(l.ch) {
        l.readChar()
    }
    return l.input[start:l.position]
}

func (l *Lexer) readNumber() string {
    start := l.position
    for isDigit(l.ch) {
        l.readChar()
    }
    return l.input[start:l.position]
}
```

### Helper: skipWhitespace

Whitespace has no meaning in our language (it just separates tokens), so we skip it entirely:

```go
func (l *Lexer) skipWhitespace() {
    for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
        l.readChar()
    }
}
```

## Worked Example

Let's trace the lexer through `let x = 5 + 10;` character by character.

**Input:** `let x = 5 + 10;`

| Step | `ch` | Action | Token produced |
|---|---|---|---|
| 1 | `l` | `isLetter` → `readIdentifier` reads `"let"` → `LookupIdent` returns `LET` | `{LET, "let"}` |
| 2 | ` ` | `skipWhitespace` skips the space | -- |
| 3 | `x` | `isLetter` → `readIdentifier` reads `"x"` → `LookupIdent` returns `IDENT` | `{IDENT, "x"}` |
| 4 | ` ` | `skipWhitespace` | -- |
| 5 | `=` | peek is `' '` (not `=`) → single `=` → ASSIGN | `{ASSIGN, "="}` |
| 6 | ` ` | `skipWhitespace` | -- |
| 7 | `5` | `isDigit` → `readNumber` reads `"5"` | `{INT, "5"}` |
| 8 | ` ` | `skipWhitespace` | -- |
| 9 | `+` | matches `'+'` case | `{PLUS, "+"}` |
| 10 | ` ` | `skipWhitespace` | -- |
| 11 | `1` | `isDigit` → `readNumber` reads `"10"` | `{INT, "10"}` |
| 12 | `;` | matches `';'` case | `{SEMICOLON, ";"}` |
| 13 | `0` | null byte → EOF | `{EOF, ""}` |

**Final token stream:**
```
LET("let")  IDENT("x")  ASSIGN("=")  INT("5")  PLUS("+")  INT("10")  SEMICOLON(";")  EOF
```

## Key Takeaways

- The lexer is essentially a **state machine** that consumes one character at a time.
- It never backtracks -- it always moves forward through the input.
- Multi-character tokens (identifiers, numbers, two-character operators) require reading ahead, but the logic remains simple loops and peek operations.
- The lexer's output is a **flat stream**. It knows nothing about syntax, precedence, or nesting. That is the parser's job.
