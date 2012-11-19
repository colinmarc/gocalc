package main

import (
  "os"
  "io"
  "bufio"
  "fmt"
)

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
      res, _ := Eval(line)
      fmt.Printf("%d\n", res)
    }
  }
}