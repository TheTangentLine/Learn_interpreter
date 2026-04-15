package repl

import (
	"bufio"
	"fmt"
	"io"

	"github.com/thetangentline/interpreter/internal/evaluator"
	"github.com/thetangentline/interpreter/internal/lexer"
	"github.com/thetangentline/interpreter/internal/object"
	"github.com/thetangentline/interpreter/internal/parser"
)

const prompt = ">> "

func Start(in io.Reader, out io.Writer) {
	scanner := bufio.NewScanner(in)
	env := object.NewEnvironment()

	for {
		fmt.Fprint(out, prompt)
		if !scanner.Scan() {
			break
		}

		line := scanner.Text()
		l := lexer.New(line)
		p := parser.New(l)
		program := p.ParseProgram()

		if len(p.Errors()) > 0 {
			for _, msg := range p.Errors() {
				fmt.Fprintf(out, "\t%s\n", msg)
			}
			continue
		}

		result := evaluator.Eval(program, env)
		if result != nil {
			fmt.Fprintln(out, result.Inspect())
		}
	}
}
