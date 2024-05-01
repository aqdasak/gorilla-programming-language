package main

import (
	"bufio"
	"fmt"
	"gorilla/debug"
	"gorilla/evaluator"
	"gorilla/lexer"
	"gorilla/object"
	"gorilla/parser"
	"io"
	"os"
)

const PROMPT = ">> "
const GORILLA_FACE = "🦍"

func check(e error) {
	if e != nil {
		fmt.Println(e)
		os.Exit(1)
	}
}

func printParserErrors(out io.Writer, errors []string) {
	io.WriteString(out, GORILLA_FACE)
	io.WriteString(out, "\nWoops! We ran into some gorilla business here!\n")
	io.WriteString(out, " parser errors:\n")
	for _, msg := range errors {
		io.WriteString(out, "\t"+msg+"\n")
	}
}

func runRepl(in io.Reader, out io.Writer) {
	fmt.Println("Gorilla 1.0.2 (main, Apr 30 2024)")
	fmt.Println(`
	⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢀⣠⣤⣤⠀⠀⠀⠀
	⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢾⣿⣿⣿⣿⣄⠀⠀⠀
	⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢠⣴⣿⣿⣶⣄⠹⣿⣿⣿⡟⠁⠀⠀
	⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⣴⣿⣿⣿⣿⣿⣿⡆⢹⣿⣿⣿⣷⡀⠀
	⠀⠀⠀⠀⠀⠀⣀⣀⣀⣀⣀⣀⣀⣠⣾⣿⣿⣿⣿⣿⣿⣿⣿⠀⢿⣿⣿⣿⡇⠀
	⠀⠀⠀⠀⣠⣾⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⡆⢸⣿⣿⠟⠁⠀
	⠀⠀⠀⣼⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⡏⠹⣿⣿⣿⣿⣷⠀⠀⠀⠀⠀⠀
	⠀⠀⠀⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⡄⢻⣿⣿⣿⣿⡆⠀⠀⠀⠀⠀
	⠀⠀⠀⣿⣿⣿⣿⣿⣿⠿⣿⣿⣿⣿⣿⣿⣿⣿⣷⠀⢿⣿⣿⣿⣿⡄⠀⠀⠀⠀
	⠀⠀⢀⣿⣿⣿⣿⣿⡟⢀⣿⣿⣿⣿⣿⣿⡿⠟⢁⡄⠸⣿⣿⣿⣿⣷⠀⠀⠀⠀
	⠀⠀⣼⣿⣿⣿⣿⠏⠀⣈⡙⠛⢛⠋⠉⠁⠀⣸⣿⣿⠀⢻⣿⣿⣿⣿⡆⠀⠀⠀
	⠀⢠⣿⣿⣿⣿⣟⠀⠀⢿⣿⣿⣿⡄⠀⠀⢀⣿⣿⡟⠃⣸⣿⣿⣿⣿⡇⠀⠀⠀
	⠀⠘⠛⠛⠛⠛⠛⠛⠀⠘⠛⠛⠛⠛⠓⠀⠛⠛⠛⠃⠘⠛⠛⠛⠛⠛⠃⠀⠀⠀`)

	scanner := bufio.NewScanner(in)
	env := object.NewEnvironment()

	// for {
	// 	fmt.Printf(PROMPT)
	// 	scanned := scanner.Scan()

	// 	if !scanned {
	// 		return
	// 	}

	// 	line := scanner.Text()
	// 	l := lexer.New(line)

	// 	for tok := l.NextToken(); tok.Type != token.EOF; tok = l.NextToken() {
	// 		fmt.Printf("%+v\n", tok)
	// 	}

	// 	p := parser.New(lexer.New((line)))
	// 	program := p.ParseProgram()

	// 	fmt.Printf("\nAST: %+v\n", program.String())
	// }

	for {
		fmt.Print(PROMPT)
		scanned := scanner.Scan()
		if !scanned {
			return
		}

		line := scanner.Text()
		l := lexer.New(line)
		p := parser.New(l)

		program := p.ParseProgram()
		if len(p.Errors()) != 0 {
			printParserErrors(out, p.Errors())
			continue
		}

		// io.WriteString(out, program.String())
		// io.WriteString(out, "\n")

		evaluated := evaluator.Eval(program, env)
		if evaluated != nil {
			if evaluated != evaluator.NULL {
				io.WriteString(out, evaluated.Inspect())
			}
			io.WriteString(out, "\n")
		}

		// p = parser.New(lexer.New((line)))
		// program = p.ParseProgram()
		// fmt.Printf("\nAST: %+v\n", program.String())

		// l = lexer.New(line)
		// for tok := l.NextToken(); tok.Type != token.EOF; tok = l.NextToken() {
		// 	fmt.Printf("%+v\n", tok)
		// }

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

func main() {
	debug.PRINTEVALUATION = false
	atleast_args := 1

	idx := 0
	for i, arg := range os.Args[1:] {
		if arg == "--debug" {
			debug.PRINTEVALUATION = true
			idx = i + 1
			atleast_args = 2
		}
	}

	// 			len	idx atleast
	// g		1	0	1
	// g -d		2	1	2
	// g f		2	0	1
	// g -d f	3	1	2
	// g f -d	3	2	2

	if len(os.Args) == atleast_args {
		runRepl(os.Stdin, os.Stdout)
	} else if len(os.Args) == atleast_args+1 {
		runFromFile(os.Args[idx%2+1])
	} else {
		fmt.Println("Wrong arguments")
	}
}
