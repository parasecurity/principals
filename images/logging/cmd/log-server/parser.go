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
	alices map[string]*aliceStamps
	allAlices *aliceStamps
)

func init() {
	var err error
	parserOutput, err = os.OpenFile("/tsi/parser.log", os.O_APPEND|os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		log.Println(err)
		return
	}
	cluster = make(map[string]*nodePods)
	attack.active = false
	attack.passed = false
	nodes = 0
	canaries = make(map[string]*canaryStamps)
	detectors = make(map[string]*detectorStamps)
	attack.active = false
	attack.reconnections = 0
	malices = make(map[string]*maliceStamps)
	alices = make(map[string]*aliceStamps)
	attack.blockedIPs = make(map[string]string)
	allAlices = initAlice(0, "all", "cluster")
}

/*** dDos metrics helpers ***/

// Refactoring TODO include only interesting timestamps
type dDos struct {
	active bool
	passed bool
	responding bool
	downTime int64
	blockedIPs map[string]string
	reconnections int
	st stats
}

func (d *dDos) start(now int64) {
	d.active = true
	d.passed = false
	d.responding = false
	d.downTime = 0
	d.st.attackInitiation = now
}

type aliceStamps struct {
	name string
	node string
	preAttackRate dataRate
	attackRate dataRate
	postAttackRate dataRate
}

func initAlice(now int64, name, node string) (m *aliceStamps){
	newAlice := aliceStamps{
		name: name,
		node: node,
	}
	newAlice.preAttackRate.init(now)
	return (*aliceStamps)(&newAlice)
}

/*** ddos malice ***/
type maliceStamps struct {
	name string
	node string
	attackRate dataRate
	respAttackRate dataRate
	// NOTE logs for failure don't show up here for some reason. They should.
	serverBlocked rippleStamp
}

