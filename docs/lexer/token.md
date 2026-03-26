# Tokens

## What Is a Token?

A token is the **smallest meaningful unit** of source code. When the lexer reads raw text like `let x = 5;`, it does not work with individual characters. Instead, it groups characters into tokens that each carry a type and a literal value:

```
"let"  →  Token{ Type: LET,       Literal: "let" }
"x"    →  Token{ Type: IDENT,     Literal: "x"   }
"="    →  Token{ Type: ASSIGN,    Literal: "="   }
"5"    →  Token{ Type: INT,       Literal: "5"   }
";"    →  Token{ Type: SEMICOLON, Literal: ";"   }
```

Tokens are the **bridge** between raw text (which is hard to reason about) and the parser (which needs structured, categorized input).

## The Token Struct

In Go, a token is a simple struct with two fields:

```go
type TokenType string

type Token struct {
    Type    TokenType
    Literal string
}
```

- **`Type`** tells the parser *what kind* of token this is (a number? a keyword? an operator?).
- **`Literal`** preserves the *original text* from the source code. For the integer `42`, the literal is the string `"42"`. The parser or evaluator will convert it to an actual number later.

## Token Types

Every distinct category of text in the language gets its own `TokenType` constant. They fall into five groups:

### Identifiers and Literals

These represent user-defined names and raw values.

| Constant | Example | Description |
|---|---|---|
| `IDENT` | `x`, `myFunc`, `counter` | A user-defined name (variable or function) |
| `INT` | `0`, `42`, `1000` | An integer literal |
| `STRING` | `"hello"` | A string literal (the quotes are not included in the literal) |

### Operators

| Constant | Literal | Description |
|---|---|---|
| `ASSIGN` | `=` | Assignment |
| `PLUS` | `+` | Addition |
| `MINUS` | `-` | Subtraction or negation |
| `ASTERISK` | `*` | Multiplication |
| `SLASH` | `/` | Division |
| `BANG` | `!` | Logical NOT |
| `EQ` | `==` | Equality comparison |
| `NOT_EQ` | `!=` | Inequality comparison |
| `LT` | `<` | Less than |
| `GT` | `>` | Greater than |

Note that `==` and `!=` are **two-character operators**. The lexer must peek ahead to distinguish `=` (ASSIGN) from `==` (EQ), and `!` (BANG) from `!=` (NOT_EQ).

### Delimiters

| Constant | Literal | Description |
|---|---|---|
| `COMMA` | `,` | Separates function parameters / arguments |
| `SEMICOLON` | `;` | Ends a statement |
| `LPAREN` | `(` | Opens a grouped expression or parameter list |
| `RPAREN` | `)` | Closes it |
| `LBRACE` | `{` | Opens a block body |
| `RBRACE` | `}` | Closes it |

### Keywords

Keywords are identifiers that the language reserves for special meaning. They look like regular names during scanning, so the lexer checks every identifier against a **keyword lookup table** to decide whether it is a keyword or a user-defined name.

| Constant | Literal | Description |
|---|---|---|
| `LET` | `let` | Declares a variable binding |
| `FUNCTION` | `fn` | Declares a function literal |
| `IF` | `if` | Starts a conditional expression |
| `ELSE` | `else` | Alternate branch of a conditional |
| `RETURN` | `return` | Returns a value from a function |
| `TRUE` | `true` | Boolean literal |
| `FALSE` | `false` | Boolean literal |
| `FOR` | `for` | Starts a loop |

The lookup table in Go is simply a map:

```go
var keywords = map[string]TokenType{
    "fn":     FUNCTION,
    "let":    LET,
    "true":   TRUE,
    "false":  FALSE,
    "if":     IF,
    "else":   ELSE,
    "return": RETURN,
    "for":    FOR,
}

func LookupIdent(ident string) TokenType {
    if tok, ok := keywords[ident]; ok {
        return tok
    }
    return IDENT
}
```

When the lexer reads a word like `let`, it calls `LookupIdent("let")`, which returns `LET`. If the word were `myVar` instead, the lookup finds no match and returns `IDENT`.

### Special

| Constant | Description |
|---|---|
| `EOF` | Signals the end of input. The parser uses this to know when to stop. |
| `ILLEGAL` | Any character the lexer does not recognize (e.g. `@`, `#`). Triggers an error. |

## Worked Example

Given this source code:

```
let add = fn(a, b) { a + b };
```

The lexer produces the following token sequence:

| # | Type | Literal |
|---|---|---|
| 1 | `LET` | `let` |
| 2 | `IDENT` | `add` |
| 3 | `ASSIGN` | `=` |
| 4 | `FUNCTION` | `fn` |
| 5 | `LPAREN` | `(` |
| 6 | `IDENT` | `a` |
| 7 | `COMMA` | `,` |
| 8 | `IDENT` | `b` |
| 9 | `RPAREN` | `)` |
| 10 | `LBRACE` | `{` |
| 11 | `IDENT` | `a` |
| 12 | `PLUS` | `+` |
| 13 | `IDENT` | `b` |
| 14 | `RBRACE` | `}` |
| 15 | `SEMICOLON` | `;` |
| 16 | `EOF` | `` |

Notice that whitespace is **discarded** -- it is only used to separate tokens. The token stream is a flat list with no notion of nesting or precedence. That structure comes later, when the parser builds the AST.
