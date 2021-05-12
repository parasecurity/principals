package main

import (
	"bufio"
	"flag"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
)

var (
	args struct {
		in_port *string
		logPath *string
	}
)

func init() {
	args.in_port = flag.String("in_port", "12345", "The server port")
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
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)
	// method invoked upon seeing signal
	go func() {
		s := <-sigs

		log.Printf("RECEIVED SIGNAL: %s", s)
		os.Exit(1)
	}()
}

func handleConnection(c net.Conn) {
	log.Printf("Serving sender %s\n", c.RemoteAddr().String())
	reader := bufio.NewReader(c)
	for {
		netData, err := reader.ReadString('\n')
		if err != nil {
			log.Println(err)
			break
		}
		log.Println("from ", c.RemoteAddr().String(), ": ", len(netData), " Bytes")

		// whenever a flow controller sends data we forward the data to all agent servers
		_, err = c.Write([]byte(netData))
		if err != nil {
			log.Print(err)
		}
	}
	c.Close()
	log.Printf("Connection closed %s\n", c.RemoteAddr().String())
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

	for {
		c, err := in_listener.Accept()
		if err != nil {
			log.Println(err)
			return
		}
		go handleConnection(c)
	}

}
