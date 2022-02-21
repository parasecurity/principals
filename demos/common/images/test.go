package main

import (
	"net"
	"time"
	"os"
)

var (
	srv *net.TCPConn
	srvAddr *net.TCPAddr
	err error
)

func init() {
	if srvAddr, err = net.ResolveTCPAddr("tcp4", "localhost:8080"); err != nil {
		println("fail to resolve address")
		println(err)
		os.Exit(1)
	}
}

func main() {

	println("dialing server")
	if srv, err = net.DialTCP("tcp4", nil, srvAddr); err != nil {
		println("error connectiong to server")
		println(err)
		os.Exit(1)
	}

	println("redialing server")
	if srv, err = net.DialTCP("tcp4", nil, srvAddr); err != nil {
		println("error connectiong to server")
		println(err)
		os.Exit(1)
	}

	time.Sleep(10*time.Second)	
	for {}

	println("closing connection")
	if err = srv.Close(); err != nil {
		println("Error terminating connection")
		println(err)
		os.Exit(1)
	}

}
