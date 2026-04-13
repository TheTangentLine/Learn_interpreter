package evaluator

import(
	"github.com/thetangentline/interpreter/internal/object"
	"github.com/thetangentline/interpreter/internal/ast"
)

func evalCallExpression(node *ast.CallExpression, env *object.Environment) object.Object {
    function := Eval(node.Function, env)
    if isError(function) { return function }

    args := evalExpressions(node.Arguments, env)
    if len(args) == 1 && isError(args[0]) { return args[0] }

    return applyFunction(function, args)
}

func evalExpressions(exps []ast.Expression, env *object.Environment) []object.Object


func applyFunction(fn object.Object, args []object.Object) object.Object {
    function, ok := fn.(*object.Function)
    if !ok {
        return newError("not a function: %s", fn.Type())
    }

    extendedEnv := extendFunctionEnv(function, args)
    evaluated := Eval(function.Body, extendedEnv)
    return unwrapReturnValue(evaluated)
}

func extendFunctionEnv(fn *object.Function, args []object.Object) *object.Environment {
    env := object.NewEnclosedEnvironment(fn.Env) // new scope, outer = closure env
    for i, param := range fn.Parameters {
        env.Define(param.Value, args[i]) // bind parameter names to argument values
    }
    return env
}

func unwrapReturnValue(obj object.Object) object.Object {
    if returnVal, ok := obj.(*object.ReturnValue); ok {
        return returnVal.Value
    }
    return obj
}