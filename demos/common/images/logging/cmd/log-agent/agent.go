package main

import (
	"net"
	"flag"
	"os"
	"os/signal"
	"syscall"
	"io"
	"bufio"
	"fmt"
	"time"
	"log"
)

var (

	control struct {
		req chan struct{}
		ans chan struct{}
	}

	srv *net.TCPConn
	srvAddr *net.TCPAddr

	args struct {
		logServer *string
	}

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




func fixit() {
	control.req<- struct{}{}
	<-control.ans
}

func stateCheck() {
	for {
		<-control.req
		n, err := fmt.Fprintf(srv, "%s ping", srv.LocalAddr().String())
		n = n
		if err != nil {
			srv.Close()
			err = connectToServer(0)
		}

		control.ans<- struct{}{}
	}
}

func init() {

	args.logServer = flag.String("logserveraddr", "localhost:4321", "The logging server listening connection in format ip:port")
	flag.Parse()

	control.req = make(chan struct{})
	control.ans = make(chan struct{})

	var err error
	srvAddr, err = net.ResolveTCPAddr("tcp4", *args.logServer)
	err = connectToServer(10)
	if err != nil {
		println("failed to connect to server")
		println(err)
		os.Exit(1)
	} else {
		println("connection to server established")
	}

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

func serverWriter(logs chan []byte){
	defer func() {
		// TODO
		println("Server connection closed")
		srv.Close()
	}()

	// writer := bufio.NewWriter(c)

	for {
		msg := <-logs
		// TODO
		for {
			_, err := fmt.Fprintf(srv, "%s %s", srv.LocalAddr().String(), msg)
			if err != nil {
				fixit()
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
		srv, err = net.DialTCP("tcp4", nil, srvAddr)
		if err == nil {
			break
		}
		// TODO log the error maybe
		time.Sleep(1 * time.Second)
	}
	return
}

func main () {

	logs := make(chan []byte, 128)
	go stateCheck()
	go serverWriter(logs)
	listenUnixAndServe(logs)
}
