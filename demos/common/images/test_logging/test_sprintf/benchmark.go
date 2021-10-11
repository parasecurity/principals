package main

import (
	"flag"
	"fmt"
	"time"
)

var (
	workload int
	wl int64
)

func init() {
	flag.IntVar(&workload, "l", 1, "number of iterations")
	flag.Parse()
	wl = int64(workload)
}

///////////////////////////
func foo(calldepth int, s string) int64 {
	te := time.Now().UnixNano()
	return te
}

func testCall(format string, v ...interface{}) int64 {
	ts := time.Now().UnixNano()
	te := foo(2, fmt.Sprintf(format, v...))
	dt := te - ts
	return dt
}


///////////////////////////
func testTime(format string, v ...interface{}) int64 {
	ts := time.Now().UnixNano()
	_ = time.Now().UnixNano()
	te := time.Now().UnixNano()
	dt := te - ts
	return dt
}

///////////////////////////
func testGoPrint(format string, v ...interface{}) int64 {
	ts := time.Now().UnixNano()
	go func(f string, vv ...interface{}) {_ = fmt.Sprintf(f, vv...)}(format, v...)
	te := time.Now().UnixNano()
	dt := te - ts
	return dt
}

///////////////////////////
func testSprintf(format string, v ...interface{}) int64 {
	ts := time.Now().UnixNano()
	_ = fmt.Sprintf(format, v...)
	te := time.Now().UnixNano()
	dt := te - ts
	return dt
}

func main() {

	var dt int64
	foo := "hostname"
	// bar := 42

	var cl int64
	var tm int64
	var gp int64
	var fp int64

	for i := 1; i <= workload; i++ {

		cl = 0
		tm = 0
		gp = 0
		fp = 0

		fmt.Printf("==== %d ====\n", i)
		for j:=0; j < i; j++ {
			dt = testCall("this is a simple %s message %d", foo, i)
			cl += dt
			dt = testTime("this is a simple %s message %d", foo, i)
			tm += dt
			dt = testGoPrint("this is a simple %s message %d", foo, i)
			gp += dt
			dt = testSprintf("this is a simple %s message %d", foo, i)
			fp += dt
		}

		fmt.Printf("  Time call: %d\n", tm/int64(i))
		fmt.Printf("  Go call: %d\n", gp/int64(i))
		fmt.Printf("  Sprint: %d\n", fp/int64(i))
		fmt.Printf("  func: %d\n\n", cl/int64(i))
	}

}
