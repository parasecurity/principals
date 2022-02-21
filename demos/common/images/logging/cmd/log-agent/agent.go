package main

import (
	"net"
	"flag"
	"os"
	"os/signal"
	"syscall"
	"io"
	"bufio"
	"log"
)

var (
	args struct {
		logServer *string
		nodeName *string
		sockPath *string
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
			// if len(data) != 0 {
				// logs<- data
			// }
			break serving
		default:
			// TODO check here!
			break serving
		}
	}
}



func init() {

	args.logServer = flag.String("logserveraddr", "localhost:4321", "The logging server listening connection in format ip:port")
	args.nodeName = flag.String("nodename", "localhost", "The node hostname")
	args.sockPath = flag.String("sockpath", "/tmp/testlog.sock", "path to agent's socket")
	flag.Parse()

	// Setup signal catching
	sigs := make(chan os.Signal, 1)
	// Catch all signals since not explicitly listing
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)
	// Method invoked upon seeing signal
	go func() {
		s := <-sigs
		log.Println("RECEIVED SIGNAL: ", s)
		log.Println("Removing socket")
		if err := os.RemoveAll(*args.sockPath); err != nil {
			println(err)
		}
		log.Println("Exiting")
		os.Exit(1)
	}()
}

func listenUnixAndServe(logs chan []byte) {
	network := "unixpacket"
	path := *args.sockPath
	sock_addr, err := net.ResolveUnixAddr(network, path)

	if err := os.RemoveAll(path); err != nil {
		println(err)
	}

	listener, err := net.ListenUnix(network, sock_addr)
	if err != nil {
		println(err)
	}

	defer func(l *net.UnixListener) {
		if err := listener.Close(); err != nil {
			log.Println(err)
		}
		if err := os.RemoveAll(*args.sockPath); err != nil {
			log.Println(err)
		}
	}(listener)

	for {
		cli, err := listener.AcceptUnix()
		if err != nil {
			log.Println("Accept failed:", err.Error())
			break
		}
		log.Println("Connection open:  ", cli.RemoteAddr())
		go handleConnection(cli, logs)
	}
}



func main () {

	logs := make(chan []byte, 128)
	go serverWriter(logs)
	listenUnixAndServe(logs)
}
