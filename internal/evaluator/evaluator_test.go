package evaluator

import (
	"testing"

	"github.com/thetangentline/interpreter/internal/lexer"
	"github.com/thetangentline/interpreter/internal/object"
	"github.com/thetangentline/interpreter/internal/parser"
)

// evalInput is the central test helper: lex → parse → eval.
func evalInput(input string) object.Object {
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	env := object.NewEnvironment()
	return Eval(program, env)
}

func assertInteger(t *testing.T, obj object.Object, want int64) {
	t.Helper()
	result, ok := obj.(*object.Integer)
	if !ok {
		t.Fatalf("object is %T (%+v), want *object.Integer", obj, obj)
	}
	if result.Value != want {
		t.Errorf("Value = %d, want %d", result.Value, want)
	}
}

func assertBoolean(t *testing.T, obj object.Object, want bool) {
	t.Helper()
	result, ok := obj.(*object.Boolean)
	if !ok {
		t.Fatalf("object is %T (%+v), want *object.Boolean", obj, obj)
	}
	if result.Value != want {
		t.Errorf("Value = %v, want %v", result.Value, want)
	}
}

func assertString(t *testing.T, obj object.Object, want string) {
	t.Helper()
	result, ok := obj.(*object.String)
	if !ok {
		t.Fatalf("object is %T (%+v), want *object.String", obj, obj)
	}
	if result.Value != want {
		t.Errorf("Value = %q, want %q", result.Value, want)
	}
}

func assertError(t *testing.T, obj object.Object, wantMsg string) {
	t.Helper()
	errObj, ok := obj.(*object.Error)
	if !ok {
		t.Fatalf("object is %T (%+v), want *object.Error", obj, obj)
	}
	if errObj.Message != wantMsg {
		t.Errorf("Error.Message = %q, want %q", errObj.Message, wantMsg)
	}
}

func assertNull(t *testing.T, obj object.Object) {
	t.Helper()
	if obj != NULL {
		t.Errorf("object = %T (%+v), want NULL", obj, obj)
	}
}

// ---------------------------------------------------------------------------
// Integer literals and arithmetic
// ---------------------------------------------------------------------------

func TestEvalIntegerExpression(t *testing.T) {
	tests := []struct {
		input string
		want  int64
	}{
		{"5", 5},
		{"42", 42},
		{"-5", -5},
		{"-42", -42},
		{"5 + 3", 8},
		{"10 - 4", 6},
		{"3 * 4", 12},
		{"10 / 2", 5},
		{"2 + 3 * 4", 14},
		{"(2 + 3) * 4", 20},
		{"50 / 2 * 2 + 10 - 5", 55},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			assertInteger(t, evalInput(tt.input), tt.want)
		})
	}
}

// ---------------------------------------------------------------------------
// Boolean literals and comparisons
// ---------------------------------------------------------------------------

func TestEvalBooleanExpression(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"true", true},
		{"false", false},
		{"1 < 2", true},
		{"1 > 2", false},
		{"1 == 1", true},
		{"1 != 1", false},
		{"1 == 2", false},
		{"1 != 2", true},
		{"true == true", true},
		{"false == false", true},
		{"true == false", false},
		{"true != false", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			assertBoolean(t, evalInput(tt.input), tt.want)
		})
	}
}

// ---------------------------------------------------------------------------
// Prefix operators
// ---------------------------------------------------------------------------

func TestBangOperator(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"!true", false},
		{"!false", true},
		{"!!true", true},
		{"!!false", false},
		{"!5", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			assertBoolean(t, evalInput(tt.input), tt.want)
		})
	}
}

// ---------------------------------------------------------------------------
// String concatenation
// ---------------------------------------------------------------------------

func TestStringConcatenation(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{`"hello" + " " + "world"`, "hello world"},
		{`"foo" + "bar"`, "foobar"},
		{`"" + "x"`, "x"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			assertString(t, evalInput(tt.input), tt.want)
		})
	}
}

// ---------------------------------------------------------------------------
// Let statements
// ---------------------------------------------------------------------------

func TestLetStatements(t *testing.T) {
	tests := []struct {
		input string
		want  int64
	}{
		{"let x = 5; x", 5},
		{"let x = 5 * 5; x", 25},
		{"let a = 5; let b = a; b", 5},
		{"let a = 5; let b = a; let c = a + b + 5; c", 15},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			assertInteger(t, evalInput(tt.input), tt.want)
		})
	}
}

// ---------------------------------------------------------------------------
// Variable reassignment
// ---------------------------------------------------------------------------

