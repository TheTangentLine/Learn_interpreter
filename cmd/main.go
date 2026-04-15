package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/thetangentline/interpreter/internal/evaluator"
	"github.com/thetangentline/interpreter/internal/lexer"
	"github.com/thetangentline/interpreter/internal/object"
	"github.com/thetangentline/interpreter/internal/parser"
	"github.com/thetangentline/interpreter/internal/repl"
)

func checkFileExtension(path string) bool {
	return filepath.Ext(path) == ".dot"
}

func main() {
	args := os.Args[1:]

	if len(args) == 0 {
		repl.Start(os.Stdin, os.Stdout)
		return
	}

	if !checkFileExtension(args[0]) {
		fmt.Fprintln(os.Stderr, "file must have a .dot extension")
		os.Exit(1)
	}

	content, err := os.ReadFile(args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading file: %s\n", err)
		os.Exit(1)
	}

	l := lexer.New(string(content))
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		for _, msg := range p.Errors() {
			fmt.Fprintf(os.Stderr, "\t%s\n", msg)
		}
		os.Exit(1)
	}

	env := object.NewEnvironment()
	result := evaluator.Eval(program, env)
	if result != nil {
		fmt.Println(result.Inspect())
	}
}
