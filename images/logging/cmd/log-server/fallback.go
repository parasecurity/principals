package main

import (
	"os"
	"time"
	"fmt"
	"os/signal"
	"syscall"
	"net"
	"log"
	"sync"
)

var (
	connWG sync.WaitGroup
	agents *connections
	stop_connections chan struct{}
	stop_listener chan struct{}
	noConns cc
)

type connections struct {
	c map[string]net.Conn
	l sync.RWMutex
}

type cc struct {
	ncon int
	m sync.Mutex
}

func (c *cc) init() {
	c.ncon = 0
}

func (c *cc) Add()  {
	c.m.Lock()	
	c.ncon++
	c.m.Unlock()	
}

func (c *cc) Done() {
	c.m.Lock()	
	c.ncon--
	c.m.Unlock()	
}

func (c *cc) hasConnections() bool {
	return c.ncon != 0
}

var ping_timeout time.Duration = 5

func (c *connections)ping() {
	for {
		c.l.Lock()
		for _, cli := range c.c {
			_, e := fmt.Fprintln(cli, "Ping")
			if e != nil {
				log.Println("Ping error", e)
			}
		}
		c.l.Unlock()
		time.Sleep(ping_timeout*time.Second)
	}
}

func init() {
	// map of all agent connections
	agents = new(connections)
	agents.c = make(map[string]net.Conn)
	go agents.ping()

	noConns.init()

	init_channels()
	// Catch all signals since not explicitly listing
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)
	// Method invoked upon seeing signal
	go func() {
		s := <-sigs
		log.Printf("RECEIVED SIGNAL: %s", s)

		log.Println("closing listnener")
		listener.Close()

		log.Println("waiting for connections to close")
		for ;noConns.hasConnections(); {
			stop_connections<- struct{}{}
			noConns.Done()
		}
		connWG.Wait()

		sortAndSend(toAnalyser, toOut, 0)
		log.Println("waiting 5 seconds before exiting")
		time.Sleep(5*time.Second)
		log.Println("Buy")
		// stop_listener<- struct{}{}
		log.Println("Exiting")
		os.Exit(1)
	}()
}

