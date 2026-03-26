# Abstract Syntax Tree (AST)

## Why Do We Need a Tree?

The lexer gives us a flat stream of tokens:

```
LET  IDENT("x")  ASSIGN  INT("2")  PLUS  INT("3")  ASTERISK  INT("4")  SEMICOLON
```

This tells us *what* the pieces are, but not *how they relate*. Is it `(2 + 3) * 4` or `2 + (3 * 4)`? A flat list cannot represent that distinction. We need a **tree** -- a hierarchical structure where the shape itself encodes precedence and nesting.

For `let x = 2 + 3 * 4;`, the correct AST looks like:

```
        LetStatement
        ├── Name: "x"
        └── Value: InfixExpression(+)
                   ├── Left:  IntegerLiteral(2)
                   └── Right: InfixExpression(*)
                              ├── Left:  IntegerLiteral(3)
                              └── Right: IntegerLiteral(4)
```

The `*` node is deeper in the tree than the `+` node, so it gets evaluated first. The tree structure **is** the precedence.

## Node Interfaces

Every node in the AST implements one of two marker interfaces, both of which extend a common `Node` interface:

```go
type Node interface {
    TokenLiteral() string // returns the literal value of the token (used for debugging)
    String() string       // prints the node back as source code (useful for tests)
}

type Statement interface {
    Node
    statementNode() // marker method -- distinguishes statements from expressions
}

type Expression interface {
    Node
    expressionNode() // marker method
}
```

**Statements** are instructions that do not produce a value (e.g. `let x = 5;`).
**Expressions** are pieces of code that *do* produce a value (e.g. `5`, `x + 1`, `fn(a) { a }`).

In practice, some "statements" contain expressions (a `let` statement contains an expression as its value), and some expressions contain statements (a function body is a block of statements). The two categories interleave.

## The Program Node

The root of every AST is a `Program` node. It holds a list of statements -- the top-level statements of the source file:

```go
type Program struct {
    Statements []Statement
}
```

Parsing an entire source file means filling this list.

## Statement Nodes

### LetStatement

```go
type LetStatement struct {
    Token token.Token // the LET token
    Name  *Identifier // the variable name being bound
    Value Expression  // the expression being assigned
}
```

Represents `let <name> = <value>;`. For example, `let x = 5;` produces a `LetStatement` where `Name` is an `Identifier("x")` and `Value` is an `IntegerLiteral(5)`.

### ReturnStatement

```go
type ReturnStatement struct {
    Token       token.Token // the RETURN token
    ReturnValue Expression  // the expression being returned
}
```

Represents `return <expression>;`.

### ExpressionStatement

```go
type ExpressionStatement struct {
    Token      token.Token // the first token of the expression
    Expression Expression
}
```

Most lines in our language are standalone expressions (e.g. `x + 1;` or `add(2, 3);`). An `ExpressionStatement` wraps an expression so it can sit in the `Program.Statements` list.

### BlockStatement

```go
type BlockStatement struct {
    Token      token.Token // the { token
    Statements []Statement
}
```

A `{ ... }` block containing zero or more statements. Used as the body of `if` branches, functions, and loops.

## Expression Nodes

### Identifier

```go
type Identifier struct {
    Token token.Token // the IDENT token
    Value string      // the name, e.g. "x", "add"
}
```

A reference to a variable or function name. Note: `Identifier` is an *expression* because using a name produces the value bound to it.

### IntegerLiteral

```go
type IntegerLiteral struct {
    Token token.Token
    Value int64
}
```

A numeric constant like `42`. The parser converts the string literal `"42"` from the token into the `int64` value `42`.

### StringLiteral

```go
type StringLiteral struct {
    Token token.Token
    Value string
}
```

A string constant like `"hello"`.

### Boolean

```go
type Boolean struct {
    Token token.Token
    Value bool
}
```

Either `true` or `false`.

### PrefixExpression

```go
type PrefixExpression struct {
    Token    token.Token // the prefix token, e.g. ! or -
    Operator string     // "!" or "-"
    Right    Expression // the operand
}
```

Represents `<operator><expression>`, such as `-5` or `!true`.

```
PrefixExpression(-)
└── Right: IntegerLiteral(5)
```

### InfixExpression

```go
type InfixExpression struct {
    Token    token.Token // the operator token
    Left     Expression
    Operator string     // "+", "-", "*", "/", "==", "!=", "<", ">"
    Right    Expression
}
```

Represents `<left> <operator> <right>`, such as `5 + 10` or `x == y`.

```
InfixExpression(+)
├── Left:  IntegerLiteral(5)
└── Right: IntegerLiteral(10)
```

### IfExpression

```go
type IfExpression struct {
    Token       token.Token     // the IF token
    Condition   Expression
    Consequence *BlockStatement // the "then" branch
    Alternative *BlockStatement // the "else" branch (may be nil)
}
```

Represents `if (<condition>) { <consequence> } else { <alternative> }`. The `else` branch is optional.

```
IfExpression
├── Condition:   InfixExpression(x > 5)
├── Consequence: BlockStatement { ... }
└── Alternative: BlockStatement { ... }   (or nil)
```

### FunctionLiteral

```go
type FunctionLiteral struct {
    Token      token.Token     // the FN token
    Parameters []*Identifier
    Body       *BlockStatement
}
```

Represents `fn(<params>) { <body> }`. Functions are values in our language -- they can be assigned to variables, passed as arguments, and returned from other functions.

```
FunctionLiteral
├── Parameters: [a, b]
└── Body: BlockStatement
          └── InfixExpression(a + b)
```

### CallExpression

```go
type CallExpression struct {
    Token     token.Token  // the ( token
    Function  Expression   // the function being called (Identifier or FunctionLiteral)
    Arguments []Expression
}
```

Represents `<function>(<args>)`, such as `add(2, 3)` or `fn(x){ x }(5)`.

```
CallExpression
├── Function:  Identifier("add")
└── Arguments: [IntegerLiteral(2), IntegerLiteral(3)]
```

### ForExpression

```go
type ForExpression struct {
    Token       token.Token
    Init        Statement       // e.g. let i = 0
    Condition   Expression      // e.g. i < 10
    Update      Statement       // e.g. i = i + 1
    Body        *BlockStatement
}
```

Represents `for (<init>; <condition>; <update>) { <body> }`.

## Full Tree Example

Source code:

```
let result = fn(a, b) { a + b }(2, 3);
```

AST:

```
Program
└── LetStatement
    ├── Name: Identifier("result")
    └── Value: CallExpression
              ├── Function: FunctionLiteral
              │             ├── Parameters: [Identifier("a"), Identifier("b")]
              │             └── Body: BlockStatement
              │                       └── ExpressionStatement
              │                           └── InfixExpression(+)
              │                               ├── Left:  Identifier("a")
              │                               └── Right: Identifier("b")
              └── Arguments: [IntegerLiteral(2), IntegerLiteral(3)]
```

This tree is what the evaluator receives. It will walk it from the top down, evaluating each node recursively to produce the final result: `5`.
