package main

import (
	"bufio"
	"flag"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"encoding/json"
)

type connections struct {
	c map[int]net.Conn
	data map[int]string
	i int
	l sync.RWMutex
}

var (
	args struct {
		statisticsAddress    *string
		APIstatisticsAddress *string
		logPath        *string
	}
)

type receivedData struct {
	NodeName     string `json:"nodename"`
	Primitive    string `json:"primitive"`
	Data         string `json:"data"`
}

func init() {
	args.statisticsAddress = flag.String("c", "localhost:30000", "The server listening connection in format ip:port")
	args.APIstatisticsAddress = flag.String("ac", "localhost:30001", "The api listening connection in format ip:port")
	args.logPath = flag.String("lp", "./statisticsServer.log", "The path to the log file")
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
		netData, err := reader.ReadBytes('\n')
		if err != nil {
			log.Println(err)
			break
		}
		log.Println("from API ", c.RemoteAddr(), ": ", string(netData))

		// whenever the API server sends data we reply with all the statistics
		// connList.l.RLock()
		// for idx, conn := range connList.c {
		// 	_, err := conn.Write([]byte(netData))
		// 	if err != nil {
		// 		connList.l.RUnlock()
		// 		connList.l.Lock()
		// 		connList.l.Unlock()
		// 		connList.l.RLock()
		// 	}
		// }
		// connList.l.RUnlock()
	}
	c.Close()
	log.Println("API Connection closed ", c.RemoteAddr())
	// if the API server connection is closed we let the handler terminate
}

func handleConnection(c net.Conn, c_idx int, connList *connections) {
	log.Printf("Serving %s, idx %d\n", c.RemoteAddr(), c_idx)
	reader := bufio.NewReader(c)
	for {
		netData, err := reader.ReadBytes('\n')
		if err != nil {
			log.Println(err)
			break
		}

		var cmd receivedData
		err = json.Unmarshal(netData, &cmd)
		if err != nil {
			log.Println(err)
			return
		}
		log.Printf("The received data are" + cmd.NodeName + cmd.Primitive + cmd.Data)
		// connList.l.Lock()

		// connList.l.Unlock()

	}
	c.Close()
	log.Println("Connection closed ", c.RemoteAddr())
	//closeOutConn(c_idx, connList.c[c_idx], connList)
}

func main() {
	listener, err := net.Listen("tcp4", *args.statisticsAddress)
	if err != nil {
		log.Println(err)
		return
	}
	defer listener.Close()

	// map of all flow server connections
	connList := new(connections)
	connList.c = make(map[int]net.Conn)
	connList.i = 0

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
	APIlistener, err := net.Listen("tcp4", *args.APIstatisticsAddress)
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
