package main

import (
	"os"
	"encoding/binary"
	"bytes"
	"sync"
	"fmt"
	"net"
	"time"
)

type delay_log struct {
	conn net.Conn
	id string
}

var (
	logger *delay_log
)

func new_delay_log(id string, server_ip string) *delay_log {
	dl := new(delay_log)
	c, err := net.Dial("tcp", server_ip)
	if err != nil {
		os.Exit(1)
	}
	dl.conn = c
	dl.id = id
	return dl
}

func (dl delay_log) stamp(msg string) {
	//vers0.2
	t := time.Now().UnixNano()

	var buff bytes.Buffer
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(t))
	buff.Write(b)
	buff.WriteString(" ")
	buff.WriteString(dl.id)
	buff.WriteString(" ")
	buff.WriteString(msg)
	buff.WriteString("\n")

	dl.conn.Write(buff.Bytes())
}

func logging_machine(wg * sync.WaitGroup, name string, server_ip string) {
	defer wg.Done()

	log := new_delay_log(name, server_ip)
	for i := 0; i < 3; i++ {
		log.stamp("action")
	}
	log.conn.Close()
}

func main() {
	var wg sync.WaitGroup
	var name string
	wg.Add(3)
	for i := 0; i < 3; i++ {
		name = fmt.Sprintf("tester_%d", i)
		go logging_machine(&wg, name, "127.0.0.1")
	}
	wg.Wait()

}
