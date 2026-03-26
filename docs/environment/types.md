# Object Types

## Why an Object System?

The evaluator needs to pass values around generically. When it evaluates `2 + 3`, the result is an integer. When it evaluates `"hello"`, the result is a string. When it evaluates `fn(x) { x }`, the result is a function. These are completely different things, but the evaluator must handle them uniformly -- `Eval` returns a single type regardless of what it evaluated.

Go solves this with **interfaces**. We define an `Object` interface that every runtime value implements. The evaluator always works with `Object`, and uses type assertions when it needs the concrete type.

## The Object Interface

```go
type ObjectType string

type Object interface {
    Type() ObjectType // returns a string constant identifying the kind of object
    Inspect() string  // returns a human-readable representation of the value
}
```

- **`Type()`** -- Returns a constant like `"INTEGER"`, `"BOOLEAN"`, `"STRING"`, etc. The evaluator uses this to decide which operations are valid (e.g. you can add two integers but not an integer and a function).
- **`Inspect()`** -- Returns a printable string. Used by the REPL to display results to the user and for debugging.

## All Object Types

### Integer

Wraps Go's `int64`. This is the only numeric type in our language.

```go
const INTEGER_OBJ = "INTEGER"

type Integer struct {
    Value int64
}

func (i *Integer) Type() ObjectType { return INTEGER_OBJ }
func (i *Integer) Inspect() string  { return fmt.Sprintf("%d", i.Value) }
```

**Created when:** The evaluator encounters an `IntegerLiteral` AST node, or produces a numeric result from arithmetic.

**Supported operations:** `+`, `-`, `*`, `/`, `<`, `>`, `==`, `!=`, unary `-`.

### Boolean

Wraps Go's `bool`. Instead of creating new `Boolean` objects every time, we use **two singletons**:

```go
const BOOLEAN_OBJ = "BOOLEAN"

type Boolean struct {
    Value bool
}

func (b *Boolean) Type() ObjectType { return BOOLEAN_OBJ }
func (b *Boolean) Inspect() string  { return fmt.Sprintf("%t", b.Value) }
```

```go
var (
    TRUE  = &Boolean{Value: true}
    FALSE = &Boolean{Value: false}
)

func nativeBoolToBooleanObject(input bool) *Boolean {
    if input {
        return TRUE
    }
    return FALSE
}
```

Since there are only two possible boolean values, singletons avoid unnecessary allocations and let us compare booleans with `==` at the pointer level instead of unpacking values.

**Created when:** Boolean literals (`true`, `false`), comparison operators (`<`, `>`, `==`, `!=`), or the `!` prefix operator.

**Supported operations:** `==`, `!=`, `!`.

### String

Wraps Go's `string`.

```go
const STRING_OBJ = "STRING"

type String struct {
    Value string
}

func (s *String) Type() ObjectType { return STRING_OBJ }
func (s *String) Inspect() string  { return s.Value }
```

**Created when:** The evaluator encounters a `StringLiteral` AST node, or concatenates two strings with `+`.

**Supported operations:** `+` (concatenation), `==`, `!=`.

### Null

Represents the absence of a value. Like booleans, it is a **singleton**:

```go
const NULL_OBJ = "NULL"

type Null struct{}

func (n *Null) Type() ObjectType { return NULL_OBJ }
func (n *Null) Inspect() string  { return "null" }
```

```go
var NULL = &Null{}
```

**Created when:** An `if` expression has no `else` branch and the condition is falsy, or when a function has no explicit return value.

`Null` is **falsy** -- it is treated as `false` in conditionals alongside `FALSE` itself.

### ReturnValue

A wrapper that signals "this value is being returned from a function." It is not a value the user ever sees directly -- it is an internal mechanism.

```go
const RETURN_VALUE_OBJ = "RETURN_VALUE"

type ReturnValue struct {
    Value Object // the actual value being returned
}

func (rv *ReturnValue) Type() ObjectType { return RETURN_VALUE_OBJ }
func (rv *ReturnValue) Inspect() string  { return rv.Value.Inspect() }
```

**Why is this needed?** Consider:

```
let f = fn() {
    let x = 10;
    return x;
    let y = 20;  // this should NOT execute
};
```

