package main

import (
	"flag"
	"fmt"
	"net"
	"strings"
)

var prefix = flag.String("prefix", "10.", "IP prefix to search for")

func main() {
	flag.Parse()

	addrs, err := net.InterfaceAddrs()
	if err != nil {
		fmt.Println("ERROR")
		return
	}

	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipv4 := ipnet.IP.To4(); ipv4 != nil {
				ipString := ipv4.String()
				if strings.HasPrefix(ipString, *prefix) {
					fmt.Println(ipString)
					return
				}
			}
		}
	}

	fmt.Println("ERROR")
}
