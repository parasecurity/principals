package main

import (
	"bufio"
	"flag"
	"fmt"
	"net"
)

var args struct {
	ip   *string
	port *string
}

func init() {
	args.ip = flag.String("ip", "localhost", "The server ip")
	args.port = flag.String("port", "23456", "The server port")
	flag.Parse()
}

func main() {
	//test client that reads data from connection and prints to stdout
	url := *args.ip + ":" + *args.port
	c, err := net.Dial("tcp4", url)
	if err != nil {
		fmt.Println(err)
		return
	}

	for {
		message, err := bufio.NewReader(c).ReadString('\n')
		if err != nil {
			//if the connection is closed, let the client terminate
			fmt.Println(err)
			break
		}
		fmt.Print("->: " + message)
	}
	c.Close()
}
