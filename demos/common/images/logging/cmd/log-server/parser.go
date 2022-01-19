package main

import (
	"fmt"
	"strings"
	"strconv"
	"os"
	"log"
)

var (
	parserOutput *os.File
	cluster map[string]*nodePods
	nodes int
	attack dDos
	canaries map[string]*canaryStamps
	detectors map[string]*detectorStamps
	malices map[string]*maliceStamps
	// alices map[string]*aliceStamps
)

func init() {
	var err error
	parserOutput, err = os.OpenFile("/tsi/parser.log", os.O_APPEND|os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		log.Println(err)
		return
	}

	cluster = make(map[string]*nodePods)
	nodes = 0
	canaries = make(map[string]*canaryStamps)
	detectors = make(map[string]*detectorStamps)
	attack.active = false
	malices = make(map[string]*maliceStamps)
	// alices = make(map[string]*aliceStamps)
}

// TODO fix float printing to something more human readable
func (s stats) printStats() {
	min := int64(0)
	for _, malice := range malices {
		stmp := malice.attackRate.firstT
		if min == 0 || min > stmp{ min = stmp}
	}
	max := int64(0)
	for _, malice := range malices {
		stmp := malice.attackRate.firstT
		fmt.Fprintln(parserOutput, malice, "data rate",malice.attackRate.getDataRate(), "MB/s")
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

/*** Utils ***/

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
* eg: NOTE
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

type dataRate struct {
	firstT int64
	data int
	latestT int64
}

func (dr *dataRate) init(now int64) {
	dr.firstT = now
	dr.latestT = now
	dr.data = 0
}

func (dr *dataRate) dataSum(log string, now int64) {
	words := strings.Split(log, " ")
	data, _ := strconv.Atoi(words[len(words) - 1])
	dr.data += data
	dr.latestT = now
}

// returns data rate in MBi/sec
// returns -1 if there are not enough data
func (dr *dataRate) getDataRate() int64 {
	if dr.latestT == dr.firstT { return -1 }
	return (int64(dr.data) / (dr.latestT - dr.firstT))
}

/*** dDos metrics helpers ***/

// Refactoring TODO include only interesting timestamps
type dDos struct {
	active bool
	downTime int64
	blockedConnections int
	reconnections int
	st stats
}

func (d *dDos) start(now int64) {
	d.active = true
	d.downTime = 0
	d.blockedConnections = 0
	d.reconnections = nodes
	// TODO d.st = 

}

/*** ddos malice ***/
type maliceStamps struct {
	name string
	node string
	attackRate dataRate
	// NOTE logs for failure don't show up here for some reason. They should.
	serverBlocked rippleStamp
}

func initMalice(now int64, name, node string) (m *maliceStamps){
	newMalice := maliceStamps{
		name: name,
		node: node,
	}
	newMalice.serverBlocked.init(now, 100, false)
	return (*maliceStamps)(&newMalice)
}

/*** canary ***/
type canaryStamps struct {
	name string
	node string
	detectorEnable int64
	serverResponsive rippleStamp
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
		detectorEnable: 0,
		}
		newCanary.serverResponsive.init(now, 2, false)
	cluster[node].canary = (* canaryStamps)(&newCanary)
	return (* canaryStamps)(&newCanary)
}

/*** detector ***/
type detectorStamps struct {
	name string
	node string
	// first time notified by canary after attack initiation
	notified int64
	firstDetection int64
	// first blocking command sent to flow server
	firstBlocking int64
}

func initDetector(now int64, name, node string) *detectorStamps {
	newDetector := detectorStamps {
		name: name,
		node: node,
		notified: now,
		firstDetection: 0,
		firstBlocking: 0,
	}
	cluster[node].detector = (* detectorStamps)(&newDetector)
	return (* detectorStamps)(&newDetector)
}

/*** node ***/
type nodePods struct {
	canary *canaryStamps
	detector *detectorStamps
	flow string
}

func initNode(name, flow string) *nodePods {
	newNode := nodePods{
		canary: nil,
		detector: nil,
		flow: flow,
	}
	cluster[name] = (*nodePods)(&newNode)
	nodes++
	return (*nodePods)(&newNode)
}

func analyseLogs(logs chan []string){

	// NOTE depr
	defer func() {
		for c, can := range canaries {
			fmt.Fprintln(parserOutput, "canary ", c, "times: ", 
						can.detectorEnable)
		}
		for c, cl := range cluster {
			fmt.Fprintln(parserOutput, "node ", c, cl)
		}
		// TODO error checking
		parserOutput.Sync()
		parserOutput.Close()
	}()

	for {
		// log deconstruction
		toks := <-logs

		node := toks[0]
		pod := toks[1]
		//cmd := toks[2]
		// OPT no error checking bellow! for it is ok
		timestamp, _ := strconv.ParseInt(toks[3], 10, 64)
		log := toks[4]
		if strings.Contains(log, "SIGNAL") {
			attack.st.printStats()
			return}

		// EVENTS
		// detect how many nodes are connected to server via flow-server
		if strings.Contains(pod, "flow-server") {
			a, e := cluster[node]
			if e {
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
					attack.blockedConnections++
					if attack.blockedConnections == len(malices) {
						attack.st.timeUntilLastBlock = timestamp - attack.downTime
					}
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
			// count malices
			m, e := malices[pod]
			if !e {
				// new malice detected.
				if strings.Contains(log, "OK"){
					malices[pod] = initMalice(timestamp, pod, node)
					malices[pod].attackRate.dataSum(log, timestamp)
					attack.start(timestamp)
					fmt.Fprintln(parserOutput, timestamp, "[!] ddos attack initiated")
				}
			} else {
				if attack.active == false {
					// EVENT new attack detected
					attack.start(timestamp)
					if attack.active == false {
						fmt.Fprintln(parserOutput, timestamp, "ERR: [X] active state of attack did not change")
					}
					fmt.Fprintln(parserOutput, timestamp, "[!] ddos attack initiated")
				} else {
					if strings.Contains(log, "OK"){
						m.attackRate.dataSum(log, timestamp)
					}
					m.serverBlocked.toggle(timestamp, false, strings.Contains(log, "OK"))
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
				if strings.Contains(log, "Canary connection timeout") {
					canaries[pod] = initCanary(timestamp, pod, node)
					fmt.Fprintln(parserOutput, timestamp, "canary ", pod,"timed out")
					attack.start(timestamp)
					fmt.Fprintln(parserOutput, timestamp, "[!] possible attack initiated")
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
							// every node can access the attacked server
							attack.active = false
							attack.st.timeUntilFullyResponsive = timestamp - attack.downTime
							attack.st.printStats()
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
					fmt.Fprintln(parserOutput, timestamp, "WRN: detector notified but no attack is present");
				} else if strings.Contains(log, "new connection") {
					fmt.Fprintln(parserOutput, timestamp, "WRN: detector new connection detected but no attack is present")
				} else if strings.Contains(log, "block") {
					fmt.Fprintln(parserOutput, timestamp, "WRN: detector blocking command sent but no attack is present")
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
						fmt.Fprintln(parserOutput, "WRN:", timestamp, "detector ", pod, "detected new connection", log)
					} else if strings.Contains(log, "block") {
						if d.firstBlocking == 0 {
							d.firstBlocking = timestamp
							if attack.st.timeUntilFirstBlock == 0 {
								attack.st.timeUntilFirstBlock = timestamp - attack.downTime
							}
						}
						fmt.Fprintln(parserOutput, timestamp, "detector ", pod, " send blocking command", log)
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

