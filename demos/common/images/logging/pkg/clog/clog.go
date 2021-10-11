package clog

import (
	"net"
	"sync"
	"os"
	"fmt"
	"time"
	"strings"
)

//////////////////////////////////
// helper utils

func debug(format string, args ...interface{}){
	fmt.Fprintf(os.Stderr, format, args...)
}

// 1633745232195634 this is a timestamp in microseconds
// 16 decimal digits long for the next 266 years

type buffer struct {
	sb strings.Builder
	next *buffer
}

const digitMap = "0123456789"

// formats a timestamp to a 16 digit decimal
// string. Used by println and print families
// but not by printf
// func (b * timeBuff)digitf(d int64) string {
// 	b.buff[16] = 0
// 	for i := 15; i >= 0; i-- {
// 		b.buff[i] = digitMap[d%10]
// 		d /= 10
// 	}
// 	return *(*string)(unsafe.Pointer(&b.buff))
// }

// buffer freelist 
var pool *buffer
var pmx sync.Mutex

var control chan struct{}

// populates the buffer pool 
// can be stopped anytime via 
// a write on ctrl channel
// func maker(ctrl chan struct{}){
// 	for i := 0; true; i++ {
// 		select {
// 		case <-ctrl:
// 			return
// 		default:
// 			if i < 10 {
// 				b := new(timeBuff)
// 				b.buff = make([]byte, 17)
// 				pmx.Lock()
// 				b.next = pool
// 				pool = b
// 				pmx.Unlock()
// 			} else {
// 				runtime.Gosched()
// 			}
// 		}
// 	}
// }

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

	// TODO build them in one string for efficiency
	host string
	command string
}


func (l logging) Write(p []byte) (n int, err error) {

	// TODO error handling and fall-back in disconnections
	if p[len(p) - 1] == p[len(p) - 2]{
		l.conn.Write(p[:len(p) - 1])
	} else {
		l.conn.Write(p)
	}

	return len(p), nil

}

var log logging

func init (){

	control = make(chan struct{})
	// go maker(control)

	network := "unixpacket"
	path := "/tmp/testlog.sock"
	//TODO error handling
	var err error
	log.sock_addr, err = net.ResolveUnixAddr(network, path)
	log.conn, err = net.DialUnix(network, nil, log.sock_addr)
	log.host, err = os.Hostname()
	log.command = os.Args[0]
	err = err // to avoid error 

	// control<- struct{}{}
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

func Println(args ...interface{}){
	log.println(args...)
}

func Printf(format string, args ...interface{}){
	log.printf(format, args...)
}

func Panic(args ...interface{}){
	log.print(args...)
}

func Panicln(args ...interface{}){
	log.println(args...)
}

func Panicf(format string, args ...interface{}){
	log.printf(format, args...)
}

func Fatal(args ...interface{}){
	log.print(args...)
}

func Fatalln(args ...interface{}){
	log.println(args...)
}

func Fatalf(format string, args ...interface{}){
	log.printf(format, args...)
}

