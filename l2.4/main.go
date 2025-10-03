package main

func main() {
	ch := make(chan int)
	go func() {
		for i := 0; i < 10; i++ {
			ch <- i
		}
		// close(ch) - add this to avoid deadlock
	}()
	for n := range ch { // work while the channel has data/is open
		println(n) // range 0...9
	} // finish work, but the channel won't close
} // fatal error: all goroutines are asleep - deadlock!