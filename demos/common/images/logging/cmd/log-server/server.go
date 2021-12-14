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
		// select for efficiency
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
*         ^				^				^
*		don't bother with this commend block yet
*/

// TODO fix float printing to something more human readable
func (s stats) printStats() {
	fmt.Fprintln(parserOutput, "=================================================================")
	fmt.Fprintln(parserOutput, "timeUntilFirstBlock", s.timeUntilFirstBlock, "usec",						float32(s.timeUntilFirstBlock/1000.0),				"msec")
	fmt.Fprintln(parserOutput, "timeUntilLastBlock", s.timeUntilLastBlock, "usec",							float32(s.timeUntilLastBlock/1000.0),				"msec")
	fmt.Fprintln(parserOutput, "timeUntilAllDetectorsEnabled", s.timeUntilAllDetectorsEnabled, "usec", 	float32(s.timeUntilAllDetectorsEnabled/1000.0),		"msec")
	fmt.Fprintln(parserOutput, "timeUntilFirstDetectorsEnabled", s.timeUntilFirstDetectorsEnabled, "usec", float32(s.timeUntilFirstDetectorsEnabled/1000.0),	"msec")
	fmt.Fprintln(parserOutput, "timeUntilFirstTimeout", s.timeUntilFirstTimeout, "usec",					float32(s.timeUntilFirstTimeout/1000.0),				"msec")
	fmt.Fprintln(parserOutput, "timeUntilResponsive", s.timeUntilResponsive, "usec",						float32(s.timeUntilResponsive/1000.0),				"msec")
	fmt.Fprintln(parserOutput, "timeUntilFullyResponsive", s.timeUntilFullyResponsive, "usec", 			float32(s.timeUntilFullyResponsive/1000.0),			"msec")
	fmt.Fprintln(parserOutput, "=================================================================")

	min := int64(0)
	for _, malice := range malices {
		stmp := malice.firstAttack.timestamp
		if min == 0 || min > stmp{ min = stmp}
	}
	max := int64(0)
	for _, malice := range malices {
		stmp := malice.firstAttack.timestamp
		if max == 0 || max < stmp{ max = stmp}
	}
	attack.st.attackInitiation = max
	point0 := max
	fmt.Fprintln(parserOutput, "from first malice to last: +", max-min)

	min = 0
	for _, can := range canaries {
		stmp := can.serverResponsive.timestampF
		if min == 0 || min > stmp {min = stmp}
	}
	fmt.Fprintln(parserOutput, "first canary timeout +", min - point0)
	max = 0
	for _, can := range canaries {
		stmp := can.serverResponsive.timestampF
		if max == 0 || max < stmp {max = stmp}
	}
	fmt.Fprintln(parserOutput, "last canary timeout +", max - point0)

	min = 0
	for _, can := range canaries {
		stmp := can.serverResponsive.timestampT
		if min == 0 || min > stmp {min = stmp}
	}
	fmt.Fprintln(parserOutput, "first canary reconnection +", min - point0)
	max = 0
	for _, can := range canaries {
		stmp := can.serverResponsive.timestampT
		if max == 0 || max < stmp {max = stmp}
	}
	fmt.Fprintln(parserOutput, "last canary reconnection +", max - point0)

}

type stats struct {
	timeUntilFirstBlock int64
	timeUntilLastBlock int64
	timeUntilAllDetectorsEnabled int64
	timeUntilFirstDetectorsEnabled int64
	timeUntilFirstTimeout int64
	timeUntilResponsive int64
	timeUntilFullyResponsive int64
	attackInitiation int64
}

type stamp struct {
	// main
	timestamp int64
	valid bool
	threshold int64 // time in microseconds, if zero no validation will be done
}

func (s *stamp) init(now, thr int64) {
	s.timestamp = now
	s.valid = thr == 0
	s.threshold = thr
}

type maliceStamps struct {
	name string
	node string

	firstAttack stamp
	// I know that I don't follow my rules, but using a big threshold it should work
	serverBlocked rippleStamp
}

func initMalice(now int64, name, node string) (m *maliceStamps){
	newMalice := maliceStamps{
		name: name,
		node: node,
	}
	newMalice.firstAttack.init(now, 1000000)
	newMalice.serverBlocked.init(now, 100, false)

	return (*maliceStamps)(&newMalice)
}

/* for now it doesn't produce any output
* it returns true only when the timestamp gets validated and false 
* in ANY other case, including when it is already validated
*
* the caller will never know if an out-of-order log 
* changed the value before validation
*/
func (s *stamp) validate(now int64) bool {
	// EVENT in case of out of order logs, check if we have
	// missed the correct starting point
	if s.valid {return false}

	if s.timestamp > now {
		s.timestamp = now
		// TODO may produce a notification
	} else if (now - s.timestamp) > s.threshold {
		// mark it as clean-start. no chance we missed 
		// log one second earlier
		s.valid = true
		return true
	}
	return false
}