When the evaluator encounters `return x;`, it wraps the result in a `ReturnValue`. As the block statement evaluator processes each statement, it checks: "Is this a `ReturnValue`?" If yes, stop evaluating further statements and propagate it upward. The `applyFunction` call unwraps the `ReturnValue` to extract the actual value.

Without this wrapper, the evaluator would have no way to short-circuit a block of statements.

### Error

Wraps an error message string. Like `ReturnValue`, it propagates upward through the evaluation, stopping execution.

```go
const ERROR_OBJ = "ERROR"

type Error struct {
    Message string
}

func (e *Error) Type() ObjectType { return ERROR_OBJ }
func (e *Error) Inspect() string  { return "ERROR: " + e.Message }
```

```go
func newError(format string, a ...interface{}) *Error {
    return &Error{Message: fmt.Sprintf(format, a...)}
}
```

**Created when:** The evaluator encounters an invalid operation: type mismatch (`5 + true`), unknown operator, undefined identifier, calling a non-function, etc.

**Propagation:** Every `Eval` case checks `isError(result)` before continuing. If an error is detected, it is returned immediately without further evaluation. This gives us simple, predictable error handling without exceptions or panic.

### Function

Represents a first-class function value. This is the richest object type -- it stores everything needed to call the function later.

```go
const FUNCTION_OBJ = "FUNCTION"

type Function struct {
    Parameters []*ast.Identifier
    Body       *ast.BlockStatement
    Env        *Environment
}

func (f *Function) Type() ObjectType { return FUNCTION_OBJ }
func (f *Function) Inspect() string {
    params := []string{}
    for _, p := range f.Parameters {
        params = append(params, p.String())
    }
    return "fn(" + strings.Join(params, ", ") + ") { " + f.Body.String() + " }"
}
```

Three fields:

- **`Parameters`** -- The list of parameter names (`[]*ast.Identifier`). When the function is called, each parameter is bound to the corresponding argument value.
- **`Body`** -- The function's body (`*ast.BlockStatement`). This is an AST subtree that is evaluated when the function is called.
- **`Env`** -- The environment that was active when the function was **defined**. This is the closure. When the function is called, a new scope is created with `Env` as its outer scope, not the caller's scope.

**Created when:** The evaluator encounters a `FunctionLiteral` AST node.

**Why store AST nodes instead of compiled code?** Because this is a tree-walking interpreter. There is no compilation step. The evaluator re-walks the body's AST every time the function is called. This is simple and correct, though slower than compiling to bytecode.

## Summary Table

| Type | Go wrapper | Singleton? | Example value | Notes |
|---|---|---|---|---|
| Integer | `int64` | No | `42` | Only numeric type |
| Boolean | `bool` | Yes (`TRUE`/`FALSE`) | `true` | Pointer-comparable |
| String | `string` | No | `"hello"` | `+` for concatenation |
| Null | -- | Yes (`NULL`) | `null` | Falsy, represents absence |
| ReturnValue | `Object` | No | -- | Internal; unwrapped by `applyFunction` |
| Error | `string` | No | `"type mismatch"` | Propagates upward, stops execution |
| Function | params + body + env | No | `fn(a, b) { a + b }` | Carries closure environment |

## How It All Connects

When the evaluator processes the program:

```
let greet = fn(name) { "Hello, " + name };
greet("World");
```

The object flow is:

1. `fn(name) { ... }` → **Function** object (captures global env).
2. `env.Set("greet", Function{...})` → stored in the global environment.
3. `greet("World")` → evaluator calls `env.Get("greet")` → retrieves the **Function**.
4. Evaluate argument `"World"` → **String** object.
5. Create new scope: `{name: String("World"), outer: closureEnv}`.
6. Evaluate body: `"Hello, " + name`:
   - `"Hello, "` → **String** object.
   - `name` → look up in scope → **String("World")**.
   - `+` on two strings → concatenate → **String("Hello, World")**.
7. Return **String("Hello, World")** to the caller.

Every intermediate result is an `Object`. The evaluator never works with raw Go values -- it always wraps and unwraps through the object system. This uniformity is what makes the evaluator's type switch clean and extensible.
