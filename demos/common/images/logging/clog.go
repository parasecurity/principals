package logging

import (
	"net"
	"sync"
	"os"
	"io"
	"fmt"
	"time"
	"strings"
)

//////////////////////////////////
// helper utils

func debug(format string, args ...interface{}){
	fmt.Fprintf(os.Stderr, format, args...)
}

type buffer struct {
	sb strings.Builder
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

	conn *net.UnixConn
	sock_addr *net.UnixAddr

	host string
	command string
}

// TODO named return values?
func (l logging) Write(p []byte) (n int, err error) {

	// TODO error handling and fall-back in disconnections
	if p[len(p) - 1] == p[len(p) - 2]{
		_, err = l.conn.Write(p[:len(p) - 1])
	} else {
		_, err = l.conn.Write(p)
	}
	if err != nil {
		// do something or suppose agent doesn't die
		panic("agent down! agent down!")
	}

	return len(p), err

}

var log logging

// for negative or 0 retries loops forever
func connectToAgent(retries int) (err error){
	for i := retries; i != 1; i-- {
		log.conn, err = net.DialUnix("unixpacket", nil, log.sock_addr)
		if err == nil {
			break
		}
		// TODO log the error maybe
		time.Sleep(1 * time.Second)
	}
	return
}

func init (){

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

}

func ( *logging) printf(format string, args ...interface{}) {

	st := append([]interface{}{log.host, log.command, time.Now().UnixNano()/1000}, args...)
	b := getBuff()
	b.sb.WriteString("%s %s %d ")
	b.sb.WriteString(format)
	if ! strings.HasSuffix(b.sb.String(), "\n") {
		b.sb.WriteString("\n")
	}

	fmt.Fprintf(log, b.sb.String(), st...)
	putBuff(b)
}

func ( *logging ) println(args ...interface{}) {

	st := append([]interface{}{log.host, log.command, time.Now().UnixNano()/1000}, args...)
	fmt.Fprintln(log, st...)
}

func ( *logging) print(args ...interface{}){

	st := append([]interface{}{log.host, " ", log.command, " ", time.Now().UnixNano()/1000, " "}, args...)
	st = append(st, "\n")
	fmt.Fprint(log, st...)
}

func Print(args ...interface{}){
	log.print(args...)
}

////////////////////////////////////////
// public 
func Println(args ...interface{}){
	log.println(args...)
}

func Printf(format string, args ...interface{}){
	log.printf(format, args...)
}

func Panic(args ...interface{}){
	log.print(args...)
	s := fmt.Sprint(args...)
	panic(s)
}

func Panicln(args ...interface{}){
	log.println(args...)
	s := fmt.Sprintln(args...)
	panic(s)
}

func Panicf(format string, args ...interface{}){
	log.printf(format, args...)
	s := fmt.Sprintf(format, args...)
	panic(s)
}

func Fatal(args ...interface{}){
	log.print(args...)
	os.Exit(1)
}

func Fatalln(args ...interface{}){
	log.println(args...)
	os.Exit(1)
}

func Fatalf(format string, args ...interface{}){
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
	// TODO 
	// nothing todo here
	// standard logger does not close the writer set by the programmer
	// whoever calls SetOutput is responsible of closing the corresponding writer
	// or hope GC will do its job correctly and quickly
	return
}
