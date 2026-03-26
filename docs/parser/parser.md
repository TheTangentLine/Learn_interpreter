# Parser

## Role

The parser is the second stage of the interpreter pipeline. It takes the **flat stream of tokens** produced by the lexer and builds an **Abstract Syntax Tree** (AST) that captures the structure, nesting, and operator precedence of the source code.

If the tokens do not form a valid program according to the language's grammar, the parser reports syntax errors.

## The Parser Struct

```go
type Parser struct {
    l         *lexer.Lexer
    curToken  token.Token // the current token being examined
    peekToken token.Token // the next token (one ahead)
    errors    []string

    prefixParseFns map[token.TokenType]prefixParseFn
    infixParseFns  map[token.TokenType]infixParseFn
}
```

- **`l`** -- A pointer to the lexer. The parser calls `l.NextToken()` to consume tokens.
- **`curToken`** / **`peekToken`** -- A two-token lookahead. At any point, `curToken` is the token we are currently deciding what to do with, and `peekToken` is the one that comes next. This lets us make decisions based on what follows.
- **`errors`** -- Accumulated parse error messages.
- **`prefixParseFns`** / **`infixParseFns`** -- Maps from token type to parsing functions (explained below).

### Advancing Tokens

```go
func (p *Parser) nextToken() {
    p.curToken = p.peekToken
    p.peekToken = p.l.NextToken()
}
```

Each call shifts the window forward by one token. What was `peekToken` becomes `curToken`, and a fresh token is read from the lexer into `peekToken`.

## Parsing Approach: Pratt Parsing

This parser uses **Pratt parsing** (also called *top-down operator precedence parsing*), invented by Vaughan Pratt in 1973. It is elegant, easy to extend, and handles operator precedence without needing a complex grammar specification.

The core ideas:

1. **Every token type can have a prefix parse function and/or an infix parse function** associated with it.
2. **Every infix operator has a precedence level** (a number). Higher numbers bind tighter.
3. The main `parseExpression` function uses these two pieces of information to build the correct tree shape.

### Precedence Levels

Precedence determines which operator "wins" when two operators compete for the same operand. We define them as increasing integer constants:

```go
const (
    _ int = iota
    LOWEST
    EQUALS      // ==, !=
    LESSGREATER  // <, >
    SUM         // +, -
    PRODUCT     // *, /
    PREFIX      // -x, !x
    CALL        // myFunc(x)
)
```

And a lookup table maps token types to their precedence:

```go
var precedences = map[token.TokenType]int{
    token.EQ:       EQUALS,
    token.NOT_EQ:   EQUALS,
    token.LT:       LESSGREATER,
    token.GT:       LESSGREATER,
    token.PLUS:     SUM,
    token.MINUS:    SUM,
    token.ASTERISK: PRODUCT,
    token.SLASH:    PRODUCT,
    token.LPAREN:   CALL,
}
```

`*` and `/` have precedence `PRODUCT` (5), while `+` and `-` have precedence `SUM` (4). Since 5 > 4, multiplication binds tighter than addition. This is how `2 + 3 * 4` becomes `2 + (3 * 4)` rather than `(2 + 3) * 4`.

### Parse Function Types

```go
type prefixParseFn func() ast.Expression
type infixParseFn  func(ast.Expression) ast.Expression
```

- A **prefix parse function** handles a token that appears at the *start* of an expression. It takes no arguments because there is nothing to its left. Examples: integer literals, identifiers, `!`, unary `-`, `if`, `fn`, grouped `(`.
- An **infix parse function** handles a token that appears *between* two expressions. It receives the already-parsed left-hand side as an argument. Examples: `+`, `-`, `*`, `/`, `==`, `!=`, `<`, `>`, and function call `(`.

During initialization, the parser registers these functions:

