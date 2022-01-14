package main

import (
	"os"
	"fmt"
	"io"
	"os/signal"
	"syscall"
	"strings"
	"net"
	"bufio"
	"log"
)

var (
	parserOutput *os.File
	logFile *os.File
	clusterLogging *os.File
)

func init() {

	// Open log file
	logFile, err := os.OpenFile("/tsi/logging-server.log", os.O_APPEND|os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		log.Println(err)
		return
	}

	parserOutput, err = os.OpenFile("/tsi/parser.log", os.O_APPEND|os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		log.Println(err)
		return
	}

	clusterLogging, err = os.OpenFile("/tsi/tsi.log", os.O_APPEND|os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		log.Println(err)
		return
	}

	log.SetFlags(log.Ldate | log.Lmicroseconds | log.LUTC)
	log.SetOutput(logFile)

	cluster = make(map[string]*nodePods)
	nodes = 0
	canaries = make(map[string]*canaryStamps)
	detectors = make(map[string]*detectorStamps)
	attack.active = false
	malices = make(map[string]*maliceStamps)


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

func handleConnection(c net.Conn, toSorter chan []byte){

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
		toSorter <- str
	}
}

/* this function prints logs to stdout in order to be managed by kubernetes logging system
* logs are also written to tsi shared directory in tsi.log
* IMPORTANT! Logs may be out of order in same cases of bursting. 
* Out of order logs do not affect the parsing system above
*/
func printLogs(logs chan []string){
	for {
		msg := <-logs
		print(strings.Join(msg, " "))
		fmt.Fprint(clusterLogging, strings.Join(msg, " "))
	}
}


func main() {

	listener, err := net.Listen("tcp4", "0.0.0.0:4321")
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Listenning on port 4321")

	toSorter := make(chan []byte, 128)
	toPrinter := make(chan []string, 128)
	toAnalyser := make(chan []string, 128)
	go printLogs(toPrinter)
	go analyseLogs(toAnalyser)
	go sortLogs(toSorter, toAnalyser, toPrinter)
	for {
		cli, err := listener.Accept()
		if err != nil {
			log.Fatal("Accept failed:", err.Error())
			break
		}
		log.Printf("Connection open: %s", cli.RemoteAddr())
		go handleConnection(cli, toSorter)
	}
	listener.Close()
}
