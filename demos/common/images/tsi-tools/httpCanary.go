package main

import (
	"flag"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var (
	server     *string
	api        *string
	ca         *string
	crt        *string
	key        *string
	threshold  *int
	failures   *int
	logPath    *string
	detectorUp bool = false
)

func connectTCP() net.Conn {
	addr := *api

	// Connect to the tls server
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		log.Println("Failed to connect: %s", err.Error())
	}
	return conn
}

func createDetector() {
	if detectorUp == true {
		return
	}

	conn := connectTCP()
	defer conn.Close()

	command := "create detector"
	_, err := conn.Write([]byte(command))
	if err != nil {
		log.Println(err)
	}
	detectorUp = true
	log.Println("Deploying detectors...")
}

func timeGet(url string) {
	t := http.DefaultTransport.(*http.Transport).Clone()
	t.MaxIdleConns = 100
	t.MaxConnsPerHost = 100
	t.MaxIdleConnsPerHost = 100
	var failureCount int
	failureCount = 0
	for {
		defer func() {
			if r := recover(); r != nil {
				log.Println("Canary connection timeout")
				if failureCount >= *failures {
					createDetector()
					failureCount = 0
				} else {
					failureCount++
				}
			}
		}()
		httpClient := &http.Client{
			Timeout:   1 * time.Second,
			Transport: t,
		}
		start := time.Now()
		r, err := httpClient.Get(*server)
		r.Body.Close()
		interval := time.Since(start)
		log.Println("Response in :", interval)

		if err != nil {
			log.Println(err)
		}

		if interval > time.Duration(*threshold)*time.Millisecond {
			log.Println("Threshold passed:", interval)
			if failureCount >= *failures {
				createDetector()
				failureCount = 0
			} else {
				failureCount++
			}
		}
		httpClient.CloseIdleConnections()
		time.Sleep(time.Second)
	}
}

func init() {
	server = flag.String("conn", "http://147.27.39.116:8080/health/", "The server url e.g. http://147.27.39.116:8080/health/")
	api = flag.String("api", "10.244.0.9:8001", "The API server url e.g. 10.244.0.9:8001")
	threshold = flag.Int("t", 1000, "The time threshold in ms")
	failures = flag.Int("f", 4, "The number of failures before we spawn a detector")
	logPath = flag.String("lp", "./canary.log", "The path to the log file")
	flag.Parse()

	// open log file
	logFile, err := os.OpenFile(*logPath, os.O_APPEND|os.O_RDWR|os.O_CREATE, 0644)
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
		log.Println("RECEIVED SIGNAL:", s)
		os.Exit(1)
	}()
}

func main() {
	for {
		timeGet(*server)
	}
}
