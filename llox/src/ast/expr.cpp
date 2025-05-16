#include "ast/expr.h"

BinaryExpr::BinaryExpr(std::unique_ptr<Expr> left, std::unique_ptr<Expr> right,
                       Token&& op)
    : left_(std::move(left)), right_(std::move(right)), op_(std::move(op)) {}

ExprResult BinaryExpr::accept(ExprVisitor &visitor) {
  return visitor.visit_binary_expr(*this);
}

GroupingExpr::GroupingExpr(std::unique_ptr<Expr> expr)
    : expr_(std::move(expr)) {}

ExprResult GroupingExpr::accept(ExprVisitor &visitor) {
  return visitor.visit_grouping_expr(*this);
}

LiteralExpr::LiteralExpr(ExprResult&& value, Token&& token)
    : value_(std::move(value)), token_(std::move(token)) {}

ExprResult LiteralExpr::accept(ExprVisitor &visitor) {
  return visitor.visit_literal_expr(*this);
}

UnaryExpr::UnaryExpr(std::unique_ptr<Expr> right, Token&& op)
    : right_(std::move(right)), op_(std::move(op)) {}
ExprResult UnaryExpr::accept(ExprVisitor &visitor) {
  return visitor.visit_unary_expr(*this);
}