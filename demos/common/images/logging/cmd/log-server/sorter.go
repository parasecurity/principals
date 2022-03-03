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
	sorter<- struct{}{}

	// log.Println("first ", firstCached, " last ", lastCached, " limit ", limit)
	if !isInOrder {
		log.Println("...sorting...")
		sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })
		isInOrder = true
		log.Println("...done")
	}

	del := 0
	for i, k := range keys {
		if limit != 0 && k >= limit {
			del = i
			break
		}
		del++
		logEntry := logBuf[k]
		log.Println("Internal send: ", logEntry)
		// select for efficiency
		select {
		case a <- logEntry:
			b <- logEntry
		case b <- logEntry:
			a <- logEntry
		}
		// delete(logBuf, k)
	}
	if del == 0 {
		log.Print("no logs were sent")
		keys = keys[:0]
	} else {
		log.Print(del, " logs were sent")
		keys = keys[del:]
	}
	if len(keys) > 0 {
		firstCached = keys[0]
		lastCached = keys[len(keys) - 1]
	} else {
		firstCached = 0
		lastCached = 0
	}
	<-sorter
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
			if lastCached == 0 { lastCached = timestamp }
			if timestamp < lastCached { 
				isInOrder = false
			} else {
				mean := (lastCached + firstCached)/2
				if timestamp - mean > 5 * tsiSecond { 
					sortAndSend(analyser, printer, mean)
					log.Println("Sorting by time cached")
				} else if len(logBuf) % 512 == 0 {
					// NOTE possible change threshold to mean
					sortAndSend(analyser, printer, lastCached - 5 * tsiSecond)
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
