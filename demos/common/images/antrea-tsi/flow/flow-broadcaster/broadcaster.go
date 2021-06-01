package main

import (
	"bufio"
	"flag"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
)

type connections struct {
	c map[int]net.Conn
	i int
	l sync.RWMutex
}

var (
	args struct {
		broadcaster    *string
		APIbroadcaster *string
		logPath        *string
	}
)

func init() {
	args.broadcaster = flag.String("c", "localhost:30000", "The server listening connection in format ip:port")
	args.APIbroadcaster = flag.String("ac", "localhost:30001", "The api listening connection in format ip:port")
	args.logPath = flag.String("lp", "./broadcaster.log", "The path to the log file")
	flag.Parse()

	// open log file
	logFile, err := os.OpenFile(*args.logPath, os.O_APPEND|os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		log.Println(err)
		return
	}

	log.SetFlags(log.Ldate | log.Lmicroseconds | log.LUTC)
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

func handleAPIConnection(c net.Conn, connList *connections) {
	log.Printf("Serving API %s\n", c.RemoteAddr())
	reader := bufio.NewReader(c)
	for {
		netData := make([]byte, 4096)
		n, err := reader.Read(netData)
		if err != nil {
			log.Println(err)
			break
		}
		log.Println("from API ", c.RemoteAddr(), ": ", string(netData[:n]))

		// whenever the API server sends data we forward the data to all flow servers
		connList.l.RLock()
		for idx, conn := range connList.c {
			_, err := conn.Write([]byte(netData[:n]))
			if err != nil {
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
	log.Println("API Connection closed ", c.RemoteAddr())
	// if the API server connection is closed we let the handler terminate
}

func handleConnection(c net.Conn, c_idx int, connList *connections) {
	log.Printf("Serving %s, idx %d\n", c.RemoteAddr(), c_idx)
	reader := bufio.NewReader(c)
	for {
		netData := make([]byte, 4096)
		n, err := reader.Read(netData)
		if err != nil {
			log.Println(err)
			break
		}
		log.Println("from ", c.RemoteAddr(), " (", c_idx, ")", ": ", string(netData[:n]))

		// whenever a flow server sends data we forward the data to all other flow servers
		connList.l.RLock()
		for idx, conn := range connList.c {
			if idx != c_idx {
				_, err := conn.Write([]byte(netData[:n]))
				if err != nil {
					connList.l.RUnlock()
					connList.l.Lock()
					closeOutConn(idx, conn, connList)
					connList.l.Unlock()
					connList.l.RLock()
				}
			}
		}
		connList.l.RUnlock()
	}
	c.Close()
	log.Println("Connection closed ", c.RemoteAddr())
	closeOutConn(c_idx, connList.c[c_idx], connList)
	// if a flow server connection is closed we let the handler terminate
}

func closeOutConn(idx int, c net.Conn, connList *connections) {
	log.Println("Closing and Removing ", c.RemoteAddr(), " from map")
	c.Close()
	delete(connList.c, idx)
}

func main() {
	// port to listen to input connections (flow servers)
	listener, err := net.Listen("tcp4", *args.broadcaster)
	if err != nil {
		log.Println(err)
		return
	}
	defer listener.Close()

	// map of all flow server connections
	connList := new(connections)
	connList.c = make(map[int]net.Conn)
	connList.i = 0

	// whenever a flow server connects we open a new handler
	go func() {
		for {
			c, err := listener.Accept()
			if err != nil {
				log.Println(err)
				return
			}

			connList.l.Lock()
			connList.c[connList.i] = c
			go handleConnection(c, connList.i, connList)
			connList.i++
			connList.l.Unlock()
		}
	}()

	// port to listen to input connections (API server)
	APIlistener, err := net.Listen("tcp4", *args.APIbroadcaster)
	if err != nil {
		log.Println(err)
		return
	}
	defer APIlistener.Close()

	// whenever an API server connects we open a new handler
	for {
		c, err := APIlistener.Accept()
		if err != nil {
			log.Println(err)
			return
		}

		go handleAPIConnection(c, connList)
	}
}
