package main

import (
	"bufio"
	"encoding/json"
	"flag"
	log "logging"
	"net"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)

var (
	args struct {
		iface      *string
		fname      *string
		snaplen    *int
		promisc    *bool
		threshold  *int
		logPath    *string
		flowServer *string
		listen     *string
		command    *string
		arguments  *string
	}
	activeConns     map[string]chan gopacket.Packet
	activeConnsLock sync.RWMutex
	monitoredIPs    []string
	monitorAll      bool
)

type message map[string]interface{}

func init() {
	args.iface = flag.String("i", "eth0", "Interface to read packets from")
	args.fname = flag.String("r", "", "Filename to read from, overrides -i")
	args.snaplen = flag.Int("s", 65536, "Snap length (number of bytes max to read per packet")
	args.threshold = flag.Int("t", 150, "Set the packet threshold, the value is packets per second")
	args.logPath = flag.String("lp", "./detector.log", "The path to the log file")
	args.flowServer = flag.String("fc", "10.1.1.201:30002", "The flow server connection in format ip:port e.g. 10.1.1.101:8080")
	args.listen = flag.String("l", "10.1.1.202:30000", "The IP and port of the secondary network that listens for connections")
	args.command = flag.String("c", "block", "The command to execute when a malicious behaviour is detected e.g. block, tarpit..")
	args.arguments = flag.String("args", "", "Arguments to pass to the command you want to execute")
	flag.Parse()

	// open log file
	logFile, err := os.OpenFile(*args.logPath, os.O_APPEND|os.O_RDWR|os.O_CREATE, 0644)
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

func arrayContains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}

	return false
}

func deviceExists(name string) bool {
	devices, err := pcap.FindAllDevs()

	if err != nil {
		log.Panic(err)
	}

	for _, device := range devices {
		if device.Name == name {
			return true
		}
	}
	return false
}

func getPacketInfo(packet gopacket.Packet, warn chan net.IP) {

	ipLayer := packet.Layer(layers.LayerTypeIPv4)
	if ipLayer != nil {
		ip, _ := ipLayer.(*layers.IPv4)

		activeConnsLock.RLock()
		connStr := ip.SrcIP.String() + ":" + ip.DstIP.String()
		conn, ok := activeConns[connStr]
		activeConnsLock.RUnlock()

		if !ok {
			log.Println("new connection: ", connStr)
			newconn := make(chan gopacket.Packet)
			go checkConnection(newconn, warn, ip.SrcIP, connStr)

			activeConnsLock.Lock()
			activeConns[connStr] = newconn
			activeConnsLock.Unlock()

			newconn <- packet
		} else {
			conn <- packet
		}
	}
}

func checkConnection(conn chan gopacket.Packet, warn chan net.IP, srcIP net.IP, connStr string) {
	checkTimer := time.NewTicker(500 * time.Millisecond)
	defer checkTimer.Stop()

	timeoutTimer := time.NewTicker(10 * time.Second)
	defer timeoutTimer.Stop()

	defer close(conn)

	defer func() {
		activeConnsLock.Lock()
		delete(activeConns, connStr)
		activeConnsLock.Unlock()
	}()

	var count int
	var used = false
	count = 0

	for {
		select {
		case p, ok := <-conn:
			if !ok {
				break
			}
			used = true
			applicationLayer := p.ApplicationLayer()
			if applicationLayer != nil {

				//payloadStr := string(applicationLayer.Payload())
				// Search for a string inside the payload
				count++
			}
			//debug

			ipLayer := p.Layer(layers.LayerTypeIPv4)
			if ipLayer != nil {
				ip, _ := ipLayer.(*layers.IPv4)
				if !ip.SrcIP.Equal(srcIP) {
					log.Println("inside check connection: IPs", ip.SrcIP.String(), " and ", srcIP.String(), " differ")
				}
			}

		case <-checkTimer.C:
			if count > *args.threshold {
				log.Println("Warning: ", connStr, " count: ", count)
				warn <- srcIP
			}
			log.Println("count: ", connStr, " count: ", count)
			count = 0
		case <-timeoutTimer.C:
			if !used {
				log.Println("Connection Timeout, closing: ", connStr)
				return
			}
			used = false
		}
	}
}

