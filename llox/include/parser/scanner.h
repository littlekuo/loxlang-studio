#ifndef SCANNER_H
#define SCANNER_H

#include "parser/token.h"
#include <stdexcept>
#include <string>
#include <unordered_map>
#include <vector>

class Scanner {
public:
  explicit Scanner(const std::string &source);
  void scan_tokens(std::vector<Token> &tokens);
  bool has_error() const;

private:
  void scan_token(std::vector<Token> &tokens);

  bool is_end() const;
  bool match(char expected);
  char advance();
  char peek() const;
  char peek_next() const;
  void add_token(std::vector<Token> &tokens, TokenType type);
  void add_token(std::vector<Token> &tokens, TokenType type, const char *start,
                 int len);
  void add_conditional_token(std::vector<Token> &tokens, char expected,
                             TokenType matched, TokenType unmatched);
  void error(const std::string &message);

  void scan_string(std::vector<Token> &tokens);
  void scan_number(std::vector<Token> &tokens);
  void scan_identifier(std::vector<Token> &tokens);
  void scan_block_comment();

  std::string source_;
  size_t start_{0};
  size_t current_{0};
  size_t start_column_{0};
  size_t current_column_{0};
  size_t start_line_{1};
  size_t current_line_{1};
  bool has_error_{false};
};

#endif // SCANNER_H
