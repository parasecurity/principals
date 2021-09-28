package main

import (
	"os"
	"io"
	"fmt"
	"sync"
	"encoding/binary"
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

	// log.SetFlags(log.Ldate | log.Lmicroseconds | log.LUTC)
	log.SetOutput(logFile)

	// Setup signal catching
	sigs := make(chan os.Signal, 1)
	// Catch all signals since not explicitly listing
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)
	// Method invoked upon seeing signal
	go func() {
		s := <-sigs
		log.Printf("RECEIVED SIGNAL: %s", s)
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
		mx.Lock()
		central_logs.WriteString(str)
		mx.Unlock()
	}
}

func handleMetricsConnection(c net.Conn){
	//vers0.2
	// Recieve metric data from client
	reader := bufio.NewReader(c)
	for {
		//str, err := reader.ReadString('\n')

		b := make([]byte, 8)
		var err error
		b[0], err = reader.ReadByte()
		b[1], err = reader.ReadByte()
		b[2], err = reader.ReadByte()
		b[3], err = reader.ReadByte()
		b[4], err = reader.ReadByte()
		b[5], err = reader.ReadByte()
		b[6], err = reader.ReadByte()
		b[7], err = reader.ReadByte()

		str, err := reader.ReadString('\n')

		if err != nil {
			if err == io.EOF {
				break;
			} else {
				log.Println(err)
			}
		} else {
			//log.Printf(str)
			t := binary.LittleEndian.Uint64(b[0:8])
			msg := fmt.Sprintf("%v %s", t, str)
			log.Printf(msg)
		}
	}
	c.Close()
	log.Printf("Connection closed: %s", c.RemoteAddr())
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
		go handleConnection(cli)
	}
	listener.Close()
}