func flowServer(warn chan net.IP) {
	conn, err := net.Dial("tcp", *args.flowServer)
	if err != nil {
		log.Println(err)
		return
	}
	defer conn.Close()

	for srcIP := range warn {
		msg := &message{
			"Action": *args.command,
			"Argument": map[string]interface{}{
				"Ip": srcIP.String(),
			},
		}
		jsonMsg, _ := json.Marshal(msg)
		jsonMsg = append(jsonMsg, []byte("\n")...)

		log.Println(string(jsonMsg))

		_, err := conn.Write(jsonMsg)
		for err != nil {
			log.Println(err)
			log.Println("Reopening connection")
			conn.Close()
			conn, err = net.Dial("tcp", *args.flowServer)
			if err != nil {
				log.Println(err)
				continue
			}
			_, err = conn.Write(jsonMsg)
		}
	}
}

func updateMonitoredIPs(handle *pcap.Handle) {
	// Initialy BPF filter to track the `lo` network
	var bpffilter string = "net 127.0.0.0"
	if err := handle.SetBPFFilter(bpffilter); err != nil {
		log.Println("BPF filter error:", err)
	}
	log.Println("Initial bpf: " + bpffilter)

	monitorAll = false
	// whenever a flow controller connects we open a new reader routine

	listener, err := net.Listen("tcp4", *args.listen)
	if err != nil {
		log.Println(err)
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

			netIP := strings.TrimSpace(string(netData))
			log.Println("Received IP from ", c.RemoteAddr().String(), ": ", netIP)

			if netIP == "all" {
				monitorAll = true
			} else {
				if !arrayContains(monitoredIPs, netIP) {
					monitorAll = false
					monitoredIPs = append(monitoredIPs, netIP)
				}
			}
			bpffilter = ""
			if !monitorAll {
				if len(monitoredIPs) == 0 {
					// monitor loopback, zero traffic inside container
					bpffilter = "net 127.0.0.0"
				} else {
					// monitor array of ips
					for n, ip := range monitoredIPs {
						if n == 0 {
							bpffilter = "((udp or tcp) and dst host  " + ip + " )"
						} else {
							bpffilter = bpffilter + " or ((udp or tcp) and host " + ip + " )"
						}
					}
				}
			} else {
				// monitor all tcp and udp traffic
				bpffilter = "(udp or tcp)"
			}

			log.Println("Updating bpf: " + bpffilter)

			if err = handle.SetBPFFilter(bpffilter); err != nil {
				log.Println("BPF filter error:", err)
			}
		}

		/*

					if action == "add" {
						if argument == "all" {
							monitorAll = true
						} else {
							//if we add an ip, should we disable all?
							monitoredIPs = append(monitoredIPs, line.Text)
						}
					} else if action == "remove" {
						if argument == "all" {
							monitorAll = false
						} else {
							index := 0
							for n, ip := range monitoredIPs {
								if ip == argument {
									index = n
								}
							}
							//swap unneeded ip with last valid ip
							monitoredIPs[len(monitoredIPs)-1], monitoredIPs[i] = monitoredIPs[i], monitoredIPs[len(monitoredIPs)-1]
							//keep a new slice with n-1 ip (drop the last ip)
			    			monitoredIPs = monitoredIPs[:len(monitoredIPs)-1]
						}
					} else if action == "clear" {
						monitoredIPs = monitoredIPs[:0]
					}

		*/
	}
}

func main() {
	flag.Parse()
	activeConns = make(map[string]chan gopacket.Packet)
	//open flow client to send warnings
	warn := make(chan net.IP)
	go flowServer(warn)

	//open pcap to get packets
	var handle *pcap.Handle
	var err error

	if *args.fname != "" {
		if handle, err = pcap.OpenOffline(*args.fname); err != nil {
			log.Println("PCAP OpenOffline error:", err)
		}
	} else {
		if !deviceExists(*args.iface) {
			log.Println("Unable to open device ", *args.iface)
		}

		handle, err = pcap.OpenLive(*args.iface, int32(*args.snaplen), true, pcap.BlockForever)

		if err != nil {
			log.Println(err)
		}
		defer handle.Close()
	}

	go updateMonitoredIPs(handle)

	source := gopacket.NewPacketSource(handle, handle.LinkType())
	for packet := range source.Packets() {
		getPacketInfo(packet, warn)
	}
}
