package main

import (
	"bufio"
	"flag"
	"log"
	"net"
	"os"
	"os/signal"
)

var (
	args struct {
		ip      *string
		port    *string
		logPath *string
	}

	sigs *chan os.Signal
)

func init() {
	args.ip = flag.String("ip", "localhost", "The server ip")
	args.port = flag.String("port", "23456", "The server port")
	args.logPath = flag.String("lp", "./out_client.log", "The path to the log file")
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

func main() {
	//test client that reads data from connection and prints to stdout
	url := *args.ip + ":" + *args.port
	c, err := net.Dial("tcp4", url)
	if err != nil {
		log.Println(err)
		return
	}

	for {
		message, err := bufio.NewReader(c).ReadString('\n')
		if err != nil {
			//if the connection is closed, let the client terminate
			log.Println(err)
			break
		}
		log.Print("->: " + message)
	}
	c.Close()
}
