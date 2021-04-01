package main

import (
	"bufio"
	"flag"
	"fmt"
	"math/rand"
	"net"
	"strings"
	"sync"
	"time"
)

type connections struct {
	c []net.Conn
	l sync.RWMutex
}

var args struct {
	port *string
}

func init() {
	args.port = flag.String("port", "12345", "The server port")
	flag.Parse()
}

func handleConnection(c net.Conn, connList *connections) {
	fmt.Printf("Serving %s\n", c.RemoteAddr().String())
	for {
		netData, err := bufio.NewReader(c).ReadString('\n')
		if err != nil {
			fmt.Println(err)
			break
		}

		temp := strings.TrimSpace(string(netData))
		if temp == "STOP" {
			break
		}
		connList.l.RLock()
		for _, conn := range connList.c {
			conn.Write([]byte(netData))
		}
		connList.l.RUnlock()
	}
	c.Close()
	fmt.Printf("Connection closed %s\n", c.RemoteAddr().String())
	for idx, conn := range connList.c {
		if conn == c {
			fmt.Printf("Removing %s from list\n", c.RemoteAddr().String())
			connList.l.Lock()
			connList.c[idx] = connList.c[len(connList.c)-1]
			connList.c = connList.c[:len(connList.c)-1]
			connList.l.Unlock()
		}
	}
}

func main() {
	url := ":" + *args.port
	l, err := net.Listen("tcp4", url)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer l.Close()
	rand.Seed(time.Now().Unix())
	connList := new(connections)
	for {
		c, err := l.Accept()
		if err != nil {
			fmt.Println(err)
			return
		}
		connList.l.Lock()
		connList.c = append(connList.c, c)
		connList.l.Unlock()
		go handleConnection(c, connList)
	}
}
