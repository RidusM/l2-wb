package main

import "fmt"

func main() {
  a := [5]int{76, 77, 78, 79, 80} // create arr with 5 elements
  var b []int = a[1:4] // slice b , refers the a address in memory, range from 1 element (inclusive) to 4 elements (exclusive)
  fmt.Println(b) // [77 78 79]
}