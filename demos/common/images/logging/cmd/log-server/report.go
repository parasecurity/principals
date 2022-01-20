package main

import (
	"fmt"
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
	for _, malice := range malices {
		stmp := malice.attackRate.firstT
		if min == 0 || min > stmp{ min = stmp}
	}
	max := int64(0)
	for _, malice := range malices {
		stmp := malice.attackRate.firstT
		if max == 0 || max < stmp{ max = stmp}
	}
	// attack.st.attackInitiation = max
	// point0 := max
	// fmt.Fprintln(parserOutput, "from first malice to last: +", max-min)

	for _, malice := range malices {
		fmt.Fprintf(parserOutput, "%s bad traffic: pre %f %%(%d), att %f %%(%d)\n", 
					malice.name,
					malice.attackRate.getPercent(),malice.attackRate.packetCount,
					malice.respAttackRate.getPercent(), malice.respAttackRate.packetCount) 
	}

	for _, alice := range alices {
		fmt.Fprintf(parserOutput, "%s goodput: pre %f %%(%d), att %f %%(%d)\n", 
					alice.name,
					alice.preAttackRate.getPercent(), alice.preAttackRate.packetCount, 
					alice.attackRate.getPercent(),alice.attackRate.packetCount)
	}

	fmt.Fprintf(parserOutput, "Response delay +%fms\n", float64(attack.st.timeUntilFirstDetectorsEnabled - attack.st.attackInitiation)/1000)
	fmt.Fprintf(parserOutput, "Attack blocked +%fms\n", float64(attack.st.timeUntilFullyResponsive - attack.st.attackInitiation)/1000)
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

