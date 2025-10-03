package main

import (
	"fmt"
	"os"
)

func Foo() error {
  var err *os.PathError = nil // type = *os.PathError, data = nil
  return err
}

func main() {
  err := Foo()
  fmt.Println(err) // nil, cuz data = nil
  fmt.Println(err == nil) // false, cuz err type is not nil | data and type should be nil
}