#ifndef EXPR_H
#define EXPR_H

#include "parser/token.h"
#include <iostream>
#include <memory>
#include <string>
#include <variant>

/**
expression     → equality ;
equality       → comparison ( ( "!=" | "==" ) comparison )* ;
comparison     → term ( ( ">" | ">=" | "<" | "<=" ) term )* ;
term           → factor ( ( "-" | "+" ) factor )* ;
factor         → unary ( ( "/" | "*" ) unary )* ;
unary          → ( "!" | "-" ) unary
               | primary ;
primary        → NUMBER | STRING | "true" | "false" | "nil"
               | "(" expression ")" ;
*/

using ExprResult = std::variant<bool, double, std::string, void *>;

class BinaryExpr;
class GroupingExpr;
class LiteralExpr;
class UnaryExpr;

class ExprVisitor {
public:
  virtual ~ExprVisitor() = default;
  virtual ExprResult visit_binary_expr(const BinaryExpr &expr) = 0;
  virtual ExprResult visit_grouping_expr(const GroupingExpr &expr) = 0;
  virtual ExprResult visit_literal_expr(const LiteralExpr &expr) = 0;
  virtual ExprResult visit_unary_expr(const UnaryExpr &expr) = 0;
};

class Expr {
public:
  virtual ~Expr() = default;
  virtual ExprResult accept(class ExprVisitor &visitor) = 0;
};

class BinaryExpr : public Expr {
private:
  std::unique_ptr<Expr> left_;
  std::unique_ptr<Expr> right_;
  Token op_;

public:
  BinaryExpr(std::unique_ptr<Expr> left, std::unique_ptr<Expr> right,
             Token&& op);
  virtual ExprResult accept(class ExprVisitor &visitor);
};

class GroupingExpr : public Expr {
private:
  std::unique_ptr<Expr> expr_;

public:
  GroupingExpr(std::unique_ptr<Expr> expr);
  virtual ExprResult accept(class ExprVisitor &visitor);
};

class LiteralExpr : public Expr {
private:
  ExprResult value_;
  Token token_;

public:
  LiteralExpr(ExprResult&& value, Token&& token);
  virtual ExprResult accept(class ExprVisitor &visitor);
};

class UnaryExpr : public Expr {
private:
  std::unique_ptr<Expr> right_;
  Token op_;

public:
  UnaryExpr(std::unique_ptr<Expr> right, Token&& op);
  virtual ExprResult accept(class ExprVisitor &visitor);
};

#endif