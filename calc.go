package main

import (
  "math"
  "strconv"
)

type Operator int
const (
  Power Operator = iota
  Product
  Division
  Addition
  Subtraction
)

var Operators = map[string]Operator {
  "^": Power,
  "*": Product,
  "/": Division,
  "+": Addition,
  "-": Subtraction,
}

type Expression interface {
  Evaluate() int
}

type ValueExpression struct {
  value int
}
func (value *ValueExpression) Evaluate() int {
  return int(value.value)
}

type NestedExpression struct {
  left Expression
  right Expression
}

type PowerExpression NestedExpression
func (expr *PowerExpression) Evaluate() int {
  left := float64(expr.left.Evaluate())
  right := float64(expr.right.Evaluate())
  return int(math.Pow(left, right))
}

type ProductExpression NestedExpression
func (expr *ProductExpression) Evaluate() int {
  return expr.left.Evaluate() * expr.right.Evaluate()
}

type DivisionExpression NestedExpression
func (expr *DivisionExpression) Evaluate() int {
  return expr.left.Evaluate() / expr.right.Evaluate()
}

type AdditionExpression NestedExpression
func (expr *AdditionExpression) Evaluate() int {
  return expr.left.Evaluate() + expr.right.Evaluate()
}

type SubtractionExpression NestedExpression
func (expr *SubtractionExpression) Evaluate() int {
  return expr.left.Evaluate() - expr.right.Evaluate()
}

func newExpression(oper Operator, left Expression, right Expression) Expression {
  var expr Expression
  switch oper {
  case Power:
    expr = &PowerExpression{left, right}
  case Product:
    expr = &ProductExpression{left, right}
  case Division:
    expr = &DivisionExpression{left, right}
  case Addition:
    expr = &AdditionExpression{left, right}
  case Subtraction:
    expr = &SubtractionExpression{left, right}
  }

  return expr
}

func unwrap_parens(lexemes []Lexeme) []Lexeme {
  for lexemes[0].lexeme_type == LeftParenLexeme {
    last := len(lexemes) - 1
    nesting := 0
    var i int
    for i = range lexemes {
      l := lexemes[i]
      if l.lexeme_type == LeftParenLexeme {
        nesting += 1
      } else if l.lexeme_type == RightParenLexeme {
        nesting -= 1
        if nesting == 0 {
          break
        }
      }
    }

    if i == last {
      lexemes = lexemes[1:last]
    } else {
      break
    }
  }

  return lexemes
}

//for this function, assume valid input
func groupLexemes(lexemes []Lexeme) Expression {

  if len(lexemes) == 1 {
    val, _ := strconv.Atoi(lexemes[0].value)
    return &ValueExpression{val}
  }

  lexemes = unwrap_parens(lexemes)

  split_at := -1
  lowest_nesting := -1
  highest_oper := Operator(-1)
  nesting := 0

  for i := range lexemes {
    l := lexemes[i]

    switch l.lexeme_type {
    case NumberLexeme:
      continue
    case LeftParenLexeme:
      nesting += 1
    case RightParenLexeme:
      nesting -= 1
    case OperatorLexeme:
      oper := Operators[l.value]
      if split_at == -1 ||
         nesting < lowest_nesting ||
         (nesting == lowest_nesting && oper > highest_oper) {
        highest_oper = oper
        split_at = i
        lowest_nesting = nesting
      }
    }
  }

  left := groupLexemes(lexemes[:split_at])
  right := groupLexemes(lexemes[split_at+1:])
  return newExpression(highest_oper, left, right)
}

func parse(lexer *Lexer) Expression {
  //TODO: validate?
  lexemes := []Lexeme{}
  for l := range lexer.stream {
    lexemes = append(lexemes, l)
  }

  return groupLexemes(lexemes)
}

func Eval(line []byte) (int, error) {
  lexer := Lex(string(line))
  expr := parse(lexer)

  return expr.Evaluate(), nil
}
