package main

import (
	"fmt"
)

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
		fmt.Fprintln(parserOutput, malice, "data rate",malice.attackRate.getDataRate(), "Kbps")
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