// Refactoring TODO include only interesting timestamps
type dDos struct {
	active bool
	startingPoint stamp
	downTime int64
	blockedConnections int
	reconnections int
	st stats
}

func (d *dDos) start(now int64) {
	d.active = true
	d.startingPoint.init(now, 1000000)
	d.downTime = 0
	d.blockedConnections = 0
	d.reconnections = nodes
	// TODO d.st = 

}

func (d *dDos) validateStart(now int64) {
	if d.startingPoint.validate(now) {
		fmt.Fprintln(parserOutput, now, "ddos: initial timestamp validated")
	}
}

/* use rippleStamp instead of stamp when all statements 
* bellow are true:
*
*	- you are trying to detect a change of network connectivity
*	- the sequense of logs that are tracked for that measurement
*	  is produced by a single pod in the same thread!
*	- last, but not least, when you want to measure both times of changing state!
*
*	timestampT: time without ripple from F to T
*	timestampF: time from T to F at beggining of a possible ripple
*
* eg: TODO
* 
* usefull for canaries, detectors, flow-servers
*/
type rippleStamp struct {
	timestampT int64
	timestampF int64
	inRipple bool
	rippleCount int
	thr int
	state bool
}

/*
* canary metrics initialization
* canary metrics should only be initialized when
* a canary times out
*/
func initCanary(now int64, name, node string) *canaryStamps {
	newCanary := canaryStamps{
		name: name, 
		node: node, 

		firstTimeout: now,
		detectorEnable: 0,
		serverUpTime: 0,
		serverUp: false,
		inRipple: false,
		rippleCount: 2 }
		newCanary.serverResponsive.init(now, 2, false)

	cluster[node].canary = (* canaryStamps)(&newCanary)
	return (* canaryStamps)(&newCanary)
}

func (rs *rippleStamp) init(now int64, thr int, state bool) {
	if state {
		rs.timestampT = now
		rs.timestampF = 0
	} else {
		rs.timestampT = 0
		rs.timestampF = now
	}
	rs.rippleCount = thr
	rs.thr = thr
	
	rs.inRipple = false
	rs.state = state
}

/*
* possibly change the state of rippleStamp
* arguments:
*	now:   timestamp of log triggering the changing
*	to:    the state we should go if check is true
*	check: flag that the event which triggers the state	
*		   changing indicated by to is true
*
*	returns true if the state is indeed changed
*
*	usage: toggle(<time in microseconds>, <state>, <condition that triggers state>)
* eg: myRippleStamp.toggle(ts, true, strings.Contains(log, "we should go to true"))
*     myRippleStamp.toggle(ts, false, strings.Contains(log, "we should go to false"))
*
*/
func (rs *rippleStamp) toggle(now int64, to, check bool) bool {
	if !check {return false}
	if rs.state {
		// T state
		// line bellow may break downtime measurement
		if !to {rs.timestampF = now }
		rs.state = to
		return !rs.state
	}else {
		// F state
		if !to {
			//revert
			rs.rippleCount = rs.thr
		} else {
			// maybe is responsive again
			if rs.rippleCount == rs.thr {
				rs.timestampT = now
			}
			rs.rippleCount--
			if rs.rippleCount == 0 {
				rs.rippleCount = rs.thr
				rs.state = true
			}
		}
		return rs.state
	}
}

type canaryStamps struct {
	name string
	node string

	firstTimeout int64
	detectorEnable int64
	serverUpTime int64
	serverUp bool
	inRipple bool
	rippleCount int
	serverResponsive rippleStamp
}

func initDetector(now int64, name, node string) *detectorStamps {
	newDetector := detectorStamps {
		name: name,
		node: node,

		init: now,
		firstDetection: 0,
		firstBlocking: 0 }

	cluster[node].detector = (* detectorStamps)(&newDetector)
	return (* detectorStamps)(&newDetector)
}

type detectorStamps struct {
	name string
	node string

	// first time notified by canary after attack initiation
	init int64
	firstDetection int64
	// first blocking command sent to flow server
	firstBlocking int64
}

func initNode(name, flow string) *nodePods {
	newNode := nodePods{
		canary: nil,
		detector: nil,
		flow: flow }

	cluster[name] = (*nodePods)(&newNode)
	nodes++
	return (*nodePods)(&newNode)
}

type nodePods struct {
	canary *canaryStamps
	detector *detectorStamps
	flow string
}

var (
	cluster map[string]*nodePods
	nodes int
	attack dDos
	canaries map[string]*canaryStamps
	detectors map[string]*detectorStamps
	malices map[string]*maliceStamps

)


