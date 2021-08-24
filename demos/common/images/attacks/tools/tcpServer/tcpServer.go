package main

import (
	"bufio"
	"flag"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
)

var (
	ip      string
	port    string
	server  string
	logPath string
)

func setupFlags() {
	flag.StringVar(&ip, "ip", "127.0.0.1", "Target IP address")
	flag.StringVar(&port, "p", "6001", "Target Port")
	flag.StringVar(&logPath, "lp", "./tcpServer.log", "The path to the log file (default ./client.log)")
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
	/* Parse the given flags
	 *
	 */
	flag.Parse()

	/* Generate the server address
	*
	 */
	generateServerAddress()
	log.Println(server)
	listener, err := net.Listen("tcp", server)
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()

	for {
		c, err := listener.Accept()
		if err != nil {
			log.Println(err)
			return
		}
		reader := bufio.NewReader(c)
		for {
			netData, err := reader.ReadString('\n')
			if err != nil {
				log.Println(err)
				break
			}

			netIP := string(netData)
			log.Println("Received IP from ", c.RemoteAddr().String(), ": ", netIP)

		}
	}

}
