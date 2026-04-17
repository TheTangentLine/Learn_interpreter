# Environment (Storage)

## What Is the Environment?

The environment is a **symbol table** -- a data structure that maps names (strings) to values (runtime objects). Every time the program executes `let x = 5;`, the evaluator stores the binding `"x" → Integer(5)` in the environment. Every time it encounters the identifier `x`, it looks it up in the environment to retrieve `Integer(5)`.

Without an environment, the interpreter would have no memory. It could evaluate `2 + 3` but could never bind the result to a name or call it back later.

## The Data Structure

The environment is implemented as a **hash map** (Go's built-in `map[string]Object`) paired with an optional pointer to an **outer** environment:

```go
type Environment struct {
    store map[string]Object
    outer *Environment
}
```

- **`store`** -- A hash map that holds the bindings for the current scope. Keys are variable/function names; values are runtime objects (integers, strings, functions, etc.).
- **`outer`** -- A pointer to the enclosing (parent) scope. When a name is not found in `store`, the lookup continues in `outer`. If `outer` is `nil`, we are at the top-level global scope and the name is undefined.

### Why a Hash Map?

A hash map provides **O(1) average-time** lookups and insertions. Since variable access is one of the most frequent operations during evaluation, this efficiency matters. Go's `map[string]Object` handles hashing, collision resolution, and resizing internally, so the implementation is straightforward.

### Constructors

```go
func NewEnvironment() *Environment {
    return &Environment{store: make(map[string]Object), outer: nil}
}

func NewEnclosedEnvironment(outer *Environment) *Environment {
    env := NewEnvironment()
    env.outer = outer
    return env
}
```

`NewEnvironment` creates a fresh global scope with no parent. `NewEnclosedEnvironment` creates a child scope that chains to an existing environment.

## Get, Define, and Assign

### Define

`Define` creates a new binding in the **current** scope:

```go
func (e *Environment) Define(name string, value Object) Object {
    e.store[name] = value
    return value
}
```

`Define` never walks up the chain. This means `let x = 10;` inside a function body creates a new local binding even if `x` already exists in an outer scope. It is called by the evaluator when it processes a `let` statement.

### Assign (Reassignment)

`Assign` updates an **existing** binding, walking up the scope chain to find it:

```go
func (e *Environment) Assign(name string, value Object) (Object, error) {
    if _, isFound := e.store[name]; isFound {
        e.store[name] = value
        return value, nil
    }
    if e.outer != nil {
        return e.outer.Assign(name, value)
    }
    return nil, fmt.Errorf("Missing variable %s ", name)
}
```

Unlike `Define`, `Assign` searches upward through enclosing scopes until it finds the name. This means `x = 99` inside a function body can update a variable that was declared in an outer scope. If the name is not found anywhere in the chain, `Assign` returns an error — you cannot assign to an undeclared variable.

### Get (Scope Chain Lookup)

Retrieving a binding walks the scope chain:

```go
func (e *Environment) Get(name string) (Object, bool) {
    obj, ok := e.store[name]
    if !ok && e.outer != nil {
        obj, ok = e.outer.Get(name)
    }
    return obj, ok
}
```

The algorithm:

1. Look in the current scope's `store`.
2. If found, return the value.
3. If not found and there is an `outer` environment, recursively search `outer`.
4. If we reach the top-level scope and still don't find it, the name is undefined.

This chain lookup is what gives us **lexical scoping** -- inner scopes can see variables from outer scopes, but not vice versa.

## Scope Chaining in Practice

Consider this program:

```
let x = 10;
let y = 20;

let add = fn(a, b) {
    let result = a + b + x;
    return result;
};

add(1, 2);
```

Here is what the environment chain looks like when the evaluator is inside the function body of `add(1, 2)`:

```
┌─────────────────────────────┐
│ Function scope              │
│   a      → Integer(1)       │
│   b      → Integer(2)       │
│   result → Integer(13)      │
│                             │
│   outer ────────────────────┼──┐
└─────────────────────────────┘  │
                                 ▼
┌─────────────────────────────┐
│ Global scope                │
│   x   → Integer(10)         │
│   y   → Integer(20)         │
│   add → Function{...}       │
│                             │
│   outer → nil               │
└─────────────────────────────┘
```

When the evaluator encounters `x` inside the function body:

1. Search **function scope** → not found (only `a`, `b`, `result`).
2. Follow `outer` pointer to **global scope** → found: `Integer(10)`.

When it encounters `a`:

1. Search **function scope** → found: `Integer(1)`. Done.

`y` is accessible from inside the function (it exists in the global scope), but it simply is not referenced in this example.

## How Scopes Are Created

### Global Scope

When the interpreter starts, it creates a single global environment:

```go
env := object.NewEnvironment()
```

All top-level `let` statements and function definitions are stored here.

### Function Call Scope

Each time a function is called, the evaluator creates a **new enclosed environment** whose `outer` pointer points to the environment captured by the function (its closure), not the caller's environment:

```go
func extendFunctionEnv(fn *object.Function, args []object.Object) *object.Environment {
    env := object.NewEnclosedEnvironment(fn.Env)
    for i, param := range fn.Parameters {
        env.Define(param.Value, args[i])
    }
    return env
}
```

This is the mechanism behind **closures**. Consider:

```
let makeAdder = fn(x) {
    fn(y) { x + y };
};
let addFive = makeAdder(5);
addFive(3);
```

When `makeAdder(5)` is called:

1. A new scope is created with `x → Integer(5)`.
2. The inner `fn(y) { x + y }` is evaluated, creating a `Function` object that captures this scope as its `Env`.
3. `makeAdder` returns this `Function` object, which is bound to `addFive`.

When `addFive(3)` is called:

1. A new scope is created with `outer` pointing to the **captured scope** (where `x = 5`), not the global scope.
2. `y → Integer(3)` is bound in this new scope.
3. `x + y` is evaluated: `y` is found locally (3), `x` is found in the outer captured scope (5). Result: `Integer(8)`.

The scope chain at this point:

```
┌──────────────────────────┐
│ addFive call scope       │
│   y → Integer(3)         │
│   outer ─────────────────┼──┐
└──────────────────────────┘  │
                              ▼
┌──────────────────────────┐
│ makeAdder call scope     │
│   x → Integer(5)         │  (this scope persists because the inner
│   outer ─────────────────┼──┐  function holds a reference to it)
└──────────────────────────┘  │
                              ▼
┌──────────────────────────┐
│ Global scope             │
│   makeAdder → Function   │
│   addFive   → Function   │
│   outer → nil            │
└──────────────────────────┘
```

## Key Takeaways

- The environment is a **linked list of hash maps**. Each node holds one scope's bindings plus a pointer to the enclosing scope.
- **`Define`** always writes to the current scope (used by `let`). **`Assign`** walks up the chain to update an existing binding (used by reassignment). **`Get`** walks up the chain until it finds the name or runs out of scopes.
- **Closures** work because function objects capture their defining environment, and calls extend _that_ environment rather than the caller's.
- This design is simple, correct, and efficient enough for an educational interpreter. Production interpreters often use more complex structures (e.g. stack-allocated frames), but the core idea of chained scopes is the same.
