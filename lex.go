package main

import "strings"

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

func (l *Lexer) Run() {
  for state := lexStart; state != nil; {
    state = state(l)
  }
  close(l.stream)
}

func Lex(input string) *Lexer {
  l := &Lexer {input: []rune(input), stream: make(chan Lexeme)}

  go l.Run()
  return l
}