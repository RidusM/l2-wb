package main

import (
	"fmt"
	"sync"
	"time"
)

func sig(after time.Duration) <-chan interface{} {
	c := make(chan interface{})
	go func() {
		defer close(c)
		time.Sleep(after)
	}()
	return c
}

func main() {
	start := time.Now()
	<-OrFanIn(
		sig(2*time.Hour),
		sig(5*time.Minute),
		sig(1*time.Second),
		sig(1*time.Hour),
		sig(1*time.Minute),
	)
	fmt.Printf("Done after: %v (expected ~1s)\n\n", time.Since(start))
}

func OrFanIn(channels ...<-chan interface{}) <-chan interface{} {
	switch len(channels) {
	case 0:
		ch := make(chan interface{})
		close(ch)
		return ch
	case 1:
		return channels[0]
	}

	orDone := make(chan interface{})

	go func() {
		defer close(orDone)

		var wg sync.WaitGroup
		wg.Add(len(channels))

		signal := make(chan struct{}, 1)

		for _, ch := range channels {
			go func(c <-chan interface{}) {
				defer wg.Done()
				select {
					case <-c:
						select {
							case signal <- struct{}{}:
							default:
						}
					case <-signal:
				}
			}(ch)
		}

		<-signal

		close(signal)

		wg.Wait()
	}()

	return orDone
}