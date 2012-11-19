package main

import (
  "os"
  "fmt"
  "io"
  "bufio"
  "strings"
  "math"
  "strconv"
//  "unicode/utf8"
)

const (
  DIGITS string = "0123456789"
  EOF rune = -1
)

type LexemeType int
type LexFn func(*Lexer) LexFn

const (
  ErrLexeme LexemeType = iota
  NumberLexeme
  OperatorLexeme
)

type Lexeme struct {
  lexeme_type LexemeType
  value string
}

type Lexer struct {
  input []rune
  window struct {
    start int
    end int
  }
  stream chan Lexeme
}

func (l *Lexer) Next() rune {
  if l.window.end >= len(l.input) {
    return EOF
  }
  return l.input[l.window.end]
}

func (l *Lexer) Expand() {
  l.window.end += 1
}

func (l *Lexer) Skip() {
  l.Expand()
  l.Discard()
}

func (l *Lexer) Discard() {
  l.window.start = l.window.end
}

func (l *Lexer) Current() []rune {
  return l.input[l.window.start:l.window.end]
}

func (l *Lexer) Emit(lexeme_type LexemeType) {
  chunk := l.Current()
  l.stream <- Lexeme{lexeme_type, string(chunk)}
  l.Discard()
}

func lexStart(l *Lexer) (next_fn LexFn) {
  for {
    r := l.Next()

    if r == EOF {
      return nil
    } else if r == ' ' {
      l.Skip()
    } else if strings.IndexRune(DIGITS, r) >= 0 {
      return lexNumber
    } else {
      return lexOperator
    }
  }

  return nil
}

func lexNumber(l *Lexer) LexFn {
  for {
    r := l.Next()
    if strings.IndexRune(DIGITS, r) >= 0{
      l.Expand()
    } else {
      l.Emit(NumberLexeme)
      break
    }
  }

  return lexStart
}

func lexOperator(l *Lexer) LexFn {
  l.Expand()
  l.Emit(OperatorLexeme)
  return lexStart
}

// func lexLeftParen(l *Lexer) LexFn {
//   return lexStart
// }

// func lexRightParen(l *Lexer) LexFn {
//   return lexStart
// }

func (l *Lexer) Run() {
  for state := lexStart; state != nil; {
    state = state(l)
  }
  close(l.stream)
}

func lex(input string) *Lexer {
  l := &Lexer {input: []rune(input), stream: make(chan Lexeme)}

  go l.Run()
  return l
}

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

func groupLexemes(lexemes []Lexeme) Expression {
  if len(lexemes) == 1 {
    val, _ := strconv.Atoi(lexemes[0].value)
    return &ValueExpression{val}
  }

  //janky
  oper_idx := 1
  lowest_oper := Operators[lexemes[1].value]
  for i := 3; i+1 < len(lexemes); i += 2 {
    oper := Operators[lexemes[i].value]
    if oper < lowest_oper {
      oper_idx = i
    }
  }

  left := groupLexemes(lexemes[:oper_idx-1])
  right := groupLexemes(lexemes[oper_idx+1:])
  return newExpression(lowest_oper, left, right)
}

func parse(lexer *Lexer) {

}

func eval(line []byte) (int, error) {
  lexer := lex(string(line))
  for lexeme := range lexer.stream {
    fmt.Printf("got: %s (%i)\n", lexeme.value, lexeme.lexeme_type)
  }
  return 0, nil
}

func main() {
  reader := bufio.NewReader(os.Stdin)

  for {
    fmt.Printf(">> ")
    raw_line, err := reader.ReadBytes('\n')

    if err != nil {
      if err == io.EOF {
        fmt.Printf("quitting...\n")
        break
      } else {
        fmt.Printf("err: %s", err)
      }
      break
    }

    line := raw_line[:len(raw_line)-1]
    if len(line) > 1 {
      eval(line)
    }
  }
}
