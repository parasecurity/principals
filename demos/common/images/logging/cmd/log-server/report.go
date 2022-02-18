package main

import (
	"fmt"
	"log"
	"strconv"
)

type stats struct {
	attackInitiation int64
	timeUntilFirstTimeout int64
	timeUntilFirstDetectorsEnabled int64
	timeUntilAllDetectorsEnabled int64
	timeUntilFirstBlock int64
	timeUntilLastBlock int64
	timeUntilResponsive int64
	timeUntilFullyResponsive int64
}

// TODO fix float printing to something more human readable
func (s stats) printStats() {
	fmt.Fprintln(parserOutput, "=== *** Report *** ===")
	fmt.Fprintln(parserOutput, s)
	min := int64(0)

	sumIP := make(map[string]int)
	for n := range cluster {
		sumIP[n] = 0
	}

	fmt.Fprintln(parserOutput, "Blocked IPs")
	for ip, n := range attack.blockedIPs {
		fmt.Fprintln(parserOutput, ip, ":", n)
		sumIP[n]++
	}

	for _, malice := range malices {
		stmp := malice.attackRate.firstT
		if min == 0 || min > stmp{ min = stmp}
	}
	max := int64(0)
	for _, malice := range malices {
		stmp := malice.attackRate.firstT
		if max == 0 || max < stmp{ max = stmp}
	}
	if attack.st.attackInitiation > min {
		log.Println("               out of order starting point")
		attack.st.attackInitiation = min
	}
	// attack.st.attackInitiation = max
	// point0 := max
	// fmt.Fprintln(parserOutput, "from first malice to last: +", max-min)

	for _, malice := range malices {
		C := malice.attackRate.packetCount
		var P string
		if C == 0 {
			P = "N/A"
		} else {
			P =	strconv.FormatFloat(malice.attackRate.getPercent(), 'f', -1, 64)
		}
		respC :=	malice.respAttackRate.packetCount 
		var respP string
		if respC == 0 {
			respP = "N/A"
		} else {
			respP =	strconv.FormatFloat(malice.respAttackRate.getPercent(), 'f', -1, 64)
		}
		fmt.Fprintf(parserOutput, "%s bad traffic: pre %s %%(%d/%d), att %s %%(%d/%d)\n", 
					malice.name,
					P, malice.attackRate.packetOK, C,
					respP, malice.respAttackRate.packetOK, respC)
	}

	for _, alice := range alices {
		preC :=	alice.preAttackRate.packetCount 
		var preP string
		if preC == 0 {
			preP = "N/A"
		} else {
			preP =	strconv.FormatFloat(alice.preAttackRate.getPercent(), 'f', -1, 64)
		}

		C := alice.attackRate.packetCount
		var P string
		if C == 0 {
			P = "N/A"
		} else {
			P =	strconv.FormatFloat(alice.attackRate.getPercent(), 'f', -1, 64)
		}

		postC :=	alice.postAttackRate.packetCount 
		var postP string
		if postC == 0 {
			postP = "N/A"
		} else {
			postP =	strconv.FormatFloat(alice.postAttackRate.getPercent(), 'f', -1, 64)
		}

		fmt.Fprintf(parserOutput, "%s goodput: pre %s %%(%d/%d), att %s %%(%d/%d), post %s(%d/%d)\n", 
					alice.name, preP, alice.preAttackRate.packetOK, preC, 
					P, alice.attackRate.packetOK, C, 
					postP, alice.postAttackRate.packetOK, postC) 
		fmt.Fprintf(parserOutput, "  pre min %fms, mean %fms, max %fms\n", alice.preAttackRate.min,( alice.preAttackRate.max + alice.preAttackRate.min )/2, alice.preAttackRate.max)
		fmt.Fprintf(parserOutput, "  att min %fms, mean %fms, max %fms\n", alice.attackRate.min,( alice.attackRate.max + alice.attackRate.min )/2, alice.attackRate.max)
		fmt.Fprintf(parserOutput, "  post min %fms, mean %fms, max %fms\n", alice.postAttackRate.min,( alice.postAttackRate.max + alice.postAttackRate.min )/2, alice.postAttackRate.max)
	}

	preC :=	allAlices.preAttackRate.packetCount 
	var preP string
	if preC == 0 {
		preP = "N/A"
	} else {
		preP =	strconv.FormatFloat(allAlices.preAttackRate.getPercent(), 'f', -1, 64)
	}

	C := allAlices.attackRate.packetCount
	var P string
	if C == 0 {
		P = "N/A"
	} else {
		P =	strconv.FormatFloat(allAlices.attackRate.getPercent(), 'f', -1, 64)
	}

	postC := allAlices.postAttackRate.packetCount 
	var postP string
	if postC == 0 {
		postP = "N/A"
	} else {
		postP =	strconv.FormatFloat(allAlices.postAttackRate.getPercent(), 'f', -1, 64)
	}

	fmt.Fprintln(parserOutput, "Summary of goodput" )
	fmt.Fprintf(parserOutput, "%s goodput: pre %s %%(%d/%d), att %s %%(%d/%d), post %s(%d/%d)\n", 
	allAlices.name, preP, allAlices.preAttackRate.packetOK, preC, 
	P, allAlices.attackRate.packetOK, C, 
	postP, allAlices.postAttackRate.packetOK, postC) 
	fmt.Fprintf(parserOutput, "  pre min %fms, mean %fms, max %fms\n", allAlices.preAttackRate.min,( allAlices.preAttackRate.max + allAlices.preAttackRate.min )/2, allAlices.preAttackRate.max)
	fmt.Fprintf(parserOutput, "  att min %fms, mean %fms, max %fms\n", allAlices.attackRate.min,( allAlices.attackRate.max + allAlices.attackRate.min )/2, allAlices.attackRate.max)
	fmt.Fprintf(parserOutput, "  post min %fms, mean %fms, max %fms\n", allAlices.postAttackRate.min,( allAlices.postAttackRate.max + allAlices.postAttackRate.min )/2, allAlices.postAttackRate.max)

	fmt.Fprintf(parserOutput, "Response delay +%fms\n", float64(attack.st.timeUntilFirstDetectorsEnabled - attack.st.attackInitiation)/1000)
	fmt.Fprintf(parserOutput, "Attack blocked +%fms\n", float64(attack.st.timeUntilFullyResponsive - attack.st.attackInitiation)/1000)
	fmt.Fprintln(parserOutput, "Summary of Blocked Ips: ")
	for n, c := range sumIP {
		if c == 0 { continue }
		if c == 1 {
			fmt.Fprintln(parserOutput, "  - 1 blocked IP on node", n)
		} else {
			fmt.Fprintln(parserOutput, "  -", c, "blocked IPs on node", n)
		}
	}
	/////////////////////////
	min = 0
	for _, can := range canaries {
		stmp := can.serverResponsive.timestampF
		if min == 0 || min > stmp {min = stmp}
	}
	// fmt.Fprintln(parserOutput, "first canary timeout +", min - point0)
	max = 0
	for _, can := range canaries {
		stmp := can.serverResponsive.timestampF
		if max == 0 || max < stmp {max = stmp}
	}
	// fmt.Fprintln(parserOutput, "last canary timeout +", max - point0)

	min = 0
	for _, can := range canaries {
		stmp := can.serverResponsive.timestampT
		if min == 0 || min > stmp {min = stmp}
	}
	// fmt.Fprintln(parserOutput, "first canary reconnection +", min - point0)
	max = 0
	for _, can := range canaries {
		stmp := can.serverResponsive.timestampT
		if max == 0 || max < stmp {max = stmp}
	}
	// fmt.Fprintln(parserOutput, "last canary reconnection +", max - point0)
}

