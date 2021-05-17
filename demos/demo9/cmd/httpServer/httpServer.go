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

	"golang.org/x/net/netutil"
)

const (
	defaultBindAddr = ":8080"

	// defaultMaxConn is the default number of max connections the
	// server will handle. 0 means no limits will be set, so the
	// server will be bound by system resources.
	defaultMaxConn = 0
)

var (
	bindAddr string
	maxConn  uint64
	logPath  string
)

func healthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
}

func init() {
	flag.StringVar(&bindAddr, "b", defaultBindAddr, "TCP address the server will bind to")
	flag.Uint64Var(&maxConn, "c", defaultMaxConn, "maximum number of client connections the server will accept, 0 means unlimited")
	flag.StringVar(&logPath, "lp", "./server.log", "The path to the log file")
	flag.Parse()

	// open log file
	logFile, err := os.OpenFile(logPath, os.O_APPEND|os.O_RDWR|os.O_CREATE, 0644)
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

	router := http.NewServeMux()
	router.Handle("/", http.FileServer(http.Dir("./static")))
	router.HandleFunc("/health/", healthCheck)

	srv := http.Server{
		ReadHeaderTimeout: time.Second * 5,
		ReadTimeout:       time.Second * 10,
		Handler:           router,
	}

	listener, err := net.Listen("tcp", ":3000")
	if err != nil {
		log.Fatal(err)
	}

	if maxConn > 0 {
		listener = netutil.LimitListener(listener, int(maxConn))

		log.Printf("max connections set to %d\n", maxConn)
	}
	defer listener.Close()

	log.Printf("listening on %s\n", listener.Addr().String())

	for {
		if err := srv.Serve(listener); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}
}
