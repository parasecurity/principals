package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"api/pkg/utils"
)

var (
	listen    *string
	listenTCP *string
	ca        *string
	crt       *string
	key       *string
	logPath   *string
)

func init() {
	listen = flag.String("l", "localhost:8000", "The server url to listen e.g. localhost:8000")
	listenTCP = flag.String("ll", "10.244.0.9:8001", "The ip to listen requests from inside the cluster")
	ca = flag.String("ca", "./internal/ca.crt", "The file path to ca certificate e.g. ./ca.crt")
	crt = flag.String("crt", "./internal/server.crt", "The file path to server crt certificate e.g. ./server.crt")
	key = flag.String("key", "./internal/server.key", "The file path to server key e.g. ./server.key")
	logPath = flag.String("lp", "./server.log", "The path to the log file e.g. ./server.log")
	flag.Parse()

	// Open log file
	logFile, err := os.OpenFile(*logPath, os.O_APPEND|os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		log.Println(err)
		return
	}

	log.SetFlags(log.Ldate | log.Ltime)
	log.SetOutput(logFile)

	// Setup signal catching
	sigs := make(chan os.Signal, 1)
	// Catch all signals since not explicitly listing
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)
	// Method invoked upon seeing signal
	go func() {
		s := <-sigs
		log.Printf("RECEIVED SIGNAL: %s", s)
		os.Exit(1)
	}()
}

func main() {

	listenerTLS := utils.CreateTLS(ca, crt, key, listen)
	listenerTCP := utils.CreateTCP(listenTCP)

	go utils.ListenAndServer(listenerTLS)
	utils.ListenAndServer(listenerTCP)
}