```go
// Prefix
p.registerPrefix(token.IDENT, p.parseIdentifier)
p.registerPrefix(token.INT, p.parseIntegerLiteral)
p.registerPrefix(token.STRING, p.parseStringLiteral)
p.registerPrefix(token.TRUE, p.parseBoolean)
p.registerPrefix(token.FALSE, p.parseBoolean)
p.registerPrefix(token.BANG, p.parsePrefixExpression)
p.registerPrefix(token.MINUS, p.parsePrefixExpression)
p.registerPrefix(token.LPAREN, p.parseGroupedExpression)
p.registerPrefix(token.IF, p.parseIfExpression)
p.registerPrefix(token.FUNCTION, p.parseFunctionLiteral)

// Infix
p.registerInfix(token.PLUS, p.parseInfixExpression)
p.registerInfix(token.MINUS, p.parseInfixExpression)
p.registerInfix(token.ASTERISK, p.parseInfixExpression)
p.registerInfix(token.SLASH, p.parseInfixExpression)
p.registerInfix(token.EQ, p.parseInfixExpression)
p.registerInfix(token.NOT_EQ, p.parseInfixExpression)
p.registerInfix(token.LT, p.parseInfixExpression)
p.registerInfix(token.GT, p.parseInfixExpression)
p.registerInfix(token.LPAREN, p.parseCallExpression)
```

Notice `LPAREN` appears in both maps. As a prefix, it starts a grouped expression `(expr)`. As an infix, it starts a function call `func(args)`.

## The Core Algorithm: parseExpression

This single function is the heart of Pratt parsing:

```go
func (p *Parser) parseExpression(precedence int) ast.Expression {
    // Step 1: find the prefix parse function for the current token
    prefix := p.prefixParseFns[p.curToken.Type]
    if prefix == nil {
        p.noPrefixParseFnError(p.curToken.Type)
        return nil
    }

    // Step 2: call it to parse the "left" side of the expression
    leftExp := prefix()

    // Step 3: as long as the next token is an infix operator with higher
    //         precedence than our current level, keep wrapping
    for precedence < p.peekPrecedence() {
        infix := p.infixParseFns[p.peekToken.Type]
        if infix == nil {
            return leftExp
        }
        p.nextToken()
        leftExp = infix(leftExp) // leftExp becomes the left child of the new node
    }

    return leftExp
}
```

Let's break this down:

1. **Find the prefix function.** Based on `curToken.Type`, look up the function that knows how to start an expression with this token. For an `INT` token, that's `parseIntegerLiteral`. For a `MINUS` token, that's `parsePrefixExpression`.

2. **Call it.** This returns the initial left-hand expression. For a simple integer like `3`, this is just an `IntegerLiteral` node.

3. **The infix loop.** Now we check: does the *next* token have a higher precedence than the `precedence` parameter we were called with? If yes, that operator should "steal" our expression as its left operand. We advance, call the infix function (passing the current `leftExp` as the left child), and the result becomes the new `leftExp`. We keep looping until we hit something with equal or lower precedence.

This loop is what makes precedence work. When we are inside a `*` parse (high precedence), the loop stops at `+` (lower precedence), so `*` gets to keep its operands. But when we are inside a `+` parse (lower precedence), the loop *continues* into `*`, letting multiplication group its operands first.

## Worked Example: Parsing `2 + 3 * 4`

Tokens: `INT(2)  PLUS  INT(3)  ASTERISK  INT(4)  EOF`

### Step-by-step trace

**Call:** `parseExpression(LOWEST)`

| # | Action | curToken | peekToken | Notes |
|---|---|---|---|---|
| 1 | Look up prefix for `INT` → `parseIntegerLiteral` | `INT(2)` | `PLUS` | |
| 2 | Call `parseIntegerLiteral()` → returns `IntegerLiteral(2)` | `INT(2)` | `PLUS` | `leftExp = IntegerLiteral(2)` |
| 3 | Check loop: `LOWEST(1) < peekPrecedence(PLUS=4)`? **Yes** | | `PLUS` | Enter loop |
| 4 | Look up infix for `PLUS` → `parseInfixExpression` | | | |
| 5 | `nextToken()` | `PLUS` | `INT(3)` | |
| 6 | Call `parseInfixExpression(IntegerLiteral(2))`: | | | |
| | - saves operator `+` and left=`IntegerLiteral(2)` | | | |
| | - calls `nextToken()` | `INT(3)` | `ASTERISK` | |
| | - calls `parseExpression(SUM)` **recursively** | | | |

**Recursive call:** `parseExpression(SUM)` (precedence = 4)

