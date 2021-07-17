package main

import (
	"flag"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	ovsctl "github.com/antrea-io/antrea/pkg/ovs/ovsctl"
)

var (
	port         *string
	detectorIP   *string
	detectorPort *int
	threshold    *int
	failures     *int
	logPath      *string
	detectorUp   bool = false
	failureCount int
	conn         net.Conn
	localaddr    net.TCPAddr
	remoteaddr   net.TCPAddr
)

type statistics struct {
	pktsIn   int
	bytesIn  int
	pktsOut  int
	bytesOut int
}

func toMbps(bytes int) int {
	return bytes / 1000000
}

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

	// Connect to the tls server
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

func getStatistics(port string) statistics {
	var statistic statistics
	client := ovsctl.NewClient("br-int")
	res, err := client.RunOfctlCmd("dump-ports", port)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	data := string(res)
	data_arr := strings.Fields(data)

	statistic.pktsOut, _ = strconv.Atoi(strings.TrimRight(strings.Split(data_arr[9], "=")[1], ","))
	statistic.bytesOut, _ = strconv.Atoi(strings.TrimRight(strings.Split(data_arr[10], "=")[1], ","))
	statistic.pktsIn, _ = strconv.Atoi(strings.TrimRight(strings.Split(data_arr[17], "=")[1], ","))
	statistic.bytesIn, _ = strconv.Atoi(strings.TrimRight(strings.Split(data_arr[18], "=")[1], ","))

	return statistic
}

func enableDetector() {
	if detectorUp {
		return
	}

	msg := string(("all" + "\n"))
	_, err := conn.Write([]byte(msg))
	for err != nil {
		// If connection closes we try again
		log.Println(err)
		log.Println("Reopening connection")
		localaddr.Port = localaddr.Port + 1
		conn, err = net.DialTCP("tcp", &localaddr, &remoteaddr)
		if err != nil {
			log.Println(err)
			continue
		}
		_, err = conn.Write([]byte(msg))
	}

	detectorUp = true
	log.Println("Enabled detectors...")
}

func timeGet(port string) {
	var statistic statistics = getStatistics(port)
	var lastStatistic statistics
	for {
		lastStatistic = statistic
		statistic = getStatistics(port)

		inMbps := toMbps(statistic.bytesIn - lastStatistic.bytesIn)
		outMbps := toMbps(statistic.bytesOut - lastStatistic.bytesOut)
		log.Println("(In/Out)", inMbps, outMbps)

		if inMbps > *threshold || outMbps > *threshold {
			log.Println("Threshold passed (In/Out)", inMbps, outMbps)
			failureCount++
			if failureCount >= *failures {
				enableDetector()
				failureCount = 0
			}
		} else {
			failureCount = 0
		}
		time.Sleep(time.Second)
	}
}

func init() {
	port = flag.String("i", "antrea-gw0", "The port interface you want to monitor e.g. coredns--ec5e46")
	detectorIP = flag.String("d", "10.1.1.203", "The API server url e.g. 10.1.1.203")
	detectorPort = flag.Int("p", 30000, "The detector port address e.g. 30002")
	threshold = flag.Int("t", 10, "The Mbps threshold")
	failures = flag.Int("f", 4, "The number of failures before we spawn a detector")
	logPath = flag.String("lp", "./canary-link.log", "The path to the log file")
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
}

func main() {
	failureCount = 0
	timeGet(*port)
}
