package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var (
	args struct {
		server    *string
		threshold *int
		logPath   *string
	}
)

func timeGet(url string) {
	t := http.DefaultTransport.(*http.Transport).Clone()
	t.MaxIdleConns = 100
	t.MaxConnsPerHost = 100
	t.MaxIdleConnsPerHost = 100

	for {
		httpClient := &http.Client{
			Timeout:   10 * time.Second,
			Transport: t,
		}
		start := time.Now()
		r, err := httpClient.Get("http://127.0.0.1:3000/health/")
		r.Body.Close()
		interval := time.Since(start)
		log.Println("Response in :", interval)

		if err != nil {
			panic(err)
		}

		if interval > time.Duration(*args.threshold)*time.Microsecond {
			log.Println("Threshold passed:", interval)
		}
		time.Sleep(time.Second)
		httpClient.CloseIdleConnections()
	}
}

func init() {
	args.server = flag.String("conn", "http://127.0.0.1:3000/health/", "The server url e.g. http://127.0.0.1:3000/health/")
	args.threshold = flag.Int("t", 1000, "The time threshold in μs")
	args.logPath = flag.String("lp", "./canary.log", "The path to the log file")
	flag.Parse()

	// open log file
	logFile, err := os.OpenFile(*args.logPath, os.O_APPEND|os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		log.Println(err)
		return
	}

	log.SetFlags(log.Ldate | log.Ltime)
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
	timeGet(*args.server)
}
