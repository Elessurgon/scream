package main

import (
	_ "embed"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"scream/evaluator"
	"scream/lexer"
	"scream/object"
	"scream/parser"
)

var version = "master/unreleased"

var stdlib string

func versionFun(args ...object.Object) object.Object {
	return &object.String{Value: version}
}

func argsFun(args ...object.Object) object.Object {
	l := len(os.Args[1:])
	result := make([]object.Object, l)
	for i, txt := range os.Args[1:] {
		result[i] = &object.String{Value: txt}
	}
	return &object.Array{Elements: result}
}

func Execute(input string) int {

	env := object.NewEnvironment()
	l := lexer.New(input)
	p := parser.New(l)

	program := p.ParseProgram()
	if len(p.Errors()) != 0 {
		for _, msg := range p.Errors() {
			fmt.Printf("\t%s\n", msg)
		}
		os.Exit(1)
	}

	evaluator.RegisterBuiltin("version",
		func(env *object.Environment, args ...object.Object) object.Object {
			return (versionFun(args...))
		})

	evaluator.RegisterBuiltin("args",
		func(env *object.Environment, args ...object.Object) object.Object {
			return (argsFun(args...))
		})

	initL := lexer.New(stdlib)
	initP := parser.New(initL)
	initProg := initP.ParseProgram()
	evaluator.Eval(initProg, env)

	evaluator.Eval(program, env)
	return 0
}

func main() {

	eval := flag.String("eval", "", "Code to execute.")
	vers := flag.Bool("version", false, "Show our version and exit.")

	flag.Parse()

	if *vers {
		fmt.Printf("monkey %s\n", version)
		os.Exit(1)
	}

	if *eval != "" {
		Execute(*eval)
		os.Exit(1)
	}

	var input []byte
	var err error

	if len(flag.Args()) > 0 {
		input, err = ioutil.ReadFile(os.Args[1])
	} else {
		input, err = ioutil.ReadAll(os.Stdin)
	}

	if err != nil {
		fmt.Printf("Error reading: %s\n", err.Error())
	}

	Execute(string(input))
}
