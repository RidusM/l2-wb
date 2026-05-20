package main

import (
	"fmt"
	"time"
)

func sig(after time.Duration) <-chan any {
    c := make(chan any)
    go func() {
        defer close(c)
        time.Sleep(after)
    }()
    return c
}

func main() {
    start := time.Now()
    <-Or(
        sig(2*time.Hour),
        sig(5*time.Minute),
        sig(1*time.Second),
        sig(1*time.Hour),
        sig(1*time.Minute),
    )
    fmt.Printf("Done after: %v (expected ~1s)\n\n", time.Since(start))
}

func Or(channels ...<-chan any) <-chan any {
    switch len(channels) {
    case 0:
        return nil
    case 1:
        return channels[0]
    }

    orDone := make(chan any)

    go func() {
        defer close(orDone)

        switch len(channels) {
        case 2:
            select {
            case <-channels[0]:
            case <-channels[1]:
            }
        default:
            mid := len(channels) / 2
            select {
            case <-Or(channels[:mid]...):
            case <-Or(channels[mid:]...):
            }
        }
    }()

    return orDone
}