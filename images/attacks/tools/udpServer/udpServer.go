package main

import (
	"flag"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

var (
	ip      string
	port    int
	logPath string
)

func setupFlags() {
	flag.StringVar(&ip, "ip", "127.0.0.1", "Target IP address")
	flag.IntVar(&port, "p", 6001, "Target Port")
	flag.StringVar(&logPath, "lp", "./udpFlooding.log", "The path to the log file (default ./client.log)")
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

	conn, err := net.ListenUDP("udp", &net.UDPAddr{
		Port: port,
		IP:   net.ParseIP(ip),
	})
	if err != nil {
		panic(err)
	}

	defer conn.Close()
	log.Printf("server listening %s\n", conn.LocalAddr().String())

	for {
		message := make([]byte, 20)
		rlen, remote, err := conn.ReadFromUDP(message[:])
		if err != nil {
			panic(err)
		}

		data := strings.TrimSpace(string(message[:rlen]))
		log.Printf("received: %s from %s\n", data, remote)
	}
}
