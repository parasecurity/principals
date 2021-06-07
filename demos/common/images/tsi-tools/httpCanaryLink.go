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

	ovsctl "github.com/vmware-tanzu/antrea/pkg/ovs/ovsctl"
)

var (
	port       *string
	api        *string
	threshold  *int
	failures   *int
	logPath    *string
	detectorUp bool = false
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

func connectTCP() net.Conn {
	addr := *api

	// Connect to the tls server
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		log.Println("Failed to connect:", err.Error())
	}
	return conn
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

func createDetector() {
	if detectorUp == true {
		return
	}

	conn := connectTCP()
	defer conn.Close()

	command := "create detector-link"
	_, err := conn.Write([]byte(command))
	if err != nil {
		log.Println(err)
	}
	detectorUp = true
	log.Println("Deploying detectors...")
}

func timeGet(port string) {
	var statistic statistics = getStatistics(port)
	var lastStatistic statistics
	var failureCount int
	for {
		lastStatistic = statistic
		statistic = getStatistics(port)

		inMbps := toMbps(statistic.bytesIn - lastStatistic.bytesIn)
		outMbps := toMbps(statistic.bytesOut - lastStatistic.bytesOut)
		log.Println("(In/Out)", inMbps, outMbps)

		if inMbps > *threshold || outMbps > *threshold {
			log.Println("Threshold passed (In/Out)", inMbps, outMbps)
			if failureCount >= *failures {
				log.Println("Creating detectors")
				createDetector()
				failureCount = 0
			} else {
				failureCount++
			}
		} else {
			failureCount = 0
		}
		time.Sleep(time.Second)
	}
}

func init() {
	port = flag.String("i", "antrea-gw0", "The port interface you want to monitor e.g. coredns--ec5e46")
	api = flag.String("api", "10.244.0.9:8001", "The API server url e.g. 10.244.0.9:8001")
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
}

func main() {
	timeGet(*port)
}
