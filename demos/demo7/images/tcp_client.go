package main

import (
	"bufio"
	"flag"
	"fmt"
	"net"
	"os"
	"strings"
)

var args struct {
	ip   *string
	port *string
}

func init() {
	args.ip = flag.String("ip", "localhost", "The server ip")
	args.port = flag.String("port", "12345", "The server port")
	flag.Parse()
}

func main() {
	url := *args.ip + ":" + *args.port
	c, err := net.Dial("tcp4", url)
	if err != nil {
		fmt.Println(err)
		return
	}

	go func() {
		for {
			message, err := bufio.NewReader(c).ReadString('\n')
			if err != nil {
				fmt.Println(err)
				break
			}
			fmt.Print("->: " + message)
		}
	}()

	for {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print(">> ")
		text, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error, exiting")
			c.Close()
			c = nil
			break
		} else {
			if strings.TrimSpace(string(text)) == "STOP" {
				fmt.Println("TCP client exiting...")
				return
			}
			fmt.Fprintf(c, text+"\n")
		}
	}
}
