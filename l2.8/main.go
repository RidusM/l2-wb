package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/beevik/ntp"
)

const (
	ntpServer = "ntp0.ntp-servers.net"
	exitCodeErr = 1
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(exitCodeErr)
	}
}

func run() error {
	ntpTime, err := ntp.Time(ntpServer)
	if err != nil {
		return fmt.Errorf("failed to get time from NTP server %s: %w", ntpServer, err)
	}

	log.Printf("Current NTP time: %s\n", ntpTime.Format(time.RFC3339))
	log.Printf("Local time:       %s\n", time.Now().Format(time.RFC3339))

	offset := time.Until(ntpTime)
	log.Printf("Time offset:      %v\n", offset.Round(time.Millisecond))

	return nil
}
