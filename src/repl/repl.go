package repl

import (
	"bufio"
	"fmt"
	"io"
	"monkey/evaluator"
	"monkey/lexer"
	"monkey/object"
	"monkey/parser"
)

const PROMPT = ">> "

func Start(in io.Reader, out io.Writer) {
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
			io.WriteString(out, evaluated.Inspect())
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

const MONKEY_FACE = "🐒"

func printParserErrors(out io.Writer, errors []string) {
	io.WriteString(out, MONKEY_FACE)
	io.WriteString(out, "\nWoops! We ran into some monkey business here!\n")
	io.WriteString(out, " parser errors:\n")
	for _, msg := range errors {
		io.WriteString(out, "\t"+msg+"\n")
	}
}
