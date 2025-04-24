package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/littlekuo/go-lox/internal/syntax"
)

func main() {
	args := os.Args[1:]

	switch len(args) {
	case 0:
		runPrompt()
	case 1:
		runFile(args[0])
	default:
		fmt.Println("Usage: glox [script]")
		os.Exit(64)
	}
}

func runFile(path string) error {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if err := run(string(bytes)); err != nil {
		os.Exit(65)
	}
	return nil
}

func runPrompt() {
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("> ")
		if !scanner.Scan() { // 处理Ctrl+D
			break
		}
		line := scanner.Text()
		run(line)
	}
}

func run(source string) error {
	scanner := syntax.NewScanner(source)
	tokens := scanner.ScanTokens()
	parser := syntax.NewParser(tokens)
	expr := parser.Parse()
	if err := scanner.GetError(); err != nil {
		return err
	}
	if err := parser.GetError(); err != nil {
		return err
	}
	astPrinter := syntax.AstPrinter{}
	fmt.Println(astPrinter.Print(expr))
	return nil
}
