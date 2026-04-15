# REPL

## Role

The REPL (Read-Eval-Print Loop) is the **interactive entry point** for the interpreter. Its sole job is to read lines from stdin, run each one through the pipeline, and print the result — in a loop.

File execution is handled entirely by `main`, which reads the file with `os.ReadFile` and runs the pipeline once. The `repl` package is not involved.

---

## Two Modes of Execution

```mermaid
flowchart TD
    A[main] -->|no args| B[repl.Start\ninteractive loop]
    A -->|file path arg| C[os.ReadFile\nrun once and exit]

    B --> D[print prompt '>> ']
    D --> E[read line from stdin]
    E --> F[Lexer → Parser → Eval]
    F -->|parse errors| G[print errors, loop again]
    F -->|success| H[print result.Inspect]
    H --> D

    C -->|read error| I[stderr + exit 1]
    C --> J[Lexer → Parser]
    J -->|parse errors| K[stderr + exit 1]
    J --> L[Eval]
    L --> M[print result.Inspect]
```

---

## `repl.Start`

The `repl` package exposes a single function:

```go
func Start(in io.Reader, out io.Writer)
```

It accepts any `io.Reader` as input and any `io.Writer` as output, so it is easy to test or redirect. In production `main` passes `os.Stdin` and `os.Stdout`.

A **single environment** is created before the loop and shared across all inputs in the session — variables defined on one line are visible on the next.

```go
func Start(in io.Reader, out io.Writer) {
    scanner := bufio.NewScanner(in)
    env := object.NewEnvironment()  // shared across the whole session

    for {
        fmt.Fprint(out, prompt)     // print ">> "
        if !scanner.Scan() { break }

        line := scanner.Text()
        l := lexer.New(line)
        p := parser.New(l)
        program := p.ParseProgram()

        if len(p.Errors()) > 0 {
            for _, msg := range p.Errors() {
                fmt.Fprintf(out, "\t%s\n", msg)
            }
            continue   // skip eval, prompt again
        }

        result := evaluator.Eval(program, env)
        if result != nil {
            fmt.Fprintln(out, result.Inspect())
        }
    }
}
```

### Session lifecycle

```mermaid
sequenceDiagram
    participant User
    participant REPL as repl.Start
    participant Lexer
    participant Parser
    participant Eval
    participant Env as Environment

    Note over Env: created once at session start

    loop each input line
        REPL->>User: print ">> "
        User->>REPL: type source line
        REPL->>Lexer: New(line)
        REPL->>Parser: New(lexer)
        Parser->>REPL: ParseProgram()

        alt parse errors
            REPL->>User: print error messages
        else no errors
            REPL->>Eval: Eval(program, env)
            Eval->>Env: reads / writes bindings
            Eval-->>REPL: result object
            REPL->>User: print result.Inspect()
        end
    end

    Note over REPL: EOF or Ctrl-D ends the loop
```

Because `env` lives outside the loop, this session works as expected:

```
>> let x = 10
>> let y = 20
>> x + y
30
```

---

## File Mode (owned by `main`)

When a file path is provided, `main` handles everything directly — no `repl` package involved:

```mermaid
sequenceDiagram
    participant Main
    participant OS as os.ReadFile
    participant Lexer
    participant Parser
    participant Eval

    Main->>OS: os.ReadFile(args[0])
    OS-->>Main: []byte content

    Main->>Lexer: New(string(content))
    Main->>Parser: New(lexer)
    Parser-->>Main: program AST

    alt parse errors
        Main->>Main: print errors to stderr
        Main->>Main: os.Exit(1)
    else
        Main->>Eval: Eval(program, env)
        Eval-->>Main: result object
        Main->>Main: print result.Inspect()
    end
```

Errors in file mode exit the process immediately with a non-zero code, rather than being swallowed and looped past as in interactive mode.

---

## The Full Pipeline (both modes)

Both modes converge on the same three-stage pipeline:

```mermaid
flowchart LR
    SRC["source string"] --> LEX["Lexer\nlexer.New(src)"]
    LEX -->|token stream| PAR["Parser\nparser.New(lexer)"]
    PAR -->|AST| EV["Evaluator\nevaluator.Eval(program, env)"]
    EV -->|object.Object| OUT["result.Inspect()"]
```

| Stage     | Input          | Output          | Doc                                           |
|-----------|----------------|-----------------|-----------------------------------------------|
| Lexer     | source string  | token stream    | [lexer.md](../lexer/lexer.md)                 |
| Parser    | token stream   | AST             | [parser.md](../parser/parser.md)              |
| Evaluator | AST + env      | `object.Object` | [evaluator.md](../evaluator/evaluator.md)     |

---

## Key Takeaways

- The `repl` package is **only** responsible for the interactive loop. File I/O is `main`'s concern.
- `repl.Start` is a plain function — no struct, no state, just `(in, out)`.
- The shared `Environment` is what gives the interactive session its memory across inputs.
- File mode and interactive mode are clearly separated: one exits hard on errors, the other recovers and prompts again.
