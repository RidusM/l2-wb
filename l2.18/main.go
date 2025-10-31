package main

import (
	"flag"
	"fmt"
)

func main() {
	addrFlag := flag.String("address", "time.google.com", "NTP server address")
	flag.Parse()

	addr := *addrFlag
	fmt.Println("begin reading NTP: ", addr)

	result, err := ntp.Time(addr)
	if err != nil{
		fmt.Errorf("something wrong with ntp: %s", err)
	}
	fmt.Println(result)
}