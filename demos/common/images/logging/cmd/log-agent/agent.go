package main

import (
	"net"
	"os"
	"os/signal"
	"syscall"
	"io"
	"bufio"
	"fmt"
	"log"
)

func init() {

	// Setup signal catching
	sigs := make(chan os.Signal, 1)
	// Catch all signals since not explicitly listing
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)
	// Method invoked upon seeing signal
	go func() {
		s := <-sigs
		log.Printf("RECEIVED SIGNAL: %s", s)
		os.Exit(1)
	}()
}

func serverWriter(c net.Conn, logs chan []byte){
	defer func() {
		// TODO
		println("Server connection closed")
		c.Close()
	}()

	// writer := bufio.NewWriter(c)

	for {
		msg := <-logs
		fmt.Fprintf(c, "%s %s",c.RemoteAddr().String(), msg)
		// writer.Flush()
	}
}

func handleConnection(c *net.UnixConn, logs chan []byte){

	defer func() {
		c.Close()
		println("Connection closed: %s ", c.RemoteAddr())
	}()

	reader := bufio.NewReader(c)
	serving: for {
		data, err := reader.ReadBytes('\n')

		switch err {
		case nil:
			logs<- data
		case io.EOF:
			println("eof error")
			break serving
		default:
			println("other error")
			// TODO check here!
			break serving
		}
	}
}

func listenUnixAndServe(logs chan []byte) {
	network := "unixpacket"
	path := "/tmp/testlog.sock"
	sock_addr, err := net.ResolveUnixAddr(network, path)

	if err := os.RemoveAll(path); err != nil {
		println(err)
	}

	listener, err := net.ListenUnix(network, sock_addr)
	if err != nil {
		println(err)
	}


	for {
		cli, err := listener.AcceptUnix()
		if err != nil {
			println("Accept failed:", err.Error())
			break
		}
		println("Connection open: %s ", cli.RemoteAddr())
		go handleConnection(cli, logs)
	}
	listener.Close()
}

func main () {

	serverConn, err := net.Dial("tcp4", "localhost:4321")
	if err != nil {
		println(err)
		os.Exit(1)
	}
	logs := make(chan []byte, 128)
	go serverWriter(serverConn, logs)
	listenUnixAndServe(logs)


}
