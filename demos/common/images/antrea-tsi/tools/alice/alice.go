package main

import (
	"flag"
	"io/ioutil"
	log "logging"
	"net/http"
	"os"
	"os/signal"
	"sort"
	"strings"
	"syscall"
	"time"

	rng "github.com/leesper/go_rng"
)

var (
	args struct {
		server       string
		sleep        int
		timeout      int
		clients      int
		logPath      string
		mean         float64
		stddev       float64
		displayDistr bool
	}
)

func init() {
	flag.StringVar(&args.server, "conn", "http://localhost:8080/", "The server connection in format ip:port (default http://localhost:8080/)")
	flag.IntVar(&args.sleep, "s", 1000, "sleep in ms (default 1000)")
	flag.IntVar(&args.timeout, "t", 1000, "timeout in ms (default 1000)")
	flag.IntVar(&args.clients, "c", 1, "number of concurrent clients (default 1)")
	flag.StringVar(&args.logPath, "lp", "./alice.log", "The path to the log file (default ./client.log)")
	flag.Float64Var(&args.mean, "m", 5.0, "The mean value for the normal distribution pattern of client")
	flag.Float64Var(&args.stddev, "dev", 2.0, "The deviation value for the normal distribution pattern of client")
	flag.BoolVar(&args.displayDistr, "print", false, "Print a representation of the normal distribution pattern of client")
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

func getGaussian(mean, stddev float64) (map[int64]int, []int64) {
	grng := rng.NewGaussianGenerator(time.Now().UnixNano())
	hist := map[int64]int{}
	for i := 0; i < 10000; i++ {
		hist[int64(grng.Gaussian(mean, stddev))]++
	}

	keys := []int64{}
	for k := range hist {
		keys = append(keys, k)
	}
	SortInt64Slice(keys)

	return hist, keys
}

func printDistribution(print bool, hist map[int64]int, keys []int64) {

	if args.displayDistr == true {
		for _, key := range keys {
			log.Printf("%d:\t%s\n", key, strings.Repeat("*", hist[key]/200))
		}
	}
}

func SortInt64Slice(slice []int64) {
	sort.Sort(int64slice(slice))
}

type int64slice []int64

func (slice int64slice) Len() int {
	return len(slice)
}

func (slice int64slice) Less(i, j int) bool {
	return slice[i] < slice[j]
}

func (slice int64slice) Swap(i, j int) {
	slice[i], slice[j] = slice[j], slice[i]
}

func realGetRequest() {
	for {
		httpClient := &http.Client{
			Timeout: 2 * time.Second,
		}

		resp, err := httpClient.Get("http://147.27.15.134")
		if err != nil {
			log.Println("No Response", err)
		}
		defer resp.Body.Close()
		log.Println("Normal Request to Google.com")

		time.Sleep(500 * time.Millisecond)
	}
}

func main() {
	t := http.DefaultTransport.(*http.Transport).Clone()
	// t.MaxIdleConns = 0
	// t.MaxConnsPerHost = 0
	// t.MaxIdleConnsPerHost = 5000
	// t.IdleConnTimeout = 0
	t.MaxIdleConns = 100
	t.MaxConnsPerHost = 100
	t.MaxIdleConnsPerHost = 100

	hist, keys := getGaussian(args.mean, args.stddev)
	printDistribution(args.displayDistr, hist, keys)
	repeat := 0

	// Start sending benigh requests to external server
	go realGetRequest()

	for {
		for _, key := range keys {
			conc := hist[key] / 20
			conc = conc * args.clients

			// log.Println(conc)
			for cl := 0; cl < conc; cl++ {
				go func(r int, conc int, c int, t *http.Transport) {
					httpClient := &http.Client{
						Timeout:   time.Duration(args.timeout) * time.Millisecond,
						Transport: t,
					}
					start := time.Now()
					resp, err := httpClient.Get(args.server)
					interval := time.Since(start)
					if err != nil {
						log.Println("No Response", err, interval)
						return
					}
					defer resp.Body.Close()

					bytes, err := ioutil.ReadAll(resp.Body)
					if err != nil {
						log.Println(err, interval)
					}
					// log.Println("Response status:", resp.Status, r, conc, c, ", bytes: ", len(bytes), interval)
					log.Println("Response status:", resp.Status, ", bytes: ", len(bytes), interval)
				}(repeat, conc, cl, t)
			}
			time.Sleep(time.Duration(args.sleep) * time.Millisecond)
		}
		repeat++
	}

}
