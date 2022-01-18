package main

import (
	"sort"
	"strconv"
	"strings"
	"time"
	"log"
)

var (
	logBuf map[int64][]string // map with timestamps : decosctructed logs
	keys []int64	// slice with cached timestamps
	isInOrder bool  // is cache in order
	lastCached int64
	firstCached int64
)

const (
	tsiSecond = int64(1000000) // a second in tsi dialect
)

func init() {
	logBuf = make(map[int64][]string)
	keys = make([]int64, 0, 32)
	isInOrder = true
	firstCached = 0
	lastCached = 0
}

// sorts logBuf and sends all logs until limit time. 
// if limit is 0 then it sends every entry in logBuf
func sortAndSend(a, b chan []string, limit int64) {

	if !isInOrder {
		sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })
		isInOrder = true
	}

	del := 0
	for i, k := range keys {
		if limit != 0 && k >= limit {
			del = i
			break
		}
		log := logBuf[k]
		// select for efficiency
		select {
		case a <- log:
			b <- log
		case b <- log:
			a <- log
		}
		delete(logBuf, k)
	}
	if del == 0 {
		keys = keys[:0]
	} else {
		keys = keys[del:]
	}
	if len(keys) > 0 {
		firstCached = keys[0]
		lastCached = keys[len(keys) - 1]
	} else {
		firstCached = 0
		lastCached = 0
	}
}

func sortLogs(logs chan []byte, analyser, printer chan []string){
	for {
		select {
		case msg := <-logs:
			// log deconstruction
			toks := strings.SplitN(string(msg), " ", 5)
			// TODO error checking
			timestamp, _ := strconv.ParseInt(toks[3], 10, 64)
			keys = append(keys, timestamp)
			logBuf[timestamp] = toks

			if firstCached == 0 { firstCached = timestamp }
			if timestamp < lastCached { 
				isInOrder = false
			} else {
				if timestamp - (lastCached + firstCached)/2 > 5 * tsiSecond { 
					sortAndSend(analyser, printer, timestamp - 5 * tsiSecond)
					log.Println("Sorting by time cached")
				} else if len(logBuf) > 600 {
					sortAndSend(analyser, printer, lastCached - 7 * tsiSecond)
					log.Println("Sorting by number of logs cached")
				} else {
					lastCached = timestamp
				}
			}
		case <- time.After(5 * time.Second):
			if lastCached != 0 {
				sortAndSend(analyser, printer, lastCached - 2 * tsiSecond)
				log.Println("Sorting by timeout")
			}
		}
	}
}
