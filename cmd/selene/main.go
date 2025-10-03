package main

import (
	"flag"
	"fmt"
	"os"

	"selenelang/internal/lexer"
	"selenelang/internal/parser"
	"selenelang/internal/runtime"
	"selenelang/internal/token"
)

func main() {
	dumpTokens := flag.Bool("tokens", false, "print the token stream")
	flag.Parse()

	if flag.NArg() == 0 {
		fmt.Fprintln(os.Stderr, "usage: selene [--tokens] <file>")
		os.Exit(1)
	}

	filename := flag.Arg(0)
	content, err := os.ReadFile(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to read %s: %v\n", filename, err)
		os.Exit(1)
	}

	l := lexer.New(string(content))
	if *dumpTokens {
		for tok := l.NextToken(); tok.Type != token.EOF; tok = l.NextToken() {
			fmt.Printf("%s\t%q\n", tok.Type, tok.Literal)
		}
		return
	}

	l = lexer.New(string(content))
	p := parser.New(l)
	program := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		for _, e := range errs {
			fmt.Fprintln(os.Stderr, "parse error:", e)
		}
		os.Exit(1)
	}

	rt := runtime.New()
	if _, err := rt.Run(program); err != nil {
		fmt.Fprintln(os.Stderr, "runtime error:", err)
		os.Exit(1)
	}
}
