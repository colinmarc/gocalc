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
  OPERATORS string = "^*/+-"
  EOF rune = -1
)

type LexemeType int
type LexFn func(*Lexer) LexFn

const (
  ErrLexeme LexemeType = iota
  NumberLexeme
  OperatorLexeme
  LeftParenLexeme
  RightParenLexeme
)

type Lexeme struct {
  lexeme_type LexemeType
  value string
}

func (l *Lexeme) String() string {
  return l.value
}

type Lexer struct {
  input []rune
  window struct {
    start int
    end int
  }
  stream chan Lexeme
}

func (l *Lexer) Peek() rune {
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
    r := l.Peek()

    if r == EOF {
      return nil
    } else if r == ' ' {
      l.Skip()
    } else if strings.IndexRune(DIGITS, r) >= 0 {
      return lexNumber
    } else if strings.IndexRune(OPERATORS, r) >= 0 {
      l.Expand()
      l.Emit(OperatorLexeme)
    } else if r == '(' {
      l.Expand()
      l.Emit(LeftParenLexeme)
    } else if r == ')' {
      l.Expand()
      l.Emit(RightParenLexeme)
    } else {
      l.Expand()
      l.Emit(ErrLexeme)
    }
  }

  return nil
}

func lexNumber(l *Lexer) LexFn {
  for {
    r := l.Peek()
    if strings.IndexRune(DIGITS, r) >= 0{
      l.Expand()
    } else {
      l.Emit(NumberLexeme)
      break
    }
  }

  return lexStart
}


func lexLeftParen(l *Lexer) LexFn {
  return lexStart
}

func lexRightParen(l *Lexer) LexFn {
  return lexStart
}

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

//for this function, assume valid input
func groupLexemes(lexemes []Lexeme) Expression {

  if len(lexemes) == 1 {
    val, _ := strconv.Atoi(lexemes[0].value)
    return &ValueExpression{val}
  }

  fmt.Printf("in group! ---\n")
  fmt.Printf("len(lexemes): %d\n", len(lexemes))

  fmt.Printf("got:")
  for i := range lexemes {
    fmt.Printf(" %s", lexemes[i].value)
  }
  fmt.Printf("\n")

  // groups := make([][]Lexeme{}, len(lexemes)/2 + 1)
  // operators := make([]Lexeme{}, len(lexemes)/2)

  // nesting := 0

  // for i := range lexemes {
  //   l := lexemes[i]

  //   switch l.lexeme_type {
  //   case NumberLexeme:
  //     if nesting == 0 {
  //       append(groups, []Lexeme{l})
  //     } else {
  //       continue
  //     }
  //   case LeftParenLexeme:
  //     nesting += 1
  //   case RightParenLexeme:
  //     nesting -= 1
  //     if nesting == 0 {

  //     }
  //   case OperatorLexeme:
  //     oper := Operators[l.value]
  //     if nesting == 0 && oper > highest_oper {
  //       highest_oper = oper
  //       split_at = i
  //     }
  //   }
  // }


  //try to unwrap the outermost parens
  last := len(lexemes) - 1
  if lexemes[0].lexeme_type == LeftParenLexeme {
    nesting := 0
    unwrap := false
    for i := range lexemes {
      l := lexemes[i]

      if l.lexeme_type == LeftParenLexeme {
        nesting += 1
      } else if l.lexeme_type == RightParenLexeme {
        nesting -= 1
        if nesting == 0 {
          if i == last {
            unwrap = true
          }
          break
        }
      }
    }

    if unwrap {
      lexemes = lexemes[1:last]
    }
  }

  split_at := -1
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
      if nesting == 0 && oper > highest_oper {
        highest_oper = oper
        split_at = i
      }
    }
  }

  fmt.Printf("split_at: %d\n", split_at)
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

func eval(line []byte) (int, error) {
  lexer := lex(string(line))
  expr := parse(lexer)

  return expr.Evaluate(), nil
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
    if len(line) > 0 {
      res, _ := eval(line)
      fmt.Printf("%d\n", res)
    }
  }
}
