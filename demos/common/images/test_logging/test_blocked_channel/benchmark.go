package main

import (
	"flag"
	"sync"
	"net"
	"fmt"
	"time"
	"os"
)

var (
	workload int
	wl int64
	f *os.File
	buf []byte
)

func init() {
	flag.IntVar(&workload, "l", 1, "number of iterations")
	flag.Parse()
	wl = int64(workload)

	f, _ = os.OpenFile("logs.log", os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
}

// Cheap integer to fixed-width decimal ASCII. Give a negative width to avoid zero-padding.
func itoa(buf *[]byte, i int64, wid int) {
	// Assemble decimal in reverse order.
	var b [20]byte
	bp := len(b) - 1
	for i >= 10 || wid > 1 {
		wid--
		q := i / 10
		b[bp] = byte('0' + i - q*10)
		bp--
		i = q
	}
	// i < 10
	b[bp] = byte('0' + i)
	*buf = append(*buf, b[bp:]...)
}

///////////////////////////
func testLogFormat(format string, v ...interface{}) (int64, int64, int64) {
	now := time.Now().UnixNano()
	host := "hostname_long_pod_name"

	//format the message
	ts := time.Now().UnixNano()
	msg := fmt.Sprintf(format, v...)
	te := time.Now().UnixNano()
	firstdt := te - ts

	//1 Sprintf all
	ts = time.Now().UnixNano()
	entry := fmt.Sprintf("%d %s %s\n", now, host, msg)
	te = time.Now().UnixNano()
	spdt := te - ts

	//2 std log style
	ts = time.Now().UnixNano()

	buf = buf[:0]
	buf = append(buf, host...)
	itoa(&buf, now, -1)
	buf = append(buf, msg...)
	buf = append(buf, '\n')

	te = time.Now().UnixNano()
	logdt := te - ts

	fmt.Fprintf(os.Stderr,entry)
	fmt.Fprintf(os.Stderr,"lalala\n")

	os.Stderr.Write(buf)

	return firstdt, spdt, logdt
}

func reader(w * sync.WaitGroup, ch chan []byte){
	data := <-ch
	data = append(data,' ')
	w.Done()
}

func testWrite() (int64, int64, int64, int64){
	// dn, _ := os.OpenFile("/dev/null", os.O_APPEND, 0644)
	c, err := net.Dial("unix", "/tmp/fastlog.sock")
	if err != nil {
		// fmt.Println("error dialing server")
		return -1, -1, -1, -1
	}
	defer c.Close()

	kati := "akjfhgkjbjadnlkafjgbjkdfah\n"

	var buff []byte
	buff = append(buff, kati...)

	ts := time.Now().UnixNano()
	f.Write(buff)
	te := time.Now().UnixNano()
	simpwr := te - ts

	ch := make(chan []byte, 1)

	var wg1 sync.WaitGroup
	wg1.Add(1)
	go reader(&wg1, ch)

	ts = time.Now().UnixNano()
	ch <- buff
	te = time.Now().UnixNano()
	chanwr := te - ts
	wg1.Wait()

	ts = time.Now().UnixNano()
	c.Write(buff)
	te = time.Now().UnixNano()
	sockwr := te - ts

	ts = time.Now().UnixNano()
	go c.Write(buff)
	te = time.Now().UnixNano()
	gowr := te - ts

	return simpwr, chanwr, sockwr, gowr
}

func main() {

	for i := 1; i <= workload; i++ {
		fmt.Printf("==== %d ====\n", i)

		f := int64(0)
		l := int64(0)
		s := int64(0)
		g := int64(0)

		for j:=0; j < i; j++ {
			dt1, dt2, dt3 := testLogFormat("this is a simple message %d",  i)
			f += dt1
			s += dt2
			l += dt3
		}

		fmt.Printf("  First format: %d\n", f/int64(i))
		fmt.Printf("  Log style: %d\n", l/int64(i))
		fmt.Printf("  Sprintf all: %d\n", s/int64(i))

		f = 0
		l = 0
		s = 0
		g = 0

		flag := 0
		for j:=0; j < i; j++ {
			dt1, dt2, dt3, dt4 := testWrite()
			if dt1 == -1 {
				flag = 1
				break
			}
			f += dt1
			s += dt2
			l += dt3
			g += dt4
		}
		if flag == 1 {
			break
		}


		fmt.Printf("  File write: %d\n", f/int64(i))
		fmt.Printf("  Channel write: %d\n", s/int64(i))
		fmt.Printf("  Socket write: %d\n", l/int64(i))
		fmt.Printf("  Go call write: %d\n", g/int64(i))
	}

}
