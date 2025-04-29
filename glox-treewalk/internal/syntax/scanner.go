package syntax

import (
	"fmt"
	"strconv"

	"github.com/littlekuo/glox-treewalk/internal/util"
)

var keywords = map[string]TokenType{
	"and":    TOKEN_AND,
	"class":  TOKEN_CLASS,
	"else":   TOKEN_ELSE,
	"false":  TOKEN_FALSE,
	"for":    TOKEN_FOR,
	"fun":    TOKEN_FUN,
	"if":     TOKEN_IF,
	"nil":    TOKEN_NIL,
	"or":     TOKEN_OR,
	"print":  TOKEN_PRINT,
	"return": TOKEN_RETURN,
	"super":  TOKEN_SUPER,
	"this":   TOKEN_THIS,
	"true":   TOKEN_TRUE,
	"var":    TOKEN_VAR,
	"while":  TOKEN_WHILE,
	"break":  TOKEN_BREAK,
}

type Scanner struct {
	source  string
	tokens  []Token
	start   int
	current int
	// line number
	line    int
	scanErr error
}

func NewScanner(source string) *Scanner {
	return &Scanner{
		source: source,
		tokens: make([]Token, 0),
		line:   1,
	}
}

func (s *Scanner) ScanTokens() []Token {
	for {
		if s.isEnd() {
			break
		}
		s.start = s.current
		s.scanToken()
	}

	// the last token
	s.tokens = append(s.tokens, NewToken(TOKEN_EOF, "", nil, s.line))
	return s.tokens
}

func (s *Scanner) scanToken() {
	c := s.advance()
	switch c {
	case '(':
		s.addSimpleToken(TOKEN_LEFT_PAREN)
	case ')':
		s.addSimpleToken(TOKEN_RIGHT_PAREN)
	case '{':
		s.addSimpleToken(TOKEN_LEFT_BRACE)
	case '}':
		s.addSimpleToken(TOKEN_RIGHT_BRACE)
	case ',':
		s.addSimpleToken(TOKEN_COMMA)
	case '.':
		s.addSimpleToken(TOKEN_DOT)
	case '-':
		s.addSimpleToken(TOKEN_MINUS)
	case '+':
		s.addSimpleToken(TOKEN_PLUS)
	case ';':
		s.addSimpleToken(TOKEN_SEMICOLON)
	case '*':
		s.addSimpleToken(TOKEN_STAR)
	case '!':
		s.addConditionalToken('=', TOKEN_BANG_EQUAL, TOKEN_BANG)
	case '=':
		s.addConditionalToken('=', TOKEN_EQUAL_EQUAL, TOKEN_EQUAL)
	case '<':
		s.addConditionalToken('=', TOKEN_LESS_EQUAL, TOKEN_LESS)
	case '>':
		s.addConditionalToken('=', TOKEN_GREATER_EQUAL, TOKEN_GREATER)
	case '/':
		if s.match('/') {
			// A comment goes until the end of the line.
			for s.peek() != '\n' && !s.isEnd() {
				s.advance()
			}
		} else if s.match('*') {
			s.scanBlockComment()
		} else {
			s.addSimpleToken(TOKEN_SLASH)
		}
	case ' ', '\r', '\t':
		// Ignore whitespace.
	case '\n':
		s.line++
	case '"':
		s.scanString()
	default:
		if isDigit(c) {
			s.scanNumber()
		} else if isAlpha(c) {
			s.scanIdentifier()
		} else {
			s.error(s.line, "Unexpected character.")
		}
	}
}

func (s *Scanner) addSimpleToken(tk TokenType) {
	s.addTokenWithLiteral(tk, nil)
}

func (s *Scanner) addTokenWithLiteral(tk TokenType, literal any) {
	text := s.source[s.start:s.current]
	s.tokens = append(s.tokens, NewToken(tk, text, literal, s.line))
}

func (s *Scanner) isEnd() bool {
	return s.current >= len(s.source)
}

func (s *Scanner) GetError() error {
	return s.scanErr
}

func (s *Scanner) addConditionalToken(expected byte, matchedType TokenType, unmatchedType TokenType) {
	if s.match(expected) {
		s.addSimpleToken(matchedType)
	} else {
		s.addSimpleToken(unmatchedType)
	}
}

func (s *Scanner) match(expected byte) bool {
	if s.isEnd() {
		return false
	}
	if s.source[s.current] != expected {
		return false
	}

	// if match, then advance
	s.current++
	return true
}

func (s *Scanner) advance() byte {
	s.current++
	return s.source[s.current-1]
}

func (s *Scanner) peek() byte {
	if s.isEnd() {
		return 0
	}
	return s.source[s.current]
}

func (s *Scanner) peekNext() byte {
	if s.current+1 >= len(s.source) {
		return 0
	}
	return s.source[s.current+1]
}

func (s *Scanner) scanString() {
	for s.peek() != '"' && !s.isEnd() {
		if s.peek() == '\n' {
			s.line++
		}
		s.advance()
	}

	if s.isEnd() {
		s.error(s.line, "Unterminated string")
		return
	}

	// the closing ".
	s.advance()

	s.addTokenWithLiteral(TOKEN_STRING, s.source[s.start+1:s.current-1])
}

// support: 1234, 12.34
// not support: 1234. , .1234
func (s *Scanner) scanNumber() {
	for isDigit(s.peek()) {
		s.advance()
	}
	if s.peek() == '.' && isDigit(s.peekNext()) {
		// Consume the "."
		s.advance()

		for isDigit(s.peek()) {
			s.advance()
		}
	}

	floatValue, _ := strconv.ParseFloat(s.source[s.start:s.current], 64)
	s.addTokenWithLiteral(TOKEN_NUMBER, floatValue)
}

func (s *Scanner) scanIdentifier() {
	for isAlphaNumeric(s.peek()) {
		s.advance()
	}

	text := s.source[s.start:s.current]
	tokenType, ok := keywords[text]
	if ok {
		s.addSimpleToken(tokenType)
	} else {
		s.addSimpleToken(TOKEN_IDENTIFIER)
	}
}

// support: /* */
func (s *Scanner) scanBlockComment() {
	nestingLevel := 1 // 初始嵌套层级
	for nestingLevel > 0 && !s.isEnd() {
		switch {
		case s.peek() == '\n':
			s.line++
			s.advance()
		case s.peek() == '/' && s.peekNext() == '*':
			s.advance()
			s.advance()
			nestingLevel++
		case s.peek() == '*' && s.peekNext() == '/':
			s.advance()
			s.advance()
			nestingLevel--
		default:
			s.advance()
		}
	}

	if nestingLevel > 0 {
		s.error(s.line, "unterminated block comment")
	}
}

func (s *Scanner) error(line int, message string) {
	s.scanErr = util.ErrorMsg(line, message)
	fmt.Printf("scann error at line %d: %s\n", line, message)
}

func isDigit(c byte) bool {
	return c >= '0' && c <= '9'
}

func isAlpha(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || c == '_'
}

func isAlphaNumeric(c byte) bool {
	return isAlpha(c) || isDigit(c)
}
