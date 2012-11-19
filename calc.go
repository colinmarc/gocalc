package main

import (
  "os"
  "fmt"
  "io"
  "bufio"
  "strings"
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
