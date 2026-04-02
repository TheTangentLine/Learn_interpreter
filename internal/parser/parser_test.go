package parser

import (
	"fmt"
	"testing"

	"github.com/thetangentline/interpreter/internal/ast"
	"github.com/thetangentline/interpreter/internal/lexer"
)

// parseProgram is a test helper that parses input and fails immediately if
// there are any parse errors or the wrong number of statements.
func parseProgram(t *testing.T, input string, wantStmts int) *ast.Program {
	t.Helper()
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()

	if errs := p.Errors(); len(errs) != 0 {
		t.Fatalf("parser produced %d error(s):\n%s", len(errs), formatErrors(errs))
	}
	if len(program.Statements) != wantStmts {
		t.Fatalf("expected %d statement(s), got %d", wantStmts, len(program.Statements))
	}
	return program
}

func formatErrors(errs []string) string {
	out := ""
	for _, e := range errs {
		out += "\t" + e + "\n"
	}
	return out
}

// expressionStmt extracts the Expression from statement index i, failing if the
// statement is not an *ast.ExpressionStatement.
func expressionStmt(t *testing.T, program *ast.Program, i int) ast.Expression {
	t.Helper()
	stmt, ok := program.Statements[i].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("statement[%d] is %T, want *ast.ExpressionStatement", i, program.Statements[i])
	}
	return stmt.Expression
}

// ---------------------------------------------------------------------------
// Let statements
// ---------------------------------------------------------------------------

