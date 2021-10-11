package main

import (
	"flag"
	"sync"
	"sync/atomic"
	"runtime"
	"fmt"
	"log"
	"net"
	"time"
)

type my_logger struct {
	c net.Conn
	ch1 chan []byte
	ch2 chan []byte
	ch3 chan []byte
	ch4 chan []byte
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
	l.ch1 = make(chan []byte)
	l.ch2 = make(chan []byte)
	l.ch3 = make(chan []byte)
	l.ch4 = make(chan []byte)
	l.wn = 0
	l.w.Add(4)
	go l.send_logs(&l.w, l.ch1)
	go l.send_logs(&l.w, l.ch2)
	go l.send_logs(&l.w, l.ch3)
	go l.send_logs(&l.w, l.ch4)

	return l
}

func (l *my_logger) Close() error {

	/* krataw afti tin ylopoiisi. ypothetw oti 
	h Waitgroup.Wait() logika tha kanei kapoio spinlock.
	edw den kanoume spin, pame kateftheian ston scheduler
	gia na afisoume porous se oti allo trexei xwris spinning.
	den thelw na kanei spin parallila me tis alles leitourgeies
	*/

	for !atomic.CompareAndSwapInt32(&l.wn, 0, l.wn) {
		runtime.Gosched()
	}

	close(l.ch1)
	close(l.ch2)
	close(l.ch3)
	close(l.ch4)
	l.w.Wait()
	err := l.c.Close()
	return err
}

func (l *my_logger) Write(p []byte) (n int, err error) {

	atomic.AddInt32(&l.wn, 1)
	temp := make([]byte, len(p))
	copy(temp, p)

	go func (msg []byte, lg *my_logger) {
		defer atomic.AddInt32(&lg.wn, -1)
		select {
		case lg.ch1 <- msg:
		case lg.ch2 <- msg:
		case lg.ch3 <- msg:
		case lg.ch4 <- msg:
		}
	}(temp, l)

	return	len(p), nil
}

func (l my_logger) send_logs(wg * sync.WaitGroup, ch chan []byte) {
	defer wg.Done()

	for {
		msg, ok := <-ch
		if !ok {
			break
		}
		l.c.Write(msg)
	}
}

func init() {
	flag.IntVar(&workload, "l", 1, "number of thread couples to run")
	flag.Parse()

	ml = init_logger("127.0.0.1:4321")

	log.SetFlags(log.Ldate | log.Lmicroseconds | log.LUTC)
	log.SetOutput(ml)
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

	defer func(){
		wg1.Wait()
		wg2.Wait()
		ml.Close()
	}()

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


}
