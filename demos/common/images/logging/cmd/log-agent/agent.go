package main

import (
	"net"
	"os"
	"os/signal"
	"syscall"
	"io"
	"bufio"
	"fmt"
	"time"
	"log"
)

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
			break serving
		default:
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



type agent struct {
	req chan struct{}
	ans chan struct{}
	srv *net.TCPConn
	srvAddr *net.TCPAddr
}

var state agent

func (c * agent) fixit() {
	c.req<- struct{}{}
	<-c.ans
}

func (c *agent) stateCheck() {
	for {
		<-c.req
		n, err := fmt.Fprintf(c.srv, "%s ping", c.srv.LocalAddr().String())
		n = n
		if err != nil {
			state.srv.Close()
			err = connectToServer(0)
		}

		c.ans<- struct{}{}
	}
}

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

func (c *agent) serverWriter(logs chan []byte){
	defer func() {
		// TODO
		println("Server connection closed")
		c.srv.Close()
	}()

	// writer := bufio.NewWriter(c)

	for {
		msg := <-logs
		// TODO
		for {
			_, err := fmt.Fprintf(c.srv, "%s %s",c.srv.LocalAddr().String(), msg)
			if err != nil {
				state.fixit()
			} else {
				break
			}
		}
		// writer.Flush()
	}
}

// for negative or 0 retries loops forever
func connectToServer(retries int) (err error){
	for i := retries; i != 1; i-- {
		state.srv, err = net.DialTCP("tcp4", nil, state.srvAddr)
		if err == nil {
			break
		}
		// TODO log the error maybe
		time.Sleep(1 * time.Second)
	}
	return
}

func main () {

	var err error
	state.srvAddr, err = net.ResolveTCPAddr("tcp4", "localhost:4321")
	err = connectToServer(10)
	if err != nil {
		println("failed to connect to server")
		println(err)
		os.Exit(1)
	} else {
		println("connection to server established")
	}

	logs := make(chan []byte, 128)
	state.req = make(chan struct{})
	state.ans = make(chan struct{})
	go state.stateCheck()
	go state.serverWriter(logs)
	listenUnixAndServe(logs)


}