func TestLetStatements(t *testing.T) {
	tests := []struct {
		input     string
		wantName  string
		wantValue string
	}{
		{"let x = 5;", "x", "5"},
		{"let y = true;", "y", "true"},
		{"let foobar = y;", "foobar", "y"},
		{"let result = 2 + 3 * 4;", "result", "2 + 3 * 4"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			program := parseProgram(t, tt.input, 1)
			stmt, ok := program.Statements[0].(*ast.LetStatement)
			if !ok {
				t.Fatalf("got %T, want *ast.LetStatement", program.Statements[0])
			}
			if stmt.Name.Value != tt.wantName {
				t.Errorf("Name = %q, want %q", stmt.Name.Value, tt.wantName)
			}
			if stmt.Value.String() != tt.wantValue {
				t.Errorf("Value = %q, want %q", stmt.Value.String(), tt.wantValue)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Return statements
// ---------------------------------------------------------------------------

func TestReturnStatements(t *testing.T) {
	tests := []struct {
		input     string
		wantValue string
	}{
		{"return 5;", "5"},
		{"return true;", "true"},
		{"return foobar;", "foobar"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			program := parseProgram(t, tt.input, 1)
			stmt, ok := program.Statements[0].(*ast.ReturnStatement)
			if !ok {
				t.Fatalf("got %T, want *ast.ReturnStatement", program.Statements[0])
			}
			if stmt.ReturnValue.String() != tt.wantValue {
				t.Errorf("ReturnValue = %q, want %q", stmt.ReturnValue.String(), tt.wantValue)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Literals
// ---------------------------------------------------------------------------

func TestIdentifierExpression(t *testing.T) {
	program := parseProgram(t, "foobar;", 1)
	ident, ok := expressionStmt(t, program, 0).(*ast.Identifier)
	if !ok {
		t.Fatalf("expression is %T, want *ast.Identifier", expressionStmt(t, program, 0))
	}
	if ident.Value != "foobar" {
		t.Errorf("Value = %q, want %q", ident.Value, "foobar")
	}
}

func TestIntegerLiteralExpression(t *testing.T) {
	program := parseProgram(t, "42;", 1)
	lit, ok := expressionStmt(t, program, 0).(*ast.IntegerLiteral)
	if !ok {
		t.Fatalf("expression is %T, want *ast.IntegerLiteral", expressionStmt(t, program, 0))
	}
	if lit.Value != 42 {
		t.Errorf("Value = %d, want 42", lit.Value)
	}
}

func TestStringLiteralExpression(t *testing.T) {
	program := parseProgram(t, `"hello world";`, 1)
	lit, ok := expressionStmt(t, program, 0).(*ast.StringLiteral)
	if !ok {
		t.Fatalf("expression is %T, want *ast.StringLiteral", expressionStmt(t, program, 0))
	}
	if lit.Value != "hello world" {
		t.Errorf("Value = %q, want %q", lit.Value, "hello world")
	}
}

func TestBooleanExpression(t *testing.T) {
	tests := []struct{ input string; want bool }{
		{"true;", true},
		{"false;", false},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			program := parseProgram(t, tt.input, 1)
			b, ok := expressionStmt(t, program, 0).(*ast.Boolean)
			if !ok {
				t.Fatalf("expression is %T, want *ast.Boolean", expressionStmt(t, program, 0))
			}
			if b.Value != tt.want {
				t.Errorf("Value = %v, want %v", b.Value, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Prefix expressions
// ---------------------------------------------------------------------------

func TestPrefixExpressions(t *testing.T) {
	tests := []struct {
		input    string
		operator string
		right    string
	}{
		{"!true;", "!", "true"},
		{"!false;", "!", "false"},
		{"-5;", "-", "5"},
		{"-foobar;", "-", "foobar"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			program := parseProgram(t, tt.input, 1)
			expr, ok := expressionStmt(t, program, 0).(*ast.PrefixExpression)
			if !ok {
				t.Fatalf("expression is %T, want *ast.PrefixExpression", expressionStmt(t, program, 0))
			}
			if expr.Operator != tt.operator {
				t.Errorf("Operator = %q, want %q", expr.Operator, tt.operator)
			}
			if expr.Right.String() != tt.right {
				t.Errorf("Right = %q, want %q", expr.Right.String(), tt.right)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Infix expressions
// ---------------------------------------------------------------------------

func TestInfixExpressions(t *testing.T) {
	tests := []struct {
		input    string
		left     string
		operator string
		right    string
	}{
		{"5 + 3;", "5", "+", "3"},
		{"5 - 3;", "5", "-", "3"},
		{"5 * 3;", "5", "*", "3"},
		{"5 / 3;", "5", "/", "3"},
		{"5 == 5;", "5", "==", "5"},
		{"5 != 3;", "5", "!=", "3"},
		{"5 < 10;", "5", "<", "10"},
		{"5 > 1;", "5", ">", "1"},
		{"a + b;", "a", "+", "b"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			program := parseProgram(t, tt.input, 1)
			expr, ok := expressionStmt(t, program, 0).(*ast.InfixExpression)
			if !ok {
				t.Fatalf("expression is %T, want *ast.InfixExpression", expressionStmt(t, program, 0))
			}
			if expr.Left.String() != tt.left {
				t.Errorf("Left = %q, want %q", expr.Left.String(), tt.left)
			}
			if expr.Operator != tt.operator {
				t.Errorf("Operator = %q, want %q", expr.Operator, tt.operator)
			}
			if expr.Right.String() != tt.right {
				t.Errorf("Right = %q, want %q", expr.Right.String(), tt.right)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Operator precedence — verified via String() output
// ---------------------------------------------------------------------------

func TestOperatorPrecedence(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"2 + 3 * 4", "2 + 3 * 4"},
		{"2 * 3 + 4", "2 * 3 + 4"},
		{"-a * b", "-a * b"},
		{"!-a", "!-a"},
		{"a + b + c", "a + b + c"},
		{"a + b * c + d / e - f", "a + b * c + d / e - f"},
		{"(a + b) * c", "a + b * c"},
		{"a * (b + c)", "a * b + c"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			program := parseProgram(t, tt.input, 1)
			got := expressionStmt(t, program, 0).String()
			if got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// If expression
// ---------------------------------------------------------------------------

func TestIfExpression(t *testing.T) {
	program := parseProgram(t, "if (x < y) { x }", 1)
	expr, ok := expressionStmt(t, program, 0).(*ast.IfExpression)
	if !ok {
		t.Fatalf("expression is %T, want *ast.IfExpression", expressionStmt(t, program, 0))
	}
	if expr.Condition.String() != "x < y" {
		t.Errorf("Condition = %q, want %q", expr.Condition.String(), "x < y")
	}
	if len(expr.Consequence.Statements) != 1 {
		t.Fatalf("Consequence has %d statement(s), want 1", len(expr.Consequence.Statements))
	}
	if expr.Alternative != nil {
		t.Errorf("Alternative should be nil for plain if")
	}
}

func TestIfElseExpression(t *testing.T) {
	program := parseProgram(t, "if (x < y) { x } else { y }", 1)
	expr, ok := expressionStmt(t, program, 0).(*ast.IfExpression)
	if !ok {
		t.Fatalf("expression is %T, want *ast.IfExpression", expressionStmt(t, program, 0))
	}
	if expr.Condition.String() != "x < y" {
		t.Errorf("Condition = %q, want %q", expr.Condition.String(), "x < y")
	}
	if len(expr.Consequence.Statements) != 1 {
		t.Fatalf("Consequence has %d statement(s), want 1", len(expr.Consequence.Statements))
	}
	if expr.Alternative == nil {
		t.Fatal("Alternative should not be nil for if-else")
	}
	if len(expr.Alternative.Statements) != 1 {
		t.Fatalf("Alternative has %d statement(s), want 1", len(expr.Alternative.Statements))
	}
}

// ---------------------------------------------------------------------------
// Function literals
// ---------------------------------------------------------------------------

func TestFunctionLiteral(t *testing.T) {
	program := parseProgram(t, "fn(x, y) { x + y; }", 1)
	fn, ok := expressionStmt(t, program, 0).(*ast.FunctionLiteral)
	if !ok {
		t.Fatalf("expression is %T, want *ast.FunctionLiteral", expressionStmt(t, program, 0))
	}
	if len(fn.Parameters) != 2 {
		t.Fatalf("Parameters count = %d, want 2", len(fn.Parameters))
	}
	if fn.Parameters[0].Value != "x" {
		t.Errorf("Parameters[0] = %q, want %q", fn.Parameters[0].Value, "x")
	}
	if fn.Parameters[1].Value != "y" {
		t.Errorf("Parameters[1] = %q, want %q", fn.Parameters[1].Value, "y")
	}
	if len(fn.Body.Statements) != 1 {
		t.Fatalf("Body has %d statement(s), want 1", len(fn.Body.Statements))
	}
}

func TestFunctionLiteralNoParams(t *testing.T) {
	program := parseProgram(t, "fn() { 42 }", 1)
	fn, ok := expressionStmt(t, program, 0).(*ast.FunctionLiteral)
	if !ok {
		t.Fatalf("expression is %T, want *ast.FunctionLiteral", expressionStmt(t, program, 0))
	}
	if len(fn.Parameters) != 0 {
		t.Errorf("Parameters count = %d, want 0", len(fn.Parameters))
	}
}

// ---------------------------------------------------------------------------
// Call expressions
// ---------------------------------------------------------------------------

func TestCallExpression(t *testing.T) {
	program := parseProgram(t, "add(1, 2 * 3, 4 + 5);", 1)
	call, ok := expressionStmt(t, program, 0).(*ast.CallExpression)
	if !ok {
		t.Fatalf("expression is %T, want *ast.CallExpression", expressionStmt(t, program, 0))
	}
	if call.Function.String() != "add" {
		t.Errorf("Function = %q, want %q", call.Function.String(), "add")
	}
	if len(call.Arguments) != 3 {
		t.Fatalf("Arguments count = %d, want 3", len(call.Arguments))
	}
	if call.Arguments[0].String() != "1" {
		t.Errorf("Arguments[0] = %q, want %q", call.Arguments[0].String(), "1")
	}
	if call.Arguments[1].String() != "2 * 3" {
		t.Errorf("Arguments[1] = %q, want %q", call.Arguments[1].String(), "2 * 3")
	}
	if call.Arguments[2].String() != "4 + 5" {
		t.Errorf("Arguments[2] = %q, want %q", call.Arguments[2].String(), "4 + 5")
	}
}

func TestCallExpressionNoArgs(t *testing.T) {
	program := parseProgram(t, "myFunc();", 1)
	call, ok := expressionStmt(t, program, 0).(*ast.CallExpression)
	if !ok {
		t.Fatalf("expression is %T, want *ast.CallExpression", expressionStmt(t, program, 0))
	}
	if call.Function.String() != "myFunc" {
		t.Errorf("Function = %q, want %q", call.Function.String(), "myFunc")
	}
	if len(call.Arguments) != 0 {
		t.Errorf("Arguments count = %d, want 0", len(call.Arguments))
	}
}

// ---------------------------------------------------------------------------
// Error reporting
// ---------------------------------------------------------------------------

func TestParserErrors(t *testing.T) {
	tests := []struct {
		input    string
		wantErrs int
	}{
		{"let x 5;", 1},    // missing =
		{"let = 5;", 2},    // missing identifier → expectPeek error, then = treated as expression
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			p.ParseProgram()
			if got := len(p.Errors()); got != tt.wantErrs {
				t.Errorf("got %d error(s), want %d: %v", got, tt.wantErrs, p.Errors())
			}
		})
	}
}

// Ensure the parser does not panic on empty input.
func TestEmptyInput(t *testing.T) {
	program := parseProgram(t, "", 0)
	if program == nil {
		t.Fatal("ParseProgram returned nil")
	}
}

var _ = fmt.Sprintf // keep fmt import used by formatErrors
