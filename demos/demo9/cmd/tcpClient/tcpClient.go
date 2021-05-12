package main

import (
	"bufio"
	"flag"
	"log"
	"math/rand"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var (
	args struct {
		server  *string
		sleep   *int
		timeout *int
		logPath *string
	}
)

func init() {
	args.server = flag.String("conn", "192.168.1.1:12345", "The server connection in format ip:port e.g. 192.168.1.1:12345")
	args.sleep = flag.Int("s", 1000, "sleep in ms e.g. 1000")
	args.timeout = flag.Int("t", 1000, "timeout in ms e.g. 1000")
	args.logPath = flag.String("lp", "./client.log", "The path to the log file")
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

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const maxString = 1536

func genRandPayload(n int64) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Int63()%int64(len(letterBytes))]
	}
	return string(b)
}

func main() {
	conn, err := net.Dial("tcp", *args.server)
	if err != nil {
		log.Println(err)
		return
	}
	defer conn.Close()

	timeoutDuration := time.Duration(*args.timeout) * time.Millisecond
	bufReader := bufio.NewReader(conn)
	var count, wrong uint64
	count = 0
	wrong = 0
	for {
		message := genRandPayload(rand.Int63()%maxString) + "\n"
		// log.Print("->: " + message)
		conn.Write([]byte(message))
		// Set a deadline for reading. Read operation will fail if no data
		// is received after deadline.
		conn.SetReadDeadline(time.Now().Add(timeoutDuration))
		// Read tokens delimited by newline
		bytes, err := bufReader.ReadBytes('\n')
		if err != nil {
			log.Println(err)
			return
		}
		if string(bytes) != message {
			log.Print("->: wrong answer: " + string(bytes))
			wrong++
		}
		count++
		log.Println("Transfers: ", count, ", failed: ", wrong)
		time.Sleep(time.Duration(*args.sleep) * time.Millisecond)
	}

}
