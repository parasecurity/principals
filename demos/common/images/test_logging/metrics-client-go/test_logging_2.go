package main

import (
	"sync"
	"fmt"
	"log"
	"net"
	"time"
)

type my_logger struct {
	c net.Conn
	ch chan []byte
	m sync.RWMutex
	w sync.WaitGroup
}

var (
	wg sync.WaitGroup
)

func init_logger(ip string) (*my_logger, error) {
	l := new(my_logger)
	// TODO try connect, timeout
	c, err := net.Dial("tcp", ip)

	if err == nil {
		l.c = c
	}
	l.ch = make(chan []byte)
	//go l.send_logs()

	return l, err
}

func (l my_logger) Write(p []byte) (n int, err error) {

	// l.m.Lock()
	l.c.Write(p)
	// l.m.Unlock()
	return	len(p), nil
}

func (l my_logger) send_logs() {

	for {
		l.m.RLock()
		msg := <-l.ch
		l.m.RUnlock()
		l.c.Write(msg)
	}
}

func init() {

	ml , _:= init_logger("127.0.0.1:4321")

	log.SetFlags(log.Ldate | log.Lmicroseconds | log.LUTC)
	log.SetOutput(ml)
	wg = ml.w
}

func logging_machine_formating(wg * sync.WaitGroup, name string, server_ip string){
	defer wg.Done()

	ts := time.Now().UnixNano()
	log.Printf("%s action1", name)
	log.Printf("%s action2", name)
	log.Printf("%s action3", name)
	log.Printf("%s action4", name)
	log.Printf("%s action5", name)
	te := time.Now().UnixNano()
	dt := te - ts
	log.Printf("%d format", dt)
}

func logging_machine(wg * sync.WaitGroup, name string, server_ip string) {
	defer wg.Done()

	ts := time.Now().UnixNano()
	log.Println("action1")
	log.Println("action2")
	log.Println("action3")
	log.Println("action4")
	log.Println("action5")
	te := time.Now().UnixNano()
	dt := te - ts
	log.Printf("%d raw_string", dt)
}

func main() {
	var wg1 sync.WaitGroup
	var wg2 sync.WaitGroup

	var name string
	workload := 3

	wg1.Add(workload)
	for i := 0; i < workload; i++ {
		name = fmt.Sprintf("tester_%d", i)
		go logging_machine(&wg1, name, "127.0.0.1")
	}

	wg2.Add(workload)
	for i := 0; i < workload; i++ {
		name = fmt.Sprintf("tester_%d", i)
		go logging_machine_formating(&wg2, name, "127.0.0.1")
	}

	wg1.Wait()
	wg2.Wait()
	// wg.Wait()
	// time.Sleep(5*time.Second)

}
