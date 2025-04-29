package syntax

import "fmt"

type TokenType int

const (
	TOKEN_LEFT_PAREN TokenType = iota + 1
	TOKEN_RIGHT_PAREN
	TOKEN_LEFT_BRACE
	TOKEN_RIGHT_BRACE
	TOKEN_COMMA
	TOKEN_DOT
	TOKEN_MINUS
	TOKEN_PLUS
	TOKEN_SEMICOLON
	TOKEN_SLASH
	TOKEN_STAR

	TOKEN_BANG
	TOKEN_BANG_EQUAL
	TOKEN_EQUAL
	TOKEN_EQUAL_EQUAL
	TOKEN_GREATER
	TOKEN_GREATER_EQUAL
	TOKEN_LESS
	TOKEN_LESS_EQUAL

	// literals
	TOKEN_IDENTIFIER
	TOKEN_STRING
	TOKEN_NUMBER

	// keywords
	TOKEN_AND
	TOKEN_CLASS
	TOKEN_ELSE
	TOKEN_FALSE
	TOKEN_FUN
	TOKEN_FOR
	TOKEN_IF
	TOKEN_NIL
	TOKEN_OR
	TOKEN_PRINT
	TOKEN_RETURN
	TOKEN_SUPER
	TOKEN_THIS
	TOKEN_TRUE
	TOKEN_VAR
	TOKEN_WHILE
	TOKEN_BREAK

	TOKEN_EOF
)

var (
	TokenTypeStr = map[TokenType]string{
		TOKEN_LEFT_PAREN:  "(",
		TOKEN_RIGHT_PAREN: ")",
		TOKEN_LEFT_BRACE:  "{",
		TOKEN_RIGHT_BRACE: "}",
		TOKEN_COMMA:       ",",
		TOKEN_DOT:         ".",
		TOKEN_MINUS:       "-",
		TOKEN_PLUS:        "+",
		TOKEN_SEMICOLON:   ";",
		TOKEN_SLASH:       "/",
		TOKEN_STAR:        "*",

		TOKEN_BANG:          "!",
		TOKEN_BANG_EQUAL:    "!=",
		TOKEN_EQUAL:         "=",
		TOKEN_EQUAL_EQUAL:   "==",
		TOKEN_GREATER:       ">",
		TOKEN_GREATER_EQUAL: ">=",
		TOKEN_LESS:          "<",
		TOKEN_LESS_EQUAL:    "<=",

		TOKEN_IDENTIFIER: "identifier",
		TOKEN_STRING:     "string",
		TOKEN_NUMBER:     "number",

		TOKEN_AND:    "and",
		TOKEN_CLASS:  "class",
		TOKEN_ELSE:   "else",
		TOKEN_FALSE:  "false",
		TOKEN_FUN:    "fun",
		TOKEN_FOR:    "for",
		TOKEN_IF:     "if",
		TOKEN_NIL:    "nil",
		TOKEN_OR:     "or",
		TOKEN_PRINT:  "print",
		TOKEN_RETURN: "return",
		TOKEN_SUPER:  "super",
		TOKEN_THIS:   "this",
		TOKEN_TRUE:   "true",
		TOKEN_VAR:    "var",
		TOKEN_WHILE:  "while",
		TOKEN_BREAK:  "break",

		TOKEN_EOF: "EOF",
	}
)

type Token struct {
	TokenType TokenType
	Lexeme    string
	Literal   any
	Line      int
}

func NewToken(tokenType TokenType, lexeme string, literal any, line int) Token {
	return Token{tokenType, lexeme, literal, line}
}

func (t Token) String() string {
	if t.Lexeme == "" {
		return fmt.Sprintf("token: {type: %s literal:%v, line: %d}",
			TokenTypeStr[t.TokenType], t.Literal, t.Line)
	}
	return fmt.Sprintf("token: {type: %s lexeme:%s literal:%v, line: %d}",
		TokenTypeStr[t.TokenType], t.Lexeme, t.Literal, t.Line)
}