func initMalice(now int64, name, node string) (m *maliceStamps){
	newMalice := maliceStamps{
		name: name,
		node: node,
	}
	newMalice.attackRate.init(now)
	newMalice.respAttackRate.init(now)
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
		newCanary.serverResponsive.init(now, 10, false)
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
		notified: 0,
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
			return
		}

		// wait 10 seconds after canaries reconnections to print stats
		// NOTE add flag for report and do not change value to passed
		if attack.passed && (timestamp - attack.st.timeUntilFullyResponsive) > 15 * tsiSecond {
			attack.active = false
			attack.passed = false
			attack.st.printStats()
		}

		// EVENTS
		// detect how many nodes are connected to server via flow-server
		if strings.Contains(pod, "flow-server") {
			a, e := cluster[node]
			if ! e {
				a = initNode(node, pod)
				fmt.Fprintln(parserOutput, timestamp, 
							"New node detected: ", node, pod)
			}
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
				words := strings.Split(log, "\"")
				IP := words[len(words)-2]
				_, ex := attack.blockedIPs[IP]
				if !ex {
					attack.blockedIPs[IP] = node
				}
				if len(attack.blockedIPs) == len(malices) {
					attack.st.timeUntilLastBlock = timestamp
				}
				fmt.Fprintln(parserOutput, timestamp, 
				"flow-server ", pod, "blocked applied", IP)
			}
			//   I don't remember why the line bellow exists <--comment before refactoring
			// now it has some value, maybe it is not needed, as the map consists of pointers
			// I'm not sure about golang's sorcery
			cluster[node] = a
		}

		// EVENTS
		// detect starting of ddos attack
		if strings.HasPrefix(pod, "alice") {
			if ! strings.Contains(log, "Response") {
				continue
			}
			a, e := alices[pod]
			if !e {
				// new alice detected.
				alices[pod] = initAlice(timestamp, pod, node)
				a, _ = alices[pod]
			}
			if attack.active {
				a.attackRate.packetCount++
				allAlices.attackRate.packetCount++
				if strings.Contains(log, "OK"){
					a.attackRate.dataSum(log, timestamp)
					a.attackRate.packetOK++
					allAlices.attackRate.dataSum(log, timestamp)
					allAlices.attackRate.packetOK++
				}
			} else {
				if attack.passed {
					a.postAttackRate.packetCount++
					allAlices.postAttackRate.packetCount++
					if strings.Contains(log, "OK"){
						a.postAttackRate.dataSum(log, timestamp)
						a.postAttackRate.packetOK++
						allAlices.postAttackRate.dataSum(log, timestamp)
						allAlices.postAttackRate.packetOK++
					}
				} else {
					a.preAttackRate.packetCount++
					allAlices.preAttackRate.packetCount++
					if strings.Contains(log, "OK"){
						a.preAttackRate.dataSum(log, timestamp)
						a.preAttackRate.packetOK++
						allAlices.preAttackRate.dataSum(log, timestamp)
						allAlices.preAttackRate.packetOK++
					}
				}
			}
			alices[pod] = a
		}

		// EVENTS
		// detect starting of ddos attack (udp/syn)
		if strings.HasPrefix(pod, *attackPodName) {
			_, e := malices[pod]
			if !e {
				if strings.Contains(log, *attackInitLog) {
					malices[pod] = initMalice(timestamp, pod, node)
					if !attack.active && !attack.passed {
						attack.start(timestamp)
						fmt.Fprintln(parserOutput, timestamp, "[!] ddos attack initiated")
					} else if attack.st.attackInitiation > timestamp {
						attack.st.attackInitiation = timestamp
					}
				}
			}
		}

		// EVENTS
		// detect starting of ddos attack (http)
		if strings.Contains(pod, "malice") {
			// count malices
			m, e := malices[pod]
			if !e {
				// new malice detected.
				// init with invalide very late timestamp so it will be validated
				// on 328 line in case of initial bad connection BUG 
				malices[pod] = initMalice(timestamp + tsiSecond*60*60*24, pod, node)
				if strings.Contains(log, "OK"){
					malices[pod] = initMalice(timestamp, pod, node)
					malices[pod].attackRate.packetOK++
					// malices[pod].attackRate.dataSum(log, timestamp)
					if !attack.active && !attack.passed{
						attack.start(timestamp)
						fmt.Fprintln(parserOutput, timestamp, "[!] ddos attack initiated")
					} else if attack.st.attackInitiation > timestamp {
						attack.st.attackInitiation = timestamp
					}
				}
				malices[pod].attackRate.packetCount++
			} else {
				if !attack.active {
					// EVENT new attack detected
					// TODO clean up, do resete or nothing
					if strings.Contains(log, "OK"){
						if m.attackRate.firstT > timestamp {
							m.attackRate.firstT = timestamp 
						}
						if !attack.passed {
							attack.start(timestamp)
							fmt.Fprintln(parserOutput, timestamp, "[!] ddos attack initiated")
						}
						malices[pod].attackRate.packetOK++
					}
					malices[pod].attackRate.packetCount++
				} else {
					if attack.responding {
						m.respAttackRate.packetCount++
					} else {
						m.attackRate.packetCount++
					}
					if strings.Contains(log, "OK"){
						if m.attackRate.firstT > timestamp {
							m.attackRate.firstT = timestamp 
						}
						if attack.st.attackInitiation > timestamp {
							attack.st.attackInitiation = timestamp
						}
						if attack.responding {
							m.respAttackRate.packetOK++
						} else {
							m.attackRate.packetOK++
						}
					}
					m.serverBlocked.toggle(timestamp, false, strings.Contains(log, "OK"))
					m.serverBlocked.toggle(timestamp, true, strings.Contains(log, "Fail"))
				}
				malices[pod] = m
			}
		}

		// EVENTS
		// canary
		if strings.Contains(pod, "canary") {
			// ingore log
			if strings.Contains(log, "SIGNAL") {continue}
			// attack is marked as started by malices' logs
			if !attack.active && !attack.passed {
				// here we produce info. TODO false detection
				c, e := canaries[pod]
				if !e {
					if strings.Contains(log, "Canary connection timeout") {
						canaries[pod] = initCanary(timestamp, pod, node)
						attack.reconnections++
						fmt.Fprintln(parserOutput, timestamp, "canary ", pod,"timed out")
					}
				} else {
					if c.serverResponsive.toggle(timestamp, true, strings.Contains(log, "response in")) {
						attack.reconnections--
						fmt.Fprintln(parserOutput, timestamp, "canary ", pod,"connected to server again")
					}
					if c.serverResponsive.toggle(timestamp, false, strings.Contains(log, "Canary connection timeout")){
						attack.reconnections++
						fmt.Fprintln(parserOutput, timestamp, "canary ", pod,"timed out")
					}
				}

			} else {
				c, e := canaries[pod]
				if e {
					if strings.Contains(log, "Enabled detectors") {
						if c.detectorEnable == 0 || c.detectorEnable > timestamp{
							c.detectorEnable = timestamp
							if attack.st.timeUntilFirstDetectorsEnabled == 0 {
								fmt.Fprintln(parserOutput, "[!] system is responding to attack")
								attack.st.timeUntilFirstDetectorsEnabled = timestamp
								attack.responding = true
							}
						}
						fmt.Fprintln(parserOutput, timestamp, "canary ", pod,"enabled detectors")
					}
					if c.serverResponsive.toggle(timestamp, true, strings.Contains(log, "Response in")) {
						attack.reconnections--
						fmt.Fprintln(parserOutput, timestamp, "canary ", pod,"connected to server again")
						if attack.reconnections == 0 {
							// every node can access the attacked server
							fmt.Fprintln(parserOutput, "[!] server fully responsive")
							attack.passed = true
							attack.active = false
							attack.st.timeUntilFullyResponsive = c.serverResponsive.timestampT
						} 
					}
					if c.serverResponsive.toggle(timestamp, false, strings.Contains(log, "Canary connection timeout")){
						fmt.Fprintln(parserOutput, timestamp, "canary ", pod,"timed out")
						attack.reconnections++
					}
					canaries[pod] = c
				} else {
					if strings.Contains(log, "Canary connection timeout") {
						canaries[pod] = initCanary(timestamp, pod, node)
						attack.reconnections++
						fmt.Fprintln(parserOutput, timestamp, "canary ", pod,"timed out")
					}
				}
			}
		}

		// EVENTS
		// detector stuff
		if strings.Contains(pod, "detector") {
			if !attack.active && !attack.passed{
				if strings.Contains(log, "Received IP") {
					fmt.Fprintln(parserOutput, timestamp, "WRN: detector notified but no attack is present");
				} else if strings.Contains(log, "new connection") {
					fmt.Fprintln(parserOutput, timestamp, "WRN: detector new connection detected but no attack is present")
				} else if strings.Contains(log, "block") {
					fmt.Fprintln(parserOutput, timestamp, "WRN: detector blocking command sent but no attack is present")
				}
			} else {
				d, e := detectors[pod]
				if !e {
					if strings.Contains(log, "Received IP") {
						d = initDetector(timestamp, pod, node)
					} else { continue }
				}
				if strings.Contains(log, "Received IP") {
					// info
					if d.notified == 0 {
						fmt.Fprintln(parserOutput, timestamp, "detector ", pod, " notified for first time");
						d.notified = timestamp
					}
				} else if strings.Contains(log, "new connection") {
					if d.firstDetection == 0 {
						d.firstDetection = timestamp
					}
					// fmt.Fprint(parserOutput, "WRN:", timestamp, "detector ", pod, "detected new connection", log)
				} else if strings.Contains(log, "block") {
					if d.firstBlocking == 0 {
						d.firstBlocking = timestamp
						if attack.st.timeUntilFirstBlock == 0 {
							attack.st.timeUntilFirstBlock = timestamp
						}
					}
					fmt.Fprintln(parserOutput, timestamp, "detector ", pod, " send blocking command")
				}
				detectors[pod] = d
			}
		}

	} // main loop
}

