package main

import (
	"flag"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var (
	server       *string
	detectorIP   *string
	detectorPort *int
	threshold    *int
	failures     *int
	logPath      *string
	failureCount int
	ips          []net.IP
	conn         net.Conn
	localaddr    net.TCPAddr
	remoteaddr   net.TCPAddr
)

func getInterfaceIpv4Addr(interfaceName string) (addr string, err error) {
	var (
		ief      *net.Interface
		addrs    []net.Addr
		ipv4Addr net.IP
	)
	if ief, err = net.InterfaceByName(interfaceName); err != nil { // get interface
		return
	}
	if addrs, err = ief.Addrs(); err != nil { // get addresses
		return
	}
	for _, addr := range addrs { // get ipv4 address
		if ipv4Addr = addr.(*net.IPNet).IP.To4(); ipv4Addr != nil {
			break
		}
	}
	if ipv4Addr == nil {
		return "", nil
	}
	return ipv4Addr.String(), nil
}

func connectTCP() net.Conn {
	// Get net1 interface ip
	ip, _ := getInterfaceIpv4Addr("net1")

	localaddr.IP = net.ParseIP(ip)
	localaddr.Port = 6000
	remoteaddr.IP = net.ParseIP(*detectorIP)
	remoteaddr.Port = *detectorPort

	// Connect to the detector
	connection, err := net.DialTCP("tcp", &localaddr, &remoteaddr)
	for err != nil {
		log.Println("Trying to connect to detector...")
		localaddr.Port = localaddr.Port + 1
		connection, err = net.DialTCP("tcp", &localaddr, &remoteaddr)
		if err != nil {
			log.Println(err)
			continue
		}
	}
	return connection
}

func sendIP(ip string) {
	msg := string((ip + "\n"))

	_, err := conn.Write([]byte(msg))
	for err != nil {
		// If connection closes we try again
		log.Println(err)
		log.Println("Reopening connection")
		localaddr.Port = localaddr.Port + 1
		conn, err = net.DialIP("tcp", &localaddr, &remoteaddr)
		if err != nil {
			log.Println(err)
			continue
		}
		_, err = conn.Write([]byte(msg))
	}
}

func enableDetector(ip string) {
	sendIP(ip)
	log.Println("Enabled detectors...")
}

func timeGet(urlC string) {
	t := http.DefaultTransport.(*http.Transport).Clone()
	t.MaxIdleConns = 100
	t.MaxConnsPerHost = 100
	t.MaxIdleConnsPerHost = 100

	for {
		defer func() {
			if r := recover(); r != nil {
				log.Println("Canary connection timeout")
				failureCount++
				if failureCount >= *failures {
					enableDetector(ips[0].String())
					failureCount = 0
				}
			}
		}()
		httpClient := &http.Client{
			Timeout:   1 * time.Second,
			Transport: t,
		}
		start := time.Now()
		r, err := httpClient.Get(urlC)
		r.Body.Close()
		interval := time.Since(start)
		log.Println("Response in :", interval)

		if err != nil {
			log.Println(err)
		}

		if interval > time.Duration(*threshold)*time.Millisecond {
			log.Println("Threshold passed:", interval)
			failureCount++
			if failureCount >= *failures {
				enableDetector(ips[0].String())
				failureCount = 0
			}
		} else {
			failureCount = 0
		}
		httpClient.CloseIdleConnections()
		time.Sleep(time.Second)
	}
}

func init() {
	server = flag.String("conn", "http://kronos.mhl.tuc.gr:30001/health/", "The server url e.g. http://147.27.39.116:8080/health/")
	detectorIP = flag.String("d", "10.1.1.202", "The detector IP address e.g. 10.1.1.202")
	detectorPort = flag.Int("p", 30000, "The detector port address e.g. 30002")
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

	// Connect to detector
	conn = connectTCP()

	log.Println("Monitor URL: ", *server)
	u, err := url.Parse(*server)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Hostname: ", u.Hostname())
	ips_t, _ := net.LookupIP(u.Hostname())
	for _, ip := range ips_t {
		if ipv4 := ip.To4(); ipv4 != nil {
			ips = append(ips, ip)
			log.Println("IPv4: ", ipv4)
		}
	}
}

func main() {
	failureCount = 0
	for {
		timeGet(*server)
	}
}
