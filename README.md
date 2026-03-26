# Learn Interpreter

A simple interpreter built from scratch in Go for educational purposes. The interpreted language supports variable bindings, arithmetic expressions, functions (with closures), conditionals, loops, strings, and console/file I/O.

## Architecture

An interpreter reads source code as text and executes it directly, without producing a standalone binary. This interpreter follows the classic three-stage pipeline, backed by an environment that tracks runtime state:

```
Source Code
    │
    ▼
┌────────┐    Tokens    ┌────────┐     AST      ┌───────────┐
│  Lexer │ ──────────►  │ Parser │ ──────────►  │ Evaluator │ ──► Result
└────────┘              └────────┘              └───────────┘
                                                     │  ▲
                                                     ▼  │
                                                ┌─────────────┐
                                                │ Environment │
                                                └─────────────┘
```

1. **Lexer** -- Scans raw source text character by character and produces a stream of tokens.
2. **Parser** -- Consumes tokens, validates syntax, and builds an Abstract Syntax Tree (AST).
3. **Evaluator** -- Walks the AST recursively and executes each node.
4. **Environment** -- A chain of scopes that stores variable and function bindings at runtime.

## Language Features

| Feature              | Example                                      |
| -------------------- | -------------------------------------------- |
| Variable binding     | `let x = 10;`                                |
| Arithmetic           | `x + 2 * 3`                                  |
| Boolean logic        | `if (x > 5) { "big" } else { "small" }`      |
| Functions & closures | `let add = fn(a, b) { a + b };`              |
| Return statements    | `return x + y;`                              |
| String values        | `"hello" + " world"`                         |
| For loops            | `for (let i = 0; i < 10; i = i + 1) { ... }` |
| Console / file I/O   | `print("hello");`                            |

## Documentation

Detailed write-ups on how each component works:

| Topic                             | Document                                                   |
| --------------------------------- | ---------------------------------------------------------- |
| Compiled vs Interpreted Languages | [docs/pattern/mechanism.md](docs/pattern/mechanism.md)     |
| Tokens                            | [docs/lexer/token.md](docs/lexer/token.md)                 |
| Lexer                             | [docs/lexer/lexer.md](docs/lexer/lexer.md)                 |
| Abstract Syntax Tree              | [docs/parser/ast.md](docs/parser/ast.md)                   |
| Parser                            | [docs/parser/parser.md](docs/parser/parser.md)             |
| Evaluator                         | [docs/evaluator/evaluator.md](docs/evaluator/evaluator.md) |
| Environment / Storage             | [docs/environment/storage.md](docs/environment/storage.md) |
| Object Types                      | [docs/environment/types.md](docs/environment/types.md)     |

## Project Structure

```
my-lang/
├── cmd/
│   └── mylang/          # Entry point -- starts the REPL or runs a file
├── internal/
│   ├── token/           # Token type definitions and keyword lookup table
│   ├── lexer/           # Character scanner that produces tokens
│   ├── ast/             # AST node interfaces and concrete node types
│   ├── parser/          # Pratt parser that builds the AST from tokens
│   ├── object/          # Runtime value types (Integer, Boolean, String, …)
│   ├── environment/     # Scope chain storing variable/function bindings
│   ├── evaluator/       # Tree-walking evaluator that executes the AST
│   └── repl/            # Read-Eval-Print Loop for interactive use
├── examples/            # Sample programs written in the language
├── docs/                # The documentation you are reading now
├── go.mod
├── Makefile
└── README.md
```
