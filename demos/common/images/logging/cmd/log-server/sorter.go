package main

import (
	"sort"
	"strconv"
	"strings"
)

func sortLogs(logs chan []byte,analyser , printer chan []string){
	logBuf := make(map[int64][]string)
	last := int64(0)
	first := int64(0)
	// lastOut := int64(0)
	for {
		// log deconstruction
		msg := <-logs
		// print(string(msg))
		toks := strings.SplitN(string(msg), " ", 5)
		//node := toks[0]
		//pod := toks[1]
		//cmd := toks[2]
		//log := toks[4]
		// OPT no error checking bellow! for it is ok

		timestamp, _ := strconv.ParseInt(toks[3], 10, 64)
		logBuf[timestamp] = toks
		if last > timestamp { last = timestamp }
		if first == 0 { first = timestamp }
		// if last < timestamp { lastOut = timestamp }

		// send them in 10s
		if len(logBuf) > 9 {
			
			keys := make([]int64, 0, len(logBuf))
			for k := range logBuf {
				keys = append(keys, k)
			}
			sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })

			for _, k := range keys {

				log := logBuf[k]
				// select for efficiency
				select {
				case printer <- log:
					analyser <- log
				case analyser <- log:
					printer <- log
				}
				delete(logBuf, k)
			}
		}
	}
}
