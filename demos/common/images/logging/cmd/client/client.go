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

	ts = time.Now().UnixNano()
	log.Println("a log println with newline\n")
	te = time.Now().UnixNano()
	dt = te - ts
	fmt.Printf("dummy send for begining ln %d\n", dt/1000)

	ts = time.Now().UnixNano()
	log.Printf("this is a log %d %s ", n, msg)
	te = time.Now().UnixNano()
	dt = te - ts
	fmt.Printf("dummy send for begining f %d\n", dt/1000)

	ts = time.Now().UnixNano()
	log.Printf("this is a log %d %s \n", n, msg)
	te = time.Now().UnixNano()
	dt = te - ts
	fmt.Printf("dummy send for begining %d f newline\n", dt/1000)

	ts = time.Now().UnixNano()
	log.Print("just print", "another printing print\n")
	te = time.Now().UnixNano()
	dt = te - ts
	fmt.Printf("dummy send for begining %d\n", dt/1000)

	ts = time.Now().UnixNano()
	log.Print("just print\n")
	te = time.Now().UnixNano()
	dt = te - ts
	fmt.Printf("dummy send for begining newline %d\n", dt/1000)

}