func analyseLogs(logs chan []byte){

	// depr
	defer func() {
		for c, can := range canaries {
			fmt.Fprintln(parserOutput, "canary ", c, "times: ", 
						can.firstTimeout, can.detectorEnable, can.serverUpTime)
		}
		for c, cl := range cluster {
			fmt.Fprintln(parserOutput, "node ", c, cl)
		}
		// TODO error checking
		parserOutput.Sync()
		parserOutput.Close()
	}()

	// OPT could use the length of cluster TODO move to init
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
				// out of scenario! TODO decouple
				if strings.Compare(a.flow, pod) != 0 {
					// EVENT for some reason flow-server restarted
					a.flow = pod
					fmt.Fprintln(parserOutput, timestamp, 
								"flow-server on node ", node, 
								"restarted. New server: ", pod)
				}	
				// info 
				if strings.Contains(log, "executed command") &&
				   strings.Contains(log, "block") {
					   fmt.Fprintln(parserOutput, timestamp, 
									"flow-server ", pod, "blocked applied")
				}
				//   I don't remember why the line bellow exists <--comment before refactoring
				// now it has some value, maybe it is not needed, as the map consists of pointers
				// I'm not sure about golang's sorcery
				cluster[node] = a
			} else {
				// INIT EVENT a new node detected
				initNode(node, pod)
				fmt.Fprintln(parserOutput, timestamp, 
							"New node detected: ", node, pod)
			}
		}

		// EVENTS
		// detect starting of ddos attack
		if strings.Contains(pod, "malice") {
			// ingore log for now it is useless
			if strings.Contains(log, "SIGNAL") {continue}
			// count malices
			m, e := malices[pod]
			if !e {
				// new malice detected.
				if strings.Contains(log, "Response status") {
					malices[pod] = initMalice(timestamp, pod, node)
				}
			} else {
				if attack.active == false {
					// EVENT new attack detected
					attack.start(timestamp)
					fmt.Fprintln(parserOutput, timestamp, "[!] ddos attack initiated")
				} else {
					if strings.Contains(log, "Response status") {
						attack.validateStart(timestamp)
						m.firstAttack.validate(timestamp)
					}
					m.serverBlocked.toggle(timestamp, false, strings.Contains(log, "Response status"))
					m.serverBlocked.toggle(timestamp, true, strings.Contains(log, "Fail"))
				}
			}
		}

		// EVENTS
		// canary
		if strings.Contains(pod, "canary") {
			// ingore log
			if strings.Contains(log, "SIGNAL") {continue}
			// attack is marked as started by malices' logs
			if !attack.active {
				// here we produce info. TODO false detection
				// TODO add enabling detectors out of attack
				if strings.Contains(log, "Canary connection timeout") {
					// canary timed out but no attack is present
					fmt.Fprintln(parserOutput, timestamp, "canary timed out but no attack is present")
				}
			} else {
				c, e := canaries[pod]
				if e {
					if strings.Contains(log, "Enabled detectors") {
						if c.detectorEnable == 0 {
							c.detectorEnable = timestamp
						}
						fmt.Fprintln(parserOutput, timestamp, "canary ", pod,"enabled detectors")
					}
					if c.serverResponsive.toggle(timestamp, true, strings.Contains(log, "Response in")) {
						attack.reconnections--
						fmt.Fprintln(parserOutput, timestamp, "canary ", pod,"connected to server again")
						if attack.reconnections == 0 {
							// every node can access the attacked server TODO fix out of order
							attack.active = false
							attack.st.timeUntilFullyResponsive = timestamp - attack.downTime
						} 
					}

					if c.serverResponsive.toggle(timestamp, false, strings.Contains(log, "Canary connection timeout")){
						fmt.Fprintln(parserOutput, timestamp, "canary ", pod,"timed out")
					}
					canaries[pod] = c
				} else {
					if strings.Contains(log, "Canary connection timeout") {
						canaries[pod] = initCanary(timestamp, pod, node)
						fmt.Fprintln(parserOutput, timestamp, "canary ", pod,"timed out")
					}
				}
			}
		}

		// EVENTS
		// detector stuff
		if strings.Contains(pod, "detector") {
			if !attack.active {
				if strings.Contains(log, "Received IP") {
					fmt.Fprintln(parserOutput, timestamp, "detector notified but no attack is present");
				} else if strings.Contains(log, "new connection") {
					fmt.Fprintln(parserOutput, timestamp, "detector new connection detected but no attack is present")
				} else if strings.Contains(log, "block") {
					fmt.Fprintln(parserOutput, timestamp, "detector blocking command sent but no attack is present")
				}
			} else {

				d, e := detectors[pod]
				if e {
					if strings.Contains(log, "Received IP") {
						// info
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
						detectors[pod] = initDetector(timestamp, pod, node)
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

/* this function prints logs to stdout in order to be managed by kubernetes logging system
* logs are also written to tsi shared directory in tsi.log
* IMPORTANT! Logs may be out of order in same cases of bursting. 
* Out of order logs do not affect the parsing system above
*/
func printLogs(logs chan []byte){
	for {
		msg := <-logs
		print(string(msg))
		fmt.Fprint(clusterLogging, string(msg))
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
