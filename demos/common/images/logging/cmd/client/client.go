package main

import (
	"time"
	log "logging/pkg/clog"
)

func main(){
	for i := 0; ; i++ {
		log.Println("This a log", i)
		time.Sleep(2 * time.Second)
	}
}
