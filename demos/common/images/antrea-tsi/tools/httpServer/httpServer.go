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

var (
	args struct {
		bindPort string
		maxConn  uint64
		logPath  string
	}
)

func healthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
}

func init() {
	flag.StringVar(&args.bindPort, "b", "8080", "TCP port the server will bind to (default 8080)")
	flag.Uint64Var(&args.maxConn, "c", 0, "maximum number of client connections the server will accept, 0 means unlimited  (default 0)")
	flag.StringVar(&args.logPath, "lp", "./server.log", "The path to the log file (default ./server.log)")
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

	router := http.NewServeMux()
	router.Handle("/", http.FileServer(http.Dir("./static")))
	router.HandleFunc("/health/", healthCheck)

	srv := http.Server{
		ReadHeaderTimeout: time.Second,
		ReadTimeout:       time.Second,
		Handler:           router,
	}

	listener, err := net.Listen("tcp", ":"+args.bindPort)
	if err != nil {
		log.Fatal(err)
	}

	if args.maxConn > 0 {
		listener = netutil.LimitListener(listener, int(args.maxConn))

		log.Printf("max connections set to %d\n", args.maxConn)
	}
	defer listener.Close()

	log.Printf("listening on %s\n", listener.Addr().String())

	for {
		if err := srv.Serve(listener); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}
}
