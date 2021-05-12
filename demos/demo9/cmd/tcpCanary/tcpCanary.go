package main

import (
	"bufio"
	"flag"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var (
	args struct {
		server    *string
		threshold *int
		logPath   *string
	}
)

func timeGet(url string) {
	conn, err := net.Dial("tcp", url)
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	reader := bufio.NewReader(conn)
	for {
		conn.Write([]byte("Test connection\n"))
		start := time.Now()
		_, err = reader.ReadString('\n')

		if err != nil {
			panic(err)
		}
		interval := time.Since(start)
		log.Println("Response in :", interval)
		if interval > time.Duration(*args.threshold)*time.Microsecond {
			log.Println("Threshold passed:", interval)
		}
		time.Sleep(time.Second)
	}

}

func init() {
	args.server = flag.String("conn", "192.168.1.1:12345", "The server connection in format ip:port e.g. 192.168.1.1:12345")
	args.threshold = flag.Int("t", 1000, "The time threshold in μs")
	args.logPath = flag.String("lp", "./canary.log", "The path to the log file")
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

func main() {
	timeGet(*args.server)
}
