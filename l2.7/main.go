package main

import (
	"fmt"
	"math/rand"
	"time"
)

func asChan(vs ...int) <-chan int {
  c := make(chan int)
  go func() {
    for _, v := range vs {
      c <- v
      // random delay (0 to 999 ms) before sending next value
      time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)
    }
    close(c) // Close chan after sending all values
  }()
  return c
}

func merge(a, b <-chan int) <-chan int {
  c := make(chan int)
  go func() {
    for {
      select {
        // Try read 'a' chan
        case v, ok := <-a:
          if ok {
            c <- v               // If ok — send in out chan
          } else {
            a = nil              // If closed — nil chan,
                                 // select doesn't try to read from it anymore
          }
        // Try read 'b' chan
        case v, ok := <-b:
          if ok {
            c <- v
          } else {
            b = nil
          }
      }
      // If both chan closed (eq nil), finish the work
      if a == nil && b == nil {
        close(c)
        return
      }
    }
  }()
  return c
}

func main() {
  rand.Seed(time.Now().Unix()) // Init gen rand int
  
  a := asChan(1, 3, 5, 7) // 1,3,5,7
  b := asChan(2, 4, 6, 8) // 2,4,6,8
  
  c := merge(a, b)        // merge two chan
  
  for v := range c {      // Read, while chan not closed
    fmt.Print(v)          // Print without spaces and hyphenation
  }
}