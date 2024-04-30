package main

import (
	"fmt"
	"gorilla/evaluator"
	"gorilla/lexer"
	"gorilla/object"
	"gorilla/parser"
	"gorilla/repl"
	"io"
	"os"
	"os/user"
)

func check(e error) {
	if e != nil {
		fmt.Println(e)
		os.Exit(1)
	}
}

const GORILLA_FACE = "🦍"

func printParserErrors(out io.Writer, errors []string) {
	io.WriteString(out, GORILLA_FACE)
	io.WriteString(out, "\nWoops! We ran into some gorilla business here!\n")
	io.WriteString(out, " parser errors:\n")
	for _, msg := range errors {
		io.WriteString(out, "\t"+msg+"\n")
	}
}

func runFromFile(file string) {
	dat, err := os.ReadFile(file)
	check(err)

	env := object.NewEnvironment()
	out := os.Stdout

	line := string(dat)
	l := lexer.New(line)
	p := parser.New(l)

	program := p.ParseProgram()
	if len(p.Errors()) != 0 {
		printParserErrors(out, p.Errors())
		os.Exit(1)
	}

	evaluated := evaluator.Eval(program, env)
	if evaluated != nil {
		evaluated.Inspect()
	}

}

func runRepl() {
	user, err := user.Current()
	if err != nil {
		panic(err)
	}
	fmt.Printf("Hello %s! This is the Gorilla programming language!\n",
		user.Username)
	fmt.Printf("Feel free to type in commands\n")
	repl.Start(os.Stdin, os.Stdout)
}

func main() {

	if len(os.Args) == 1 {
		runRepl()
	} else {
		runFromFile(os.Args[1])
	}

}
