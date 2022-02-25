package logging

import (
	"fmt"
	"io"
	"errors"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

//////////////////////////////////
// helper utils

func debug(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format, args...)
}

type buffer struct {
	sb   strings.Builder
	next *buffer
}

// buffer freelist
var pool *buffer
var pmx sync.Mutex

// returns a buffer from pool
func getBuff() *buffer {
	pmx.Lock()
	b := pool
	if b != nil {
		pool = b.next
	}
	pmx.Unlock()
	if b == nil {
		b = new(buffer)
	}
	return b
}

// puts a buffer back to pool
func putBuff(b *buffer) {
	pmx.Lock()
	b.next = pool
	pool = b
	b.sb.Reset()
	pmx.Unlock()
}

///////////////////////////////////////
// logging struct
type logging struct {
	conn      *net.UnixConn
	sock_addr *net.UnixAddr

	host    string
	command string
}

var missedLogs int = 0

func (l logging) Write(p []byte) (n int, err error) {

	// TODO error handling and fall-back in disconnections
	n = len(p)
	np := n
	if p[n-1] == p[n-2] {
		np = n-1
	} 
	if l.conn == nil { 
		err = net.ErrClosed 
	} else {
		err = l.conn.SetWriteDeadline(time.Now().Add(time.Second))
	}

	if err == nil {
		_, err = l.conn.Write(p[:np])
	}

	if err != nil {
		print(string(p[:np]))
		missedLogs++
		if !err.(net.Error).Timeout() {
			// reconnect to agent
			fixit()
		}
		// panic("agent down! agent down!")
	} else { 
		if missedLogs != 0 {
			st := fmt.Sprintf("%s %s %d at least %d logs were missed\n", log.host, log.command, 
								time.Now().UnixNano() / 1000, missedLogs)
			l.conn.Write([]byte(st))
		}
		missedLogs = 0 
	}

	return
}

var log logging

// for negative or 0 retries loops forever
func connectToAgent(retries int) (err error) {
	println("Connecting to logging agent")
	for i := retries; i != 1; i-- {
		log.conn, err = net.DialUnix("unixpacket", nil, log.sock_addr)
		if err == nil {
			break
		} else {
			var enof *net.OpError
			if errors.As(err, &enof) {
				if !errors.Is(enof.Err, os.ErrNotExist) {
					// discard printing file not fount for closed socket - not interesting
					println(err.Error())
				}
			}
		}
		// TODO log the error maybe
		time.Sleep(1 * time.Second)
	}
	return
}

var control struct {
	req chan struct{}
	ans chan struct{}
}

func fixit() {
	agMx.Lock()
	if !tryingToConnect{
		agMx.Unlock()
		control.req<- struct{}{}
		<-control.ans
	} else { agMx.Unlock() }
}

var agMx sync.Mutex
var tryingToConnect bool = false

func stateCheck() {
	for {
		<-control.req
		control.ans<- struct{}{}
		agMx.Lock()
		if !tryingToConnect { 
			tryingToConnect = true
			agMx.Unlock()
			connectToAgent(0)
			agMx.Lock()
			tryingToConnect=false
		}
		agMx.Unlock()
	}
}

func init() {

	control.req = make(chan struct{})
	control.ans = make(chan struct{})
	network := "unixpacket"
	path := "/tmp/testlog.sock"
	//TODO error handling
	var err error
	log.sock_addr, err = net.ResolveUnixAddr(network, path)
	err = connectToAgent(20)
	if err != nil {
		// do something or suppose agent doesn't die
		panic("agent down! agent down!")
	}
	log.host, err = os.Hostname()
	log.command = os.Args[0]
	go stateCheck()

}

func (*logging) printf(format string, args ...interface{}) {

	st := append([]interface{}{log.host, log.command, time.Now().UnixNano() / 1000}, args...)
	b := getBuff()
	b.sb.WriteString("%s %s %d ")
	b.sb.WriteString(format)
	if !strings.HasSuffix(b.sb.String(), "\n") {
		b.sb.WriteString("\n")
	}

	fmt.Fprintf(log, b.sb.String(), st...)
	putBuff(b)
}

func (*logging) println(args ...interface{}) {

	st := append([]interface{}{log.host, log.command, time.Now().UnixNano() / 1000}, args...)
	fmt.Fprintln(log, st...)
}

func (*logging) print(args ...interface{}) {

	st := append([]interface{}{log.host, " ", log.command, " ", time.Now().UnixNano() / 1000, " "}, args...)
	st = append(st, "\n")
	fmt.Fprint(log, st...)
}

func Print(args ...interface{}) {
	log.print(args...)
}

////////////////////////////////////////
// public
func Println(args ...interface{}) {
	log.println(args...)
}

func Printf(format string, args ...interface{}) {
	log.printf(format, args...)
}

func Panic(args ...interface{}) {
	log.print(args...)
	s := fmt.Sprint(args...)
	panic(s)
}

func Panicln(args ...interface{}) {
	log.println(args...)
	s := fmt.Sprintln(args...)
	panic(s)
}

func Panicf(format string, args ...interface{}) {
	log.printf(format, args...)
	s := fmt.Sprintf(format, args...)
	panic(s)
}

func Fatal(args ...interface{}) {
	log.print(args...)
	os.Exit(1)
}

func Fatalln(args ...interface{}) {
	log.println(args...)
	os.Exit(1)
}

func Fatalf(format string, args ...interface{}) {
	log.printf(format, args...)
	os.Exit(1)
}

//////////////////
// dummy calls for instant compatibility with standard logger
// only calls used by tsi go code are here

const (
	Ldate         = 1 << iota     // the date in the local time zone: 2009/01/23
	Ltime                         // the time in the local time zone: 01:23:23
	Lmicroseconds                 // microsecond resolution: 01:23:23.123123.  assumes Ltime.
	Llongfile                     // full file name and line number: /a/b/c/d.go:23
	Lshortfile                    // final file name element and line number: d.go:23. overrides Llongfile
	LUTC                          // if Ldate or Ltime is set, use UTC rather than the local time zone
	Lmsgprefix                    // move the "prefix" from the beginning of the line to before the message
	LstdFlags     = Ldate | Ltime // initial values for the standard logger
)

func SetFlags(flag int) {
	return
}

func SetOutput(w io.Writer) {
	// nothing todo here
	// standard logger does not close the writer set by the programmer anyway
	// whoever calls SetOutput is responsible of closing the corresponding writer
	// or hope GC will do its job correctly and quickly (which is not the case)
	return
}
