/*
 * JA3 - TLS Client Hello Hash
 * Copyright (c) 2017, Salesforce.com, Inc.
 * this code was created by Philipp Mieden <dreadl0ck [at] protonmail [dot] ch>
 *
 * THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES
 * WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF
 * MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR
 * ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES
 * WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN
 * ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF
 * OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.
 */

package main

import (
	"flag"
	"fmt"
	"github.com/dreadl0ck/ja3"
	"github.com/google/gopacket/pcap"
	"os"
	"log"
	"os/signal"
	"syscall"
	"tls_fingerprinting/pkg/statistics"
)

var (
	flagJSON      = flag.Bool("json", true, "print as JSON array")
	flagCSV       = flag.Bool("csv", false, "print as CSV")
	flagTSV       = flag.Bool("tsv", false, "print as TAB separated values")
	flagSeparator = flag.String("separator", ",", "set a custom separator")
	flagInput     = flag.String("read", "", "read PCAP file")
	flagDebug     = flag.Bool("debug", false, "toggle debug mode")
	flagInterface = flag.String("iface", "", "specify network interface to read packets from")
	flagJa3S      = flag.Bool("ja3s", true, "include ja3 server hashes (ja3s)")
	flagOnlyJa3S  = flag.Bool("ja3s-only", false, "dump ja3s only")
	flagSnaplen   = flag.Int("snaplen", 1514, "default snap length for ethernet frames")
	flagPromisc   = flag.Bool("promisc", true, "capture in promiscuous mode (requires root)")
	flagTimeout   = flag.Duration("timeout", pcap.BlockForever, "timeout for collecting packet batches")
	flagLogPath   = flag.String("logpath", "./tls_fingerprinting_statistics.log", "the path to the logfile")
	flagServerIp  = flag.String("serverIp", "10.1.1.205", "the statistics server ip")
	flagServerPort= flag.String("serverPort", "30002", "the statistics server port")
	flagPollingRate= flag.Int("pollingRate", 5, "The rate in which is going to check for new data")
	logFile 	  = os.Stdout
)


func init() {
	flag.Parse()

	// open log files
	logFile, _ = os.OpenFile(*flagLogPath, os.O_APPEND|os.O_RDWR|os.O_CREATE, 0644)
	errlogFile, _ := os.OpenFile("tls_fingerprinting_err.log", os.O_APPEND|os.O_RDWR|os.O_CREATE, 0644)

	log.SetFlags(log.Ldate | log.Ltime)
	log.SetOutput(errlogFile)

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

	flag.Parse()
	primitive:="tls-fingerprinting"
	go statistics.HandleStatistics(&primitive, flagServerIp, flagServerPort, flagLogPath, *flagPollingRate)

	ja3.Debug = *flagDebug

	if *flagInterface != "" {
		ja3.ReadInterface(*flagInterface, logFile, *flagSeparator, *flagJa3S, *flagJSON, *flagSnaplen, *flagPromisc, *flagTimeout)
		return
	}

	if *flagInput == "" {
		fmt.Println("use the -read flag to supply an input file.")
		os.Exit(1)
	}

	if *flagOnlyJa3S {
		ja3.ReadFileJa3s(*flagInput, logFile)
		return
	}

	if *flagTSV {
		ja3.ReadFileCSV(*flagInput, logFile, "\t", *flagJa3S)
		return
	}

	if *flagCSV {
		ja3.ReadFileCSV(*flagInput, logFile, *flagSeparator, *flagJa3S)
		return
	}

	if *flagJSON {
		ja3.ReadFileJSON(*flagInput, logFile, *flagJa3S)
	}
}
