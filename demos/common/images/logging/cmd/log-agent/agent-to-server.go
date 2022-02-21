package main

import (
	"net"
	"os"
	"io"
	"bufio"
	"fmt"
	"time"
	"log"
	"sync"
)

var (
	control struct {
		req chan struct{}
		ans chan struct{}
	}
	// logging server socket and address
	srv *net.TCPConn
	srvAddr *net.TCPAddr
	serverClosed bool
	serverUp bool
	srvMx sync.Mutex
)

func closeServerConnection() {
	log.Println("locking")
	srvMx.Lock()
	log.Println("locked")
	if !serverClosed {
		log.Println("servers is still open")
		serverClosed = true
		log.Println("safe closing server")
		srv.Close()
	}
	log.Println("unlocking")
	srvMx.Unlock()
	log.Println("unlocked")
}

func fixit() {
	log.Println("Fix it")
	control.req<- struct{}{}
	<-control.ans
}

var ping_timeout time.Duration = 1

func checkConnection() {
	serverUp = true
	for {
		reader := bufio.NewReader(srv)
		srv.SetReadDeadline(time.Now().Add(ping_timeout*time.Second))
		_, err := reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				log.Println("Server exited")
				fixit()
			} else if err.(net.Error).Timeout() {
				log.Println("Ping timed-out")
				if serverUp {
					log.Println("server is up", serverUp)
					serverUp = false
				} else {fixit()}
			} else {
				log.Println("CHECKER: ", err)
			}
		}else {
			serverUp = true
			log.Println("Received ping from logging server")
		}
	}
}

func stateCheck() {
	for {
		<-control.req
		_, err := fmt.Fprintf(srv, "%s agentPing\n", *args.nodeName)
		if err != nil {
			log.Println("Closing broken connection to server")
			closeServerConnection()
			err = connectToServer(0)
		}
		control.ans<- struct{}{}
	}
}

func init() {

	control.req = make(chan struct{})
	control.ans = make(chan struct{})
	serverClosed = true

	var err error
	srvAddr, err = net.ResolveTCPAddr("tcp4", *args.logServer)
	err = connectToServer(10)
	if err != nil {
		println("failed to connect to server")
		println(err)
		os.Exit(1)
	} else {
		println("connection to server established")
	}
	go stateCheck()
	go checkConnection()
}


func serverWriter(logs chan []byte){
	defer func() {
		log.Println("Server connection closed")
		closeServerConnection()
	}()

	for {
		msg := <-logs
		// TODO
		for {
			var err error
			if !serverClosed {
				_, err = fmt.Fprintf(srv, "%s %s", *args.nodeName, msg)
				log.Printf("%s %s", *args.nodeName, msg)
			} else {
				err = net.ErrClosed
			}
			if err != nil {
				log.Printf("error sending: %s ", msg)
				fixit()
			} else {
				break
			}
			log.Println("retrying")
		}
	}
}

// for negative or 0 retries loops forever
func connectToServer(retries int) (err error){
	for i := retries; i != 1; i-- {
		println("Connecting to server ", srvAddr.IP.String())
		if srv != nil {
			closeServerConnection()
		}
		srv, err = net.DialTCP("tcp4", nil, srvAddr)
		if err == nil {
			serverClosed = false
			break
		}
		println("Error", err)
		time.Sleep(1 * time.Second)
	}
	serverUp = err == nil 
	return
}

