package main

import "fmt"

func test() (x int) { // 'x' is a named return value, initialized to 0.
  defer func() { // execute before return
    x++ // increment named return value
  }()
  x = 1
  return // return val set x = 1, defer increment x to 2, return val still x, but it's now = 2
}

func anotherTest() int { // returned valued is not named
  var x int // 'x' is a local variable, initialized to 0.
  defer func() { // execute before returns
    x++ // increment local value
  }()
  x = 1
  return x // return value from 18, cuz is copied to be func return val, defer func increment local val and it doesn't affect the copied return val
}

func main() {
  fmt.Println(test()) // 2
  fmt.Println(anotherTest()) // 1
}