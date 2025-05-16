#include "parser/scanner.h"
#include <any>
#include <cctype>
#include <cmath>
#include <iostream>

const std::unordered_map<std::string, TokenType> keywords_ = {
    {"and", TokenType::TOKEN_AND},
    {"class", TokenType::TOKEN_CLASS},
    {"else", TokenType::TOKEN_ELSE},
    {"false", TokenType::TOKEN_FALSE},
    {"for", TokenType::TOKEN_FOR},
    {"fun", TokenType::TOKEN_FUN},
    {"if", TokenType::TOKEN_IF},
    {"nil", TokenType::TOKEN_NIL},
    {"or", TokenType::TOKEN_OR},
    {"print", TokenType::TOKEN_PRINT},
    {"return", TokenType::TOKEN_RETURN},
    {"super", TokenType::TOKEN_SUPER},
    {"this", TokenType::TOKEN_THIS},
    {"true", TokenType::TOKEN_TRUE},
    {"var", TokenType::TOKEN_VAR},
    {"while", TokenType::TOKEN_WHILE},
    {"break", TokenType::TOKEN_BREAK},
    {"continue", TokenType::TOKEN_CONTINUE}};

Scanner::Scanner(const std::string &source) : source_(source) {}

void Scanner::scan_tokens(std::vector<Token> &tokens) {
  while (!is_end()) {
    start_ = current_;
    start_column_ = current_column_;
    start_line_ = current_line_;
    scan_token(tokens);
  }
  tokens.emplace_back(TokenType::TOKEN_EOF, "", 0, start_line_, start_column_);
}

bool Scanner::has_error() const { return has_error_; }

void Scanner::scan_token(std::vector<Token> &tokens) {
  char c = advance();
  switch (c) {
  case '(':
    add_token(tokens, TokenType::TOKEN_LEFT_PAREN);
    break;
  case ')':
    add_token(tokens, TokenType::TOKEN_RIGHT_PAREN);
    break;
  case '{':
    add_token(tokens, TokenType::TOKEN_LEFT_BRACE);
    break;
  case '}':
    add_token(tokens, TokenType::TOKEN_RIGHT_BRACE);
    break;
  case ',':
    add_token(tokens, TokenType::TOKEN_COMMA);
    break;
  case '.':
    add_token(tokens, TokenType::TOKEN_DOT);
    break;
  case '-':
    add_token(tokens, TokenType::TOKEN_MINUS);
    break;
  case '+':
    add_token(tokens, TokenType::TOKEN_PLUS);
    break;
  case ';':
    add_token(tokens, TokenType::TOKEN_SEMICOLON);
    break;
  case '*':
    add_token(tokens, TokenType::TOKEN_STAR);
    break;

  case '!':
    add_conditional_token(tokens, '=', TokenType::TOKEN_BANG_EQUAL,
                          TokenType::TOKEN_BANG);
    break;
  case '=':
    add_conditional_token(tokens, '=', TokenType::TOKEN_EQUAL_EQUAL,
                          TokenType::TOKEN_EQUAL);
    break;
  case '<':
    add_conditional_token(tokens, '=', TokenType::TOKEN_LESS_EQUAL,
                          TokenType::TOKEN_LESS);
    break;
  case '>':
    add_conditional_token(tokens, '=', TokenType::TOKEN_GREATER_EQUAL,
                          TokenType::TOKEN_GREATER);
    break;

  case '/':
    if (match('/')) { // handle single-line comment
      while (peek() != '\n' && !is_end())
        advance();
    } else if (match('*')) { // handle multi-line comment
      scan_block_comment();
    } else {
      add_token(tokens, TokenType::TOKEN_SLASH);
    }
    break;

  case ' ':
  case '\r':
  case '\t':
    break;

  case '\n':
    break;

  case '"':
    scan_string(tokens);
    break;

  default:
    if (isdigit(c)) {
      scan_number(tokens);
    } else if (isalpha(c) || c == '_') {
      scan_identifier(tokens);
    } else {
      error("Unexpected character");
    }
  }
}

bool Scanner::is_end() const { return current_ >= source_.size(); }

bool Scanner::match(char expected) {
  if (is_end())
    return false;
  if (source_[current_] != expected)
    return false;
  current_++;
  return true;
}

char Scanner::advance() {
  char c = source_[current_];
  current_++;
  if (c == '\n') {
    current_line_++;
    current_column_ = 0;
  } else {
    current_column_++;
  }
  return c;
}

char Scanner::peek() const {
  if (is_end())
    return '\0';
  return source_[current_];
}

char Scanner::peek_next() const {
  if (current_ + 1 >= source_.size())
    return '\0';
  return source_[current_ + 1];
}

void Scanner::add_token(std::vector<Token> &tokens, TokenType type) {
  tokens.emplace_back(type, source_.c_str() + start_, current_ - start_,
                      start_line_, start_column_);
}

void Scanner::add_token(std::vector<Token> &tokens, TokenType type,
                        const char *start, int len) {
  tokens.emplace_back(type, start, len, start_line_, start_column_);
}

void Scanner::add_conditional_token(std::vector<Token> &tokens, char expected,
                                    TokenType matched, TokenType unmatched) {
  if (match(expected)) {
    add_token(tokens, matched);
  } else {
    add_token(tokens, unmatched);
  }
}

void Scanner::error(const std::string &message) {
  std::cout << "error at line " << current_line_ << ": " << message
            << std::endl;
  has_error_ = true;
}

void Scanner::scan_string(std::vector<Token> &tokens) {
  while (peek() != '"' && !is_end()) {
    advance();
  }
  if (is_end()) {
    error("Unterminated string");
    return;
  }
  advance(); // consume the closing quote
  add_token(tokens, TokenType::TOKEN_STRING, source_.c_str() + start_ + 1,
            current_ - start_ - 2);
}

void Scanner::scan_number(std::vector<Token> &tokens) {
  while (isdigit(peek()))
    advance();
  if (peek() == '.' && isdigit(peek_next())) {
    advance();
    while (isdigit(peek()))
      advance();
  }
  add_token(tokens, TokenType::TOKEN_NUMBER, source_.c_str() + start_,
            current_ - start_);
}

void Scanner::scan_identifier(std::vector<Token> &tokens) {
  while (isalnum(peek()) || peek() == '_')
    advance();
  std::string text = source_.substr(start_, current_ - start_);
  TokenType type = keywords_.count(text) > 0 ? keywords_.at(text)
                                             : TokenType::TOKEN_IDENTIFIER;
  add_token(tokens, type);
}

void Scanner::scan_block_comment() {
  int depth = 1;
  while (depth > 0 && !is_end()) {
    if (peek() == '/' && peek_next() == '*') {
      advance();
      advance();
      depth++;
    } else if (peek() == '*' && peek_next() == '/') {
      advance();
      advance();
      depth--;
    } else {
      advance();
    }
  }
  if (depth > 0) {
    error("Unterminated block comment");
  }
}