func TestAssignStatement(t *testing.T) {
	tests := []struct {
		input string
		want  int64
	}{
		{"let x = 5; x = 10; x", 10},
		{"let x = 1; x = x + 1; x", 2},
		{"let a = 10; let b = 20; a = b; a", 20},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			assertInteger(t, evalInput(tt.input), tt.want)
		})
	}
}

func TestAssignUndeclaredVariable(t *testing.T) {
	obj := evalInput("x = 5")
	if _, ok := obj.(*object.Error); !ok {
		t.Errorf("expected *object.Error for assignment to undeclared variable, got %T", obj)
	}
}

func TestAssignUpdatesOuterScope(t *testing.T) {
	// Reassignment inside a function body should update the closed-over variable.
	input := `
let x = 1;
let f = fn() { x = 99; };
f();
x
`
	assertInteger(t, evalInput(input), 99)
}

// ---------------------------------------------------------------------------
// If / else expressions
// ---------------------------------------------------------------------------

func TestIfElseExpression(t *testing.T) {
	tests := []struct {
		input    string
		wantNull bool
		wantInt  int64
	}{
		{"if (true) { 10 }", false, 10},
		{"if (false) { 10 }", true, 0},
		{"if (1 < 2) { 10 }", false, 10},
		{"if (1 > 2) { 10 }", true, 0},
		{"if (1 > 2) { 10 } else { 20 }", false, 20},
		{"if (1 < 2) { 10 } else { 20 }", false, 10},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			obj := evalInput(tt.input)
			if tt.wantNull {
				assertNull(t, obj)
			} else {
				assertInteger(t, obj, tt.wantInt)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Return statements
// ---------------------------------------------------------------------------

func TestReturnStatements(t *testing.T) {
	tests := []struct {
		input string
		want  int64
	}{
		{"return 10;", 10},
		{"return 10; 9;", 10},
		{"return 2 * 5; 9;", 10},
		{"9; return 2 * 5; 9;", 10},
		{`
if (10 > 1) {
  if (10 > 1) {
    return 10;
  }
  return 1;
}`, 10},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			assertInteger(t, evalInput(tt.input), tt.want)
		})
	}
}

// ---------------------------------------------------------------------------
// Functions and calls
// ---------------------------------------------------------------------------

func TestFunctionObject(t *testing.T) {
	obj := evalInput("fn(x) { x + 2; }")
	fn, ok := obj.(*object.Function)
	if !ok {
		t.Fatalf("object is %T, want *object.Function", obj)
	}
	if len(fn.Parameters) != 1 {
		t.Fatalf("Parameters count = %d, want 1", len(fn.Parameters))
	}
	if fn.Parameters[0].Value != "x" {
		t.Errorf("Parameters[0] = %q, want %q", fn.Parameters[0].Value, "x")
	}
}

func TestFunctionCall(t *testing.T) {
	tests := []struct {
		input string
		want  int64
	}{
		{"let identity = fn(x) { x; }; identity(5);", 5},
		{"let identity = fn(x) { return x; }; identity(5);", 5},
		{"let double = fn(x) { x * 2; }; double(5);", 10},
		{"let add = fn(x, y) { x + y; }; add(5, 5);", 10},
		{"let add = fn(x, y) { x + y; }; add(5 + 5, add(5, 5));", 20},
		{"fn(x) { x; }(5)", 5},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			assertInteger(t, evalInput(tt.input), tt.want)
		})
	}
}

func TestClosures(t *testing.T) {
	input := `
let newAdder = fn(x) {
  fn(y) { x + y };
};
let addTwo = newAdder(2);
addTwo(3);
`
	assertInteger(t, evalInput(input), 5)
}

// ---------------------------------------------------------------------------
// Error handling
// ---------------------------------------------------------------------------

func TestErrorHandling(t *testing.T) {
	tests := []struct {
		input   string
		wantMsg string
	}{
		{"5 + true;", "unknown operator: INTEGER + BOOLEAN"},
		{"5 + true; 5;", "unknown operator: INTEGER + BOOLEAN"},
		{"-true", "unknown operator: -BOOLEAN"},
		{"true + false;", "unknown operator: BOOLEAN + BOOLEAN"},
		{"5; true + false; 5", "unknown operator: BOOLEAN + BOOLEAN"},
		{"if (10 > 1) { true + false; }", "unknown operator: BOOLEAN + BOOLEAN"},
		{"foobar", "identifier not found: foobar"},
		{`"hello" - "world"`, "Unknown operator: &{hello} - &{world}"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			assertError(t, evalInput(tt.input), tt.wantMsg)
		})
	}
}
