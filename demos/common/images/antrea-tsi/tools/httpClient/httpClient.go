package main

import (
	"flag"
	"io/ioutil"
	log "logging"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

var (
	args struct {
		server  string
		sleep   int
		timeout int
		clients int
		logPath string
	}
)

func init() {
	flag.StringVar(&args.server, "conn", "http://localhost:8080/", "The server connection in format ip:port (default http://localhost:8080/)")
	flag.IntVar(&args.sleep, "s", 1000, "sleep in ms (default 1000)")
	flag.IntVar(&args.timeout, "t", 1000, "timeout in ms (default 1000)")
	flag.IntVar(&args.clients, "c", 1, "number of concurrent clients (default 1)")
	flag.StringVar(&args.logPath, "lp", "./client.log", "The path to the log file (default ./client.log)")
	flag.Parse()

	// open log file
	logFile, err := os.OpenFile(args.logPath, os.O_APPEND|os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		log.Println(err)
		return
	}

	log.SetFlags(log.Ldate | log.Lmicroseconds | log.LUTC)
	log.SetOutput(logFile)

	// setup signal catching
	sigs := make(chan os.Signal, 1)
	// catch all signals since not explicitly listing
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)
	// method invoked upon seeing signal
	go func() {
		s := <-sigs
		log.Printf("RECEIVED SIGNAL: %s", s)
		os.Exit(1)
	}()
}

func main() {
	t := http.DefaultTransport.(*http.Transport).Clone()
	t.MaxIdleConns = 100
	t.MaxConnsPerHost = 100
	t.MaxIdleConnsPerHost = 100

	var wg sync.WaitGroup

	for c := 0; c < args.clients; c++ {
		wg.Add(1)
		go func(c int, wg *sync.WaitGroup) {
			// this code will never run
			// TODO clean up
			defer func() {
				if r := recover(); r != nil {
					log.Println("client ", c, " closed. Error: ", r)
				}
			}()
			defer wg.Done()
			httpClient := &http.Client{
				Timeout:   time.Duration(args.timeout) * time.Millisecond,
				Transport: t,
			}

			var count uint64
			count = 0
			for {
				resp, err := httpClient.Get(args.server)
				if err != nil {
					log.Print(err)
					continue
				}
				defer resp.Body.Close()

				log.Println("Response status:", resp.Status)

				bytes, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					log.Println(err)
					return
				}
				count++
				log.Println("Transfers: ", count, ", bytes: ", len(bytes))
				time.Sleep(time.Duration(args.sleep) * time.Millisecond)
			}
		}(c, &wg)
	}
	wg.Wait()
}
