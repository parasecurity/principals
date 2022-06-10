package main

import (
	"time"
	log "logging"
)

func main(){

	go func() {
		for i := 0; ; i++ {
			log.Println("This a log", i)
			time.Sleep(1 * time.Second)
		}
	} ()

	for i := 0; ; i++ {
		log.Println("This a log", i)
		time.Sleep(1 * time.Second)
	}

}
