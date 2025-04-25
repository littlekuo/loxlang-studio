package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/littlekuo/glox-treewalk/internal/syntax"
)

var (
	filePath string
	content  string
)

func main() {
	fs := flag.NewFlagSet("ast-printer", flag.ExitOnError)
	fs.StringVar(&filePath, "filePath", "", "path to the source file")
	fs.StringVar(&content, "content", "", "source code")
	if len(os.Args[1:]) == 0 {
		fs.Usage()
		return
	}
	if err := fs.Parse(os.Args[1:]); err != nil {
		fmt.Printf("parse failed, err [%s]", err.Error())
		return
	}
	if filePath == "" && content == "" {
		fmt.Println("file path and content are both empty")
		return
	}
	if filePath != "" && content != "" {
		fmt.Println("file path and content are both set")
		return
	}
	var source []byte
	if content != "" {
		source = []byte(content)
	} else {
		var err error
		source, err = os.ReadFile(filePath)
		if err != nil {
			fmt.Printf("read file failed, err [%s]", err.Error())
			return
		}
	}
	scanner := syntax.NewScanner(string(source))
	tokens := scanner.ScanTokens()
	if err := scanner.GetError(); err != nil {
		fmt.Printf("scan tokens failed, err [%s]", err.Error())
	}
	parser := syntax.NewParser(tokens)
	stmts := parser.Parse()
	if err := parser.GetError(); err != nil {
		fmt.Printf("parse failed, err [%s]", err.Error())
	}
	astPrinter := &syntax.AstPrinter{}
	astPrinter.PrintStmts(stmts)
}
