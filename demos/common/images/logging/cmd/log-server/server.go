package main

import (
	"os"
	"fmt"
	"io"
	"os/signal"
	"syscall"
	"strings"
	"strconv"
	"net"
	"bufio"
	"log"
)

var parserOutput *os.File

func init() {

	// Open log file
	logFile, err := os.OpenFile("local.log", os.O_APPEND|os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		log.Println(err)
		return
	}

	parserOutput, err = os.OpenFile("/tsi/parser.log", os.O_APPEND|os.O_RDWR|os.O_CREATE, 0644)

	log.SetFlags(log.Ldate | log.Lmicroseconds | log.LUTC)
	log.SetOutput(logFile)

	//stop = make(chan struct{})
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

/* parsing conventions
*    STEPS: smallest measurable action/reaction
*    EVENTS: consist of STEPS
*/

type dDos struct {
	started bool
	cleanStart bool
	startTime int64
	downTime int64
	blockedConnections int
	reconnections int
	st stats
}

type stats struct {
	timeUntilFirstBlock int64
	timeUntilLastBlock int64
	timeUntilAllDetectorsEnabled int64
	timeUntilFirstDetectorsEnabled int64
	timeUntilFirstTimeout int64
	timeUntilResponsive int64
	timeUntilFullyResponsive int64
}

func (s stats) printStats() {
	// TODO fix float printing to something more human readable
	fmt.Fprintln(parserOutput, "=================================================================")
	fmt.Fprintln(parserOutput, "timeUntilFirstBlock", s.timeUntilFirstBlock, "usec",						float32(s.timeUntilFirstBlock/1000.0),				"msec")
	fmt.Fprintln(parserOutput, "timeUntilLastBlock", s.timeUntilLastBlock, "usec",							float32(s.timeUntilLastBlock/1000.0),				"msec")
	fmt.Fprintln(parserOutput, "timeUntilAllDetectorsEnabled", s.timeUntilAllDetectorsEnabled, "usec", 	float32(s.timeUntilAllDetectorsEnabled/1000.0),		"msec")
	fmt.Fprintln(parserOutput, "timeUntilFirstDetectorsEnabled", s.timeUntilFirstDetectorsEnabled, "usec", float32(s.timeUntilFirstDetectorsEnabled/1000.0),	"msec")
	fmt.Fprintln(parserOutput, "timeUntilFirstTimeout", s.timeUntilFirstTimeout, "usec",					float32(s.timeUntilFirstTimeout/1000.0),				"msec")
	fmt.Fprintln(parserOutput, "timeUntilResponsive", s.timeUntilResponsive, "usec",						float32(s.timeUntilResponsive/1000.0),				"msec")
	fmt.Fprintln(parserOutput, "timeUntilFullyResponsive", s.timeUntilFullyResponsive, "usec", 			float32(s.timeUntilFullyResponsive/1000.0),			"msec")
	fmt.Fprintln(parserOutput, "=================================================================")
}

type canaryStamps struct {
	timeout int64
	detectorEnable int64
	serverUp int64
}

type detectorStamps struct {
	init int64
	firstDetection int64
	firstBlocking int64
}

func analyseLogs(logs chan []byte){
	cluster := make(map[string]string)
	canaries := make(map[string]canaryStamps)
	detectors := make(map[string]detectorStamps)
	attacks := make([]dDos,0)
	malices := make(map[string]string)

	defer func() {
		for c, can := range canaries {
			fmt.Fprintln(parserOutput, "canary ", c, "times: ", can.timeout, can.detectorEnable, can.serverUp)
		}
		for c, cl := range cluster {
			fmt.Fprintln(parserOutput, "node ", c, cl)
		}
		for i, a := range attacks {
			fmt.Fprintln(parserOutput, "attack: ", i, a.started, a.cleanStart, a.startTime)
		}
		// TODO error checking
		parserOutput.Sync()
		parserOutput.Close()
	}()

	// OPT could use the length of cluster
	nodes := 0
	for {
		// log deconstruction
		msg := <-logs
		// print(string(msg))
		toks := strings.SplitN(string(msg), " ", 5)
		node := toks[0]
		pod := toks[1]
		//cmd := toks[2]
		// OPT no error checking bellow! for it is ok
		timestamp, _ := strconv.ParseInt(toks[3], 10, 64)
		log := toks[4]
		if strings.Contains(log, "SIGNAL") {return}

		// EVENTS
		// detect how many nodes are connected to server via flow-server
		if strings.Contains(pod, "flow-server") {
			a, e := cluster[node]
			if e {
				if strings.Compare(a, pod) != 0 {
					// EVENT for some reason flow-server restarted
					cluster[node] = pod
					fmt.Fprintln(parserOutput, timestamp, "flow-server on node ", node, "restarted. New server: ", pod)
				}	
				if strings.Contains(log, "executed command") &&
				   strings.Contains(log, "block") {
					   fmt.Fprintln(parserOutput, timestamp, "flow-server ", pod, "blocked applied")
				   }
				// I don't remember why the line bellow exists
				cluster[node] = a
			} else {
				// INIT EVENT a new node detected
				cluster[node] = pod
				nodes++
				fmt.Fprintln(parserOutput, timestamp, "New node detected: ", node, pod)
			}
		}

		// EVENTS
		// detect starting of ddos attack
		if strings.Contains(pod, "malice") {
			// ingore log for now it is useless
			if strings.Contains(log, "SIGNAL") {continue}

			// count malices
			_, e := malices[pod]
			if !e {
				malices[pod] = node
			}
			// attacks is a list, and for now it will have only one element
			if len(attacks) == 0 {
				// EVENT new attack detected
				d := dDos{true, false, timestamp, 0, 0, nodes, stats{0, 0, 0, 0, 0, 0, 0}}
				attacks = append(attacks, d)
				fmt.Fprintln(parserOutput, timestamp, "[!] ddos attack initiated")
			} else {
				attack := &(attacks[len(attacks) - 1])
				// EVENT in case of out of order logs, check if we have
				// missed the correct starting point
				if !attack.cleanStart {
					if attack.startTime > timestamp {
						attack.startTime = timestamp
						fmt.Fprintln(parserOutput, timestamp, "missed initial log of attack")
					} else if (timestamp - attack.startTime) > 1000000 {
						// mark it as clean-start. no chance we missed 
						// log one second earlier
						attack.cleanStart = true
						fmt.Fprintln(parserOutput, timestamp, "ddos: initial timestamp validated")
					}
				}
			}
		}

		// EVENTS
		// detect first timeout of canary
		if strings.Contains(pod, "canary") {
			// ingore log
			if strings.Contains(log, "SIGNAL") {continue}
			// attacks is a list, and for now it will have only one element
			if len(attacks) == 0 {
				if strings.Contains(log, "Canary connection timeout") {
					// canary timed out but no attack is present
					fmt.Fprintln(parserOutput, timestamp, "canary timed out but no attack is present")
				}
			} else {
				attack := &(attacks[len(attacks) - 1])
				c, e := canaries[pod]
				if e {
					if attack.started {
						if strings.Contains(log, "Enabled detectors") {
							if c.detectorEnable == 0 {
								c.detectorEnable = timestamp
							}
							fmt.Fprintln(parserOutput, timestamp, "canary ", pod,"enabled detectors")
						}
						if c.serverUp == 0 && 
						   strings.Contains(log, "Response in") {
						   // server is responve again
							c.serverUp = timestamp
							fmt.Fprintln(parserOutput, timestamp, "canary ", pod,"connected to server again")
							if attack.reconnections == nodes {
								// every node can access the attacked server
								attack.st.timeUntilResponsive = timestamp - attack.downTime
							} 
							attack.reconnections--
							if attack.reconnections == 0 {
								// every node can access the attacked server
								attack.st.timeUntilFullyResponsive = timestamp - attack.downTime
								attack.st.printStats()
							} 
						}
					}
					canaries[pod] = c
				} else {
					if strings.Contains(log, "Canary connection timeout") {
						canaries[pod] = canaryStamps{timestamp, 0, 0}
						fmt.Fprintln(parserOutput, timestamp, "canary ", pod,"timed out")
						if attack.st.timeUntilFirstTimeout == 0 {
							attack.st.timeUntilFirstTimeout = timestamp - attack.startTime
							attack.downTime = timestamp
						}
					}
				}
			}
		}

		// EVENTS
		// detector stuff
		if strings.Contains(pod, "detector") {
			if len(attacks) == 0 {
				if strings.Contains(log, "Received IP") {
					fmt.Fprintln(parserOutput, timestamp, "detector notified but no attack is present");
				} else if strings.Contains(log, "new connection") {
					fmt.Fprintln(parserOutput, timestamp, "detector new connection detected but no attack is present")
				} else if strings.Contains(log, "block") {
					fmt.Fprintln(parserOutput, timestamp, "detector blocking command sent but no attack is present")
				}
			} else {

				attack := &(attacks[len(attacks) - 1])
				d, e := detectors[pod]
				if e {
					if strings.Contains(log, "Received IP") {
						fmt.Fprintln(parserOutput, timestamp, "detector ", pod, "notified again");
					} else if strings.Contains(log, "new connection") {
						if d.firstDetection == 0 {
							d.firstDetection = timestamp
						}
						fmt.Fprintln(parserOutput, timestamp, "detector ", pod, "detected new connection")
					} else if strings.Contains(log, "block") {
						if d.firstBlocking == 0 {
							d.firstBlocking = timestamp
							if attack.st.timeUntilFirstBlock == 0 {
								attack.st.timeUntilFirstBlock = timestamp - attack.downTime
							}
						}
						attack.blockedConnections++
						if attack.blockedConnections == len(malices) {
							attack.st.timeUntilLastBlock = timestamp - attack.downTime
						}
						fmt.Fprintln(parserOutput, timestamp, "detector ", pod, " send blocking command")
					}
					detectors[pod] = d
				} else {
					if strings.Contains(log, "Received IP") {
						detectors[pod] = detectorStamps{timestamp, 0 ,0}
						if attack.st.timeUntilFirstDetectorsEnabled == 0 {
							attack.st.timeUntilFirstDetectorsEnabled = timestamp - attack.downTime
						}
						if len(detectors) == nodes {
							attack.st.timeUntilAllDetectorsEnabled = timestamp - attack.downTime
						}
						fmt.Fprintln(parserOutput, timestamp, "detector ", pod, " notified for first time");
					}
				}

			}
		}

	} // main loop
}

func printLogs(logs chan []byte){
	for {
		msg := <-logs
		print(string(msg))
		// _ = <-logs
	}
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
