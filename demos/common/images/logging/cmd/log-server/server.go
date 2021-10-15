package main

import (
	"os"
	"io"
	// "fmt"
	"sync"
	// "encoding/binary"
	"os/signal"
	"syscall"
	"net"
	"bufio"
	"log"
)

var (
	foo *string
	central_logs *os.File
	mx sync.Mutex
	stop chan struct{}
)

func init() {

	// Open log file
	temp, err := os.OpenFile("logs.log", os.O_APPEND|os.O_RDWR|os.O_CREATE, 0644)
	central_logs = temp
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

func handleConnection(c net.Conn){

	defer func() {
		c.Close()
		log.Printf("Connection closed: %s", c.RemoteAddr())
	}()

	reader := bufio.NewReader(c)
	log.Printf("Serving %s", c.RemoteAddr().String())

	for {

		str, err := reader.ReadString('\n')

		if err != nil {
			if err == io.EOF {
				break;
			} else {
				log.Println(err)
			}
		}
		// mx.Lock()
		// central_logs.WriteString(str)
		print(str)
		// mx.Unlock()
	}
}
func ping(c net.Conn) {
	b := []byte{65}
	for {
		select{
		case <-stop:
			log.Println("received stop signal")
			return
		default:
			log.Println("writing ping stop signal")
			c.Write(b)
		}
	}
}

func main() {

	listener, err := net.Listen("tcp4", "0.0.0.0:4321")
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Listenning on port 4321")

	for {
		cli, err := listener.Accept()
		if err != nil {
			log.Fatal("Accept failed:", err.Error())
			break
		}
		log.Printf("Connection open: %s", cli.RemoteAddr())
		// go ping(cli)
		go handleConnection(cli)
	}
	listener.Close()
}
