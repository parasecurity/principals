package main

import (
	"fmt"
	"time"
	log "logging/pkg/clog"
)

func main(){

	format := "this is a log "
	n := 42
	msg := "lalala"


	ts := time.Now().UnixNano()
	log.Println(format, n, msg)
	te := time.Now().UnixNano()
	dt := te - ts
	fmt.Printf("dummy send for begining ln %d\n", dt/1000)

	ts = time.Now().UnixNano()
	log.Println(format, n, msg)
	te = time.Now().UnixNano()
	dt = te - ts
	fmt.Printf("dummy send for begining ln %d\n", dt/1000)

	for i:=0; i<1000; i++{

		go func(i int){
		for {

			ts := time.Now().UnixNano()
			log.Printf(" spaaaAAA %d aAaAAaam", i)
			te = time.Now().UnixNano()
			dt := te - ts
			fmt.Printf("spam %d\n", dt/1000)
		}}(i)
	}
	for {

		ts = time.Now().UnixNano()
		log.Printf(" main thread spaaaaaam %d ", 420)
		te = time.Now().UnixNano()
		dt = te - ts
		fmt.Printf("spam %d\n", dt/1000)
	}

}
