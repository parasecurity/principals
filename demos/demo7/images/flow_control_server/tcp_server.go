package main

import (
	"bufio"
	"bytes"
	"flag"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
)

type connections struct {
	c []net.Conn
	l sync.RWMutex
}

var (
	args struct {
		in_port  *string
		out_port *string
		logPath  *string
	}
	sigs *chan os.Signal
)

func init() {
	args.in_port = flag.String("in_port", "12345", "The server port")
	args.out_port = flag.String("out_port", "23456", "The server port")
	args.logPath = flag.String("lp", "./server.log", "The path to the log file")
	flag.Parse()

	// open log file
	logFile, err := os.OpenFile(*args.logPath, os.O_APPEND|os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		log.Println(err)
		return
	}

	log.SetFlags(log.Ldate | log.Ltime)
	log.SetOutput(logFile)

	// setup signal catching
	sigs := make(chan os.Signal, 1)
	// catch all signals since not explicitly listing
	signal.Notify(sigs)
	// method invoked upon seeing signal
	go func() {
		s := <-sigs
		log.Printf("RECEIVED SIGNAL: %s", s)
		os.Exit(1)
	}()
}

func handleConnection(c net.Conn, connList *connections) {
	log.Printf("Serving sender %s\n", c.RemoteAddr().String())
	reader := bufio.NewReader(c)
	for {
		netData := make([]byte, 4096)
		_, err := reader.Read(netData)
		if err != nil {
			log.Println(err)
			break
		}
		netData = bytes.Trim(netData, "\x00")
		log.Print("from " + c.RemoteAddr().String() + ": " + string(netData))

		// whenever a flow controller sends data we forward the data to all agent servers
		connList.l.RLock()
		for idx, conn := range connList.c {
			_, err := conn.Write([]byte(netData))
			if err != nil {
				// if an agent server connection is closed we remove it from the list
				connList.l.RUnlock()
				connList.l.Lock()
				closeOutConn(idx, conn, connList)
				connList.l.Unlock()
				connList.l.RLock()
			}
		}
		connList.l.RUnlock()
	}
	c.Close()
	log.Printf("Connection closed %s\n", c.RemoteAddr().String())
	// if a flow controller connection is closed we let the handler terminate
}

func closeOutConn(idx int, c net.Conn, connList *connections) {
	c.Close()
	log.Printf("Removing %s from list\n", c.RemoteAddr().String())
	connList.c[idx] = connList.c[len(connList.c)-1]
	connList.c = connList.c[:len(connList.c)-1]
}

func main() {
	// port to listen to input connections (flow controllers)
	in_url := ":" + *args.in_port
	in_listener, err := net.Listen("tcp4", in_url)
	if err != nil {
		log.Println(err)
		return
	}
	defer in_listener.Close()

	// port to listen to output connections (agent servers)
	out_url := ":" + *args.out_port
	out_listener, err := net.Listen("tcp4", out_url)
	if err != nil {
		log.Println(err)
		return
	}
	defer out_listener.Close()

	// list of all agent server connections
	connList := new(connections)

	// whenever a flow controller connects we open a new handler
	go func() {
		for {
			c, err := in_listener.Accept()
			if err != nil {
				log.Println(err)
				return
			}
			go handleConnection(c, connList)
		}
	}()

	// whenever an agent server connects we add the connection to the list
	for {
		c, err := out_listener.Accept()
		if err != nil {
			log.Println(err)
			return
		}
		connList.l.Lock()
		connList.c = append(connList.c, c)
		connList.l.Unlock()
		log.Printf("Serving receiver %s\n", c.RemoteAddr().String())
	}
}
