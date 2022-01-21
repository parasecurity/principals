package main

import (
	"strconv"
	"strings"
	"log"
)
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
	data float64
	min float64
	max float64
	packetCount int
	packetOK int
	latestT int64
}

func (dr *dataRate) init(now int64) {
	dr.firstT = now
	dr.latestT = now
	dr.data = 0
	dr.packetCount = 0
	dr.packetOK = 0
}

func (dr *dataRate) dataSum(msg string, now int64) {
	words := strings.Split(strings.TrimSpace(msg), " ")
	ping := words[len(words) - 1]
	data, err := strconv.ParseFloat(ping[:len(ping)-4], 64)
	if err != nil {
		log.Println(err)
		return
	}
	// log.Println(len(words), "wordsLast", words[len(words)-1], "data sum:", data)
	dr.data += data
	if dr.max < data {
		dr.max = data
	}
	if dr.min ==0 || dr.min > data {
		dr.min = data
	}
	dr.latestT = now
}

// returns data rate in KBps
// returns -1 if there are not enough data
func (dr *dataRate) getPercent() float64 {
	if dr.packetCount == 0 { return -1 }
	// NOTE maybe we need to multiply by 8
	return (float64(dr.packetOK) / float64(dr.packetCount))*100
}

// returns data rate in KBps
// returns -1 if there are not enough data
func (dr *dataRate) getDataRate() float64 {
	if dr.latestT == dr.firstT { return -1 }
	// NOTE maybe we need to multiply by 8
	return (float64(dr.data) / float64(dr.latestT - dr.firstT))*8*1000
}

