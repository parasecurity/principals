package main

import (
	"os"
	"fmt"
	"flag"
	"io"
	"strconv"
	"time"
	"strings"
	"net"
	"bufio"
	"log"
)

var (
	logFile *os.File
	clusterLogging *os.File
	listener net.Listener
	toSorter chan []byte
	toOut chan []string
	toAnalyser chan []string
	sigs chan os.Signal
	sorter chan struct{}
	attackPodName *string
	attackInitLog *string
	localLogFile *string
	clusterLogFile *string
)

func init_channels() {

	toSorter = make(chan []byte, 128)
	toOut = make(chan []string, 128)
	toAnalyser = make(chan []string, 128)
	// Setup signal catching
	sigs = make(chan os.Signal, 1)
	// Used for syncing during shutdown
	stop_connections = make(chan struct{}, 1)
	// some-kinda mutex for func sortAndSend
	sorter = make(chan struct{}, 1)
}

func init() {

	attackPodName = flag.String("attack-pod-name", "attack", "the prefix of the name of attackers")
	attackInitLog = flag.String("attack-init-log", "Start attacking", "The exact log entry an attacker sends just before start attacking")
	localLogFile = flag.String("server-log-file", "/tsi/logging-server.log", "File where server logs are saved")
	clusterLogFile = flag.String("cluster-log-file", "/tsi/tsi.log", "File where all cluster logs are saved")
	flag.Parse()
	// Open log file
	logFile, err := os.OpenFile(*localLogFile, os.O_APPEND|os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		log.Println(err)
		return
	}
	log.SetFlags(log.Ldate|log.Lmicroseconds|log.LUTC)

	clusterLogging, err = os.OpenFile(*clusterLogFile, os.O_APPEND|os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		log.Println(err)
		return
	}

	log.SetFlags(log.Ldate | log.Lmicroseconds | log.LUTC)
	log.SetOutput(logFile)
}

func handleConnection(c net.Conn, toSorter chan []byte){

	defer func() {
		agents.l.Lock()
		delete(agents.c, c.RemoteAddr().String())
		agents.l.Unlock()
		c.Close()
		log.Printf("Handler Connection closed: %s", c.RemoteAddr())
		connWG.Done()
	}()

	reader := bufio.NewReader(c)
	log.Printf("Serving %s", c.RemoteAddr().String())

	serve: for {
		select {
		case <-stop_connections:
			log.Println("Handler notified to stop")
			break serve
		default:
			// set timeout for reading in case of Signal Interrupt
			c.SetReadDeadline(time.Now().Add(time.Second))
			str, err := reader.ReadBytes('\n')
			if err != nil {
				if err == io.EOF {
					log.Println("EOF")
					noConns.Done()
					break serve;
				} else if !err.(net.Error).Timeout() {
					// TODO die
					log.Println(err)
				}
			} else if strings.Contains(string(str), "agentPing") {
				log.Println("An agent is worried")
			} else {
				log.Printf("Received log: %s", string(str))
				toSorter <- str
			}
		}
	}
}

/* this function prints logs to stdout in order to be managed by kubernetes logging system
* logs are also written to tsi shared directory in tsi.log
* IMPORTANT! Logs may be out of order in same cases of bursting. 
* Out of order logs do not affect the parsing system above
*/
func outputLogs(logs chan []string){
	for {
		msg := <-logs
		timestamp, _ := strconv.ParseInt(msg[3], 10, 64)
		tt := time.Unix(timestamp / 1000000, 1000*(timestamp % 1000000))
		print(tt.UTC().Format(time.StampMicro), " ", strings.Join(msg[:3], " "), " ", strings.Join(msg[4:], " "))
		fmt.Fprint(clusterLogging, tt.UTC().Format(time.StampMicro), " ", strings.Join(msg[:3], " "), " ", strings.Join(msg[4:], " "))
	}
}


func main() {
	var err error
	listener, err = net.Listen("tcp4", "0.0.0.0:4321")
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Listening on port 4321")
	defer func(){
		log.Println("defer closing listnener")
		listener.Close()
	}()

	go outputLogs(toOut)
	go analyseLogs(toAnalyser)
	go sortLogs(toSorter, toAnalyser, toOut)

	listening: for {
		select {
		case <-stop_listener:
			log.Println("Closing listener")
			listener.Close()
			stop_listener<- struct{}{}
			break listening
		default:
			cli, err := listener.Accept()
			if err != nil {
				log.Println("Accept failed:", err.Error())
				break listening
			}
			log.Printf("Connection open: %s", cli.RemoteAddr())
			connWG.Add(1)
			noConns.Add()

			agents.l.Lock()
			agents.c[cli.RemoteAddr().String()] = cli
			go handleConnection(cli, toSorter)
			agents.l.Unlock()
		}
	}
	log.Println("main blocked")
	<-stop_listener
	log.Println("main finished")
}
