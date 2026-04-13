package evaluator

import(
	"github.com/thetangentline/interpreter/internal/object"
)

func isError(obj object.Object) bool {
    return obj != nil && obj.Type() == object.ERROR_OBJ
}

func newError(format string, a ...interface{}) *object.Error