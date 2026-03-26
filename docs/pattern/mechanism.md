# Compiled vs Interpreted Languages

## Compiled Languages

Languages like C, C++, and Rust are **compiled**. Before you can run a program, a separate tool called a **compiler** translates the entire source code into a machine-code binary. The operating system then loads that binary into memory and executes it directly on the CPU.

```
                ┌──────────┐        ┌──────────┐        ┌─────────────┐
 Source Code ──►│ Compiler │──────► │  Binary  │──────► │ OS executes │
  (.c file)     └──────────┘        │ (machine │        │  directly   │
                                    │   code)  │        └─────────────┘
                                    └──────────┘
```

Key points:

- The compilation step happens **before** execution. You run `gcc main.c -o main`, then `./main`.
- The resulting binary is self-contained -- it no longer needs the compiler or the original source.
- Because machine code runs directly on the CPU, compiled programs tend to be very fast.
- The trade-off is a slower development cycle: edit, compile, run, repeat.

## Interpreted Languages

Languages like Python, Ruby, and JavaScript work differently. You never manually compile them into a binary. Instead, you just run the source file:

```bash
python main.py
```

A common misconception is that the source code is secretly compiled to a binary behind the scenes and then executed. That is not what happens. What actually runs is a **separate program** -- the interpreter -- which you installed beforehand (e.g. `python3`). The interpreter is itself a compiled binary, but its job is to **read your source file as text, analyze it, and execute it on the fly**.

```
                ┌─────────────────────────────────────────────┐
                │              Interpreter (binary)           │
                │                                             │
 Source Code ──►│  1. Read source text                        │──► Output
  (.py file)    │  2. Break it into tokens          (Lexer)   │
                │  3. Build a syntax tree           (Parser)  │
                │  4. Walk the tree and execute   (Evaluator) │
                └─────────────────────────────────────────────┘
```

Key points:

- There is **no separate compilation step** visible to the user. You just run the program.
- The interpreter does all the work at runtime: reading, parsing, and executing the code.
- This makes the development cycle faster (edit, run) at the cost of slower execution speed compared to compiled code.

## The Interpreter Pipeline

Whether you look at CPython, a JavaScript engine, or the interpreter built in this project, most interpreters share the same internal pipeline:

```
Source Code ──► Lexer ──► Tokens ──► Parser ──► AST ──► Evaluator ──► Result
```

Each stage transforms the program into a progressively more structured representation:

| Stage         | Input           | Output               | What it does                                                                                            |
| ------------- | --------------- | -------------------- | ------------------------------------------------------------------------------------------------------- |
| **Lexer**     | Raw source text | Stream of tokens     | Scans characters and groups them into meaningful units (numbers, keywords, operators, etc.)             |
| **Parser**    | Tokens          | Abstract Syntax Tree | Validates syntax and arranges tokens into a tree that reflects the structure and precedence of the code |
| **Evaluator** | AST             | Computed result      | Walks the tree node by node, executing operations and producing values                                  |

A fourth component, the **Environment**, sits alongside the evaluator. It is a data structure (essentially a chain of hash maps) that records variable and function bindings so the evaluator can look up names and store new values as the program runs.

## Why Build One From Scratch?

Building an interpreter teaches you what happens between typing `python main.py` and seeing output. You gain a deep understanding of:

- How source code is broken into tokens (lexical analysis).
- How grammar rules turn a flat list of tokens into a tree (parsing).
- How that tree is walked to produce a result (evaluation).
- How variables, functions, and scopes are tracked at runtime (environments).

In this project, we implement all four stages in Go.
