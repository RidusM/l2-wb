package main

import (
	"fmt"
)

func main() {
  var s = []string{"1", "2", "3"} // len = 3, cap = 3, addr = 0x0001
  modifySlice(s) // in this case return slice with addr = 0x0001
  fmt.Println(s) // [3, 2, 3]
}

func modifySlice(i []string) {
  i[0] = "3" // [3, 2, 3] = len = 3, cap = 3, addr = 0x0001
  i = append(i, "4") // [3, 2, 3, 4] len = 4, cap = 6, addr = 0x0002! | New slice, in arr with addr = 0x0001 not enough space
  i[1] = "5" // [3, 5, 3, 4] len = 4, cap = 6, addr = 0x002!
  i = append(i, "6") // [3, 5, 3, 4, 6] len = 5, cap = 6, addr = 0x002!
}