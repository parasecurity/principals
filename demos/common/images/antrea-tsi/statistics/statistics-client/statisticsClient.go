package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type sendData struct {
	NodeName     string `json:"nodename"`
	Primitive    string `json:"primitive"`
	Data         string `json:"data"`
}

type receivedData struct {
	Primitive string  `json:"primitive"`
	Data 	string   `json:"data"`
}

var (
	args struct {
		server      *string
		broadcaster *string
		logPath     *string
		noNtpSync     *bool
	}
	subnet *net.IPNet
	nodeName string
	primitive string 
)

func init() {
	args.server = flag.String("c", "localhost:12345", "The server listening connection in format ip:port")
	args.broadcaster = flag.String("bc", "localhost:23456", "The statistics broadcaster connection that the server will connect to in format ip:port")
	args.logPath = flag.String("lp", "./statisticsServer.log", "The path to the log file")
	args.noNtpSync = flag.Bool("no-ntp", false, "Do ntp sync")
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
	signal.Notify(sigs, syscall.SIGTERM, syscall.SIGINT)
	// method invoked upon seeing signal
	go func() {
		s := <-sigs
		log.Printf("RECEIVED SIGNAL: %s", s)
		os.Exit(1)
	}()

	// get node name
	nodeName = os.Getenv("NODE_NAME")
}

func connectionWriter(c net.Conn, toBroadcaster chan []byte) {
	defer func() {
		log.Printf("Writer Connection closed %s\n", c.RemoteAddr().String())
		c.Close()
	}()

	log.Printf("Serving writer %s\n", c.RemoteAddr().String())
	for {
		message := <-toBroadcaster
		_, err := c.Write(message)
		if err != nil {
			log.Println(err)
		}
	}
}

func connectionReader(c net.Conn, toBroadcaster chan []byte) {
	defer func() {
		log.Printf("Reader Connection closed %s\n", c.RemoteAddr().String())
		c.Close()
	}()

	log.Printf("Serving reader %s\n", c.RemoteAddr().String())
	reader := bufio.NewReader(c)
	for {
		netData, err := reader.ReadBytes('\n')
		if err != nil {
			log.Println(err)
			break
		}

		var rData receivedData
		err = json.Unmarshal(netData, &rData)
		if err != nil {
			log.Println(err)
			return
		}
		
		sData := sendData{
			nodeName,
			rData.Primitive,
			rData.Data,
		}
		jsonMsg, _ := json.Marshal(sData)
		jsonMsg = append(jsonMsg, []byte("\n")...)
		log.Println(string(jsonMsg))
		toBroadcaster <- jsonMsg
	}
}

func main() {
	// port to listen to input connections (flow controllers)
	var err error
	var retries int = 0
	var connBroadcaster net.Conn

	if !(*args.noNtpSync) {
		ntpSync()
	}

	for retries < 10 {
		connBroadcaster, err = net.Dial("tcp4", *args.broadcaster)
		if err == nil {
			break
		}

		log.Println(err)
		retries++
		if retries < 10 {
			log.Println("Retrying")
		} else {
			log.Println("Failed to connect to broadcaster")
		}
		time.Sleep(5 * time.Second)
	}

	toBroadcaster := make(chan []byte)

	go connectionWriter(connBroadcaster, toBroadcaster)

	listener, err := net.Listen("tcp4", *args.server)
	if err != nil {
		log.Println(err)
		return
	}
	defer listener.Close()

	for {
		c, err := listener.Accept()
		if err != nil {
			log.Println(err)
			return
		}
		go connectionReader(c, toBroadcaster)
	}
}