| # | Action | curToken | peekToken | Notes |
|---|---|---|---|---|
| 7 | Prefix for `INT` → `parseIntegerLiteral()` → `IntegerLiteral(3)` | `INT(3)` | `ASTERISK` | `leftExp = IntegerLiteral(3)` |
| 8 | Check loop: `SUM(4) < peekPrecedence(ASTERISK=5)`? **Yes** | | `ASTERISK` | Enter loop |
| 9 | Infix for `ASTERISK` → `parseInfixExpression` | | | |
| 10 | `nextToken()` | `ASTERISK` | `INT(4)` | |
| 11 | Call `parseInfixExpression(IntegerLiteral(3))`: | | | |
| | - saves operator `*`, left=`IntegerLiteral(3)` | | | |
| | - calls `nextToken()` | `INT(4)` | `EOF` | |
| | - calls `parseExpression(PRODUCT)` **recursively** | | | |

**Recursive call:** `parseExpression(PRODUCT)` (precedence = 5)

| # | Action | curToken | peekToken | Notes |
|---|---|---|---|---|
| 12 | Prefix for `INT` → `IntegerLiteral(4)` | `INT(4)` | `EOF` | `leftExp = IntegerLiteral(4)` |
| 13 | Check loop: `PRODUCT(5) < peekPrecedence(EOF=0)`? **No** | | | Exit loop |
| 14 | Return `IntegerLiteral(4)` | | | |

**Back in step 11:** `parseInfixExpression` now has right=`IntegerLiteral(4)`, builds:

```
InfixExpression(*)
├── Left:  IntegerLiteral(3)
└── Right: IntegerLiteral(4)
```

**Back in step 8 loop:** Check again: `SUM(4) < peekPrecedence(EOF=0)`? **No** → exit loop, return the `*` node.

**Back in step 6:** `parseInfixExpression` now has right=`InfixExpression(*)`, builds:

```
InfixExpression(+)
├── Left:  IntegerLiteral(2)
└── Right: InfixExpression(*)
           ├── Left:  IntegerLiteral(3)
           └── Right: IntegerLiteral(4)
```

**Back in step 3 loop:** Check again: `LOWEST(1) < peekPrecedence(EOF=0)`? **No** → exit loop, return the `+` node.

### Final AST

```
    (+)
   /   \
  2    (*)
      /   \
     3     4
```

Multiplication is deeper in the tree, so it executes first. The result is `2 + 12 = 14`.

## Parsing Statements

Not everything is an expression. Statements are parsed by checking the current token:

```go
func (p *Parser) parseStatement() ast.Statement {
    switch p.curToken.Type {
    case token.LET:
        return p.parseLetStatement()
    case token.RETURN:
        return p.parseReturnStatement()
    default:
        return p.parseExpressionStatement()
    }
}
```

### parseLetStatement

Expects the pattern `let <identifier> = <expression>;`:

```go
func (p *Parser) parseLetStatement() *ast.LetStatement {
    stmt := &ast.LetStatement{Token: p.curToken}

    p.expectPeek(token.IDENT)           // advance, expect identifier
    stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

    p.expectPeek(token.ASSIGN)          // advance, expect =

    p.nextToken()
    stmt.Value = p.parseExpression(LOWEST) // parse the right-hand side

    if p.peekToken.Type == token.SEMICOLON {
        p.nextToken()                   // consume optional semicolon
    }
    return stmt
}
```

### The Top-Level Loop

The parser's entry point repeatedly parses statements until EOF:

```go
func (p *Parser) ParseProgram() *ast.Program {
    program := &ast.Program{}
    for p.curToken.Type != token.EOF {
        stmt := p.parseStatement()
        if stmt != nil {
            program.Statements = append(program.Statements, stmt)
        }
        p.nextToken()
    }
    return program
}
```

## Key Takeaways

- **Pratt parsing** makes precedence easy: each infix operator has a numeric precedence, and the recursive `parseExpression` loop uses it to decide whether to keep grouping or to stop and return.
- **Prefix and infix parse functions** are registered per token type, making the parser easy to extend -- adding a new operator is just registering a new function with the right precedence.
- **Two-token lookahead** (`curToken` and `peekToken`) is enough to parse the entire language without backtracking.
- The parser produces an AST that the evaluator can walk. The parser itself never executes anything.
