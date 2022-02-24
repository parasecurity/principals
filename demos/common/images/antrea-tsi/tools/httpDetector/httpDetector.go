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
		syn        *bool
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
	args.syn = flag.Bool("syn", false, "Check if it is an syn attack")
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
		connStr := ip.SrcIP.String() // + ":" + ip.DstIP.String()
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
	checkTimer := time.NewTicker(2000 * time.Millisecond)
	defer checkTimer.Stop()

	timeoutTimer := time.NewTicker(10 * time.Second)
	defer timeoutTimer.Stop()

	defer close(conn)

	defer func() {
		activeConnsLock.Lock()
		delete(activeConns, connStr)
		activeConnsLock.Unlock()
	}()

	var used = false
	distinct_IP := -1
	var firstIP, firstNet net.IP
	totalPackets := 0
	tcpPackets := 0
	udpPackets := 0
	tcpSYNPackets := 0
	httpPackets := 0

	for {
		select {
		case p, ok := <-conn:
			if !ok {
				break
			}
			used = true
			checkForDistict := false

			applicationLayer := p.ApplicationLayer()

			if applicationLayer != nil {
				payloadStr := string(applicationLayer.Payload())
				// Search for a string inside the payload
				if strings.Contains(payloadStr, "HTTP") && strings.Contains(payloadStr, "POST") {
					httpPackets++
				}
			}

			if tcpLayer := p.Layer(layers.LayerTypeTCP); tcpLayer != nil {
				// Get actual TCP data from this layer
				tcp, _ := tcpLayer.(*layers.TCP)

				if tcp.SYN {
					tcpSYNPackets++
					checkForDistict = true
				}
				tcpPackets++
			}

			if udpLayer := p.Layer(layers.LayerTypeUDP); udpLayer != nil {
				// Get actual TCP data from this layer
				udpPackets++
				checkForDistict = true
			}

			totalPackets++

			ipLayer := p.Layer(layers.LayerTypeIPv4)
			if (ipLayer != nil) && checkForDistict {
				ip, _ := ipLayer.(*layers.IPv4)
				ipDst_Str := ip.DstIP.String()
				if distinct_IP == -1 {
					firstIP = ip.DstIP
					firstNet = ip.DstIP.Mask(ip.DstIP.DefaultMask())
					distinct_IP = 0
				} else if distinct_IP == 0 && ipDst_Str != "0.0.0.0" {
					newNet := ip.DstIP.Mask(ip.DstIP.DefaultMask())
					if (!ip.DstIP.Equal(firstIP)) && (!newNet.Equal(firstNet)) {
						distinct_IP = 1
					}
				}
			}

		case <-checkTimer.C:
			if totalPackets == 0 {
				break
			}
			// TODO: Remove the coredns from here
			// pass it from the canary
			if srcIP.String() == "10.10.0.96" || srcIP.String() == "10.10.0.67" || srcIP.String() == "130.207.39.90" {
				break
			}

			httpPers := (httpPackets * 100) / totalPackets
			udpPers := (udpPackets * 100) / totalPackets
			synPers := (tcpSYNPackets * 100) / totalPackets

			if (httpPers > 30) && (httpPackets > 30) {
				log.Println("Http Attack: ", httpPers, "Packets: ", httpPackets, "Ip to block: ", srcIP.String())
				warn <- srcIP
			}

			if (udpPers > 15) && (distinct_IP != 1) && (udpPackets > 30) {
				log.Println("Udp Attack: ", udpPers, "Packets: ", udpPackets, "Ip to block: ", srcIP.String())
				warn <- srcIP
			}

			if (synPers > 40) && (distinct_IP != 1) && (tcpSYNPackets > 30) {
				log.Println("Syn Attack: ", synPers, "Packets: ", tcpSYNPackets, "Ip to block: ", srcIP.String())
				warn <- srcIP
			}

			log.Println("udpPer: ", udpPers, " SynPer: ", synPers, " distinct_IP ", distinct_IP, "srcIP ", srcIP.String())
			log.Println("tcp", tcpPackets, "tcpSyn", tcpSYNPackets, "udp", udpPackets, "total", totalPackets, "srcIP ", srcIP.String())

			//count = 0
			udpPackets = 0
			tcpSYNPackets = 0
			udpPackets = 0
			totalPackets = 0
			tcpPackets = 0

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
			netDataSpl := strings.Split(netData, "|")
			netIP := strings.TrimSpace(string(netDataSpl[0]))
			canaryIP := strings.TrimSpace(netDataSpl[1])
			log.Println("Received IP from ", canaryIP, ": ", netIP)

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
							bpffilter = "((udp or tcp) and not host " + canaryIP + " )"
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
