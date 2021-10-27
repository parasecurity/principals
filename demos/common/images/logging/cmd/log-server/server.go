package main

import (
	"os"
	"io"
	"os/signal"
	"syscall"
	"net"
	"bufio"
	"log"
)

func init() {

	// Open log file
	logFile, err := os.OpenFile("local.log", os.O_APPEND|os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		log.Println(err)
		return
	}

	log.SetFlags(log.Ldate | log.Lmicroseconds | log.LUTC)
	log.SetOutput(logFile)

	stop = make(chan struct{})
	// Setup signal catching
	sigs := make(chan os.Signal, 1)
	// Catch all signals since not explicitly listing
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)
	// Method invoked upon seeing signal
	go func() {
		s := <-sigs
		log.Printf("RECEIVED SIGNAL: %s", s)
		// stop<- struct{}{}
		os.Exit(1)
	}()
}

func handleConnection(c net.Conn, toPrinter chan []byte, toAnalyser chan []byte){

	defer func() {
		c.Close()
		log.Printf("Connection closed: %s", c.RemoteAddr())
	}()

	reader := bufio.NewReader(c)
	log.Printf("Serving %s", c.RemoteAddr().String())

	for {

		str, err := reader.ReadBytes('\n')

		if err != nil {
			if err == io.EOF {
				break;
			} else {
				log.Println(err)
			}
		}
		// TODO select for efficiency
		select {
		case toPrinter <- str:
			toAnalyser <- str
		case toAnalyser <- str:
			toPrinter <- str
		}
	}
}

func analyseLogs(logs chan []byte){
		_ = <-logs
}

func printLogs(logs chan []byte){
		msg := <-logs
		print(msg)
}

func main() {

	listener, err := net.Listen("tcp4", "0.0.0.0:4321")
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Listenning on port 4321")

	toPrinter := make(chan []byte, 128)
	toAnalyser := make(chan []byte, 128)
	go printLogs(toPrinter)
	go analyseLogs(toAnalyser)
	for {
		cli, err := listener.Accept()
		if err != nil {
			log.Fatal("Accept failed:", err.Error())
			break
		}
		log.Printf("Connection open: %s", cli.RemoteAddr())
		go handleConnection(cli, toPrinter, toAnalyser)
	}
	listener.Close()
}
