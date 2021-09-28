package main

import (
	"flag"
	"sync"
	// "sync/atomic"
	// "runtime"
	"fmt"
	"log"
	"net"
	"time"
)

type my_logger struct {
	c net.Conn
	ch chan []byte
	m sync.Mutex
	mw sync.Mutex
	w sync.WaitGroup
	wn int32
}

var (
	workload int
	wg sync.WaitGroup
	ml *my_logger
)

func init_logger(ip string) (*my_logger) {
	l := new(my_logger)
	// TODO try connect, timeout
	c, err := net.Dial("tcp", ip)

	if err == nil {
		l.c = c
	}
	l.ch = make(chan []byte)
	l.wn = 0
	//go l.send_logs()

	return l
}

func (l *my_logger) Close() error {

	l.w.Wait()
	// for !atomic.CompareAndSwapInt32(&l.wn, 0, l.wn) {
	// 	runtime.Gosched()
	// }
	// err := l.c.Close()
	// return err
	return nil
}

func (l *my_logger) Write(p []byte) (n int, err error) {
	// l.m.Lock()
	// l.ch <- p
	// l.m.Unlock()

	// l.mw.Lock()
	l.w.Add(1)
	// l.mw.Unlock()
	// atomic.AddInt32(&l.wn, 1)
	temp := make([]byte, len(p))
	copy(temp, p)

	go func (msg []byte, lg *my_logger) {
		defer lg.w.Done()
		// defer atomic.AddInt32(&lg.wn, -1)
		lg.m.Lock()
		lg.c.Write(msg)
		lg.m.Unlock()
	}(temp, l)

	return	len(p), nil
}

func (l my_logger) send_logs() {

	for {
		l.m.Lock()
		msg := <-l.ch
		l.m.Unlock()
		l.c.Write(msg)
	}
}

func init() {
	flag.IntVar(&workload, "l", 1, "number of thread couples to run")
	flag.Parse()

	ml = init_logger("127.0.0.1:4321")

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
	ml.Close()
	// wg.Wait()
	// time.Sleep(5*time.Second)

}
