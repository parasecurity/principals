package main

import (
	"encoding/binary"
	"encoding/json"
	"flag"
	"log"
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
		monitorIp  *string
		threshold  *int
		logPath    *string
		flowServer *string
		command    *string
		arguments  *string
	}
	activeConns     map[uint32]chan gopacket.Packet
	activeConnsLock sync.RWMutex
)

type message map[string]interface{}

func init() {
	args.iface = flag.String("i", "eth0", "Interface to read packets from")
	args.fname = flag.String("r", "", "Filename to read from, overrides -i")
	args.snaplen = flag.Int("s", 65536, "Snap length (number of bytes max to read per packet")
	args.monitorIp = flag.String("ip", "", "Set monitor ip, if empty monitor all")
	args.threshold = flag.Int("t", 300, "Set the packet threshold, the value is packets per second")
	args.logPath = flag.String("lp", "./detector.log", "The path to the log file")
	args.flowServer = flag.String("fc", "10.1.1.201:30002", "The flow server connection in format ip:port e.g. 10.1.1.101:8080")
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

func ip2int(ip net.IP) uint32 {
	return binary.BigEndian.Uint32(ip)
}

func int2ip(nn uint32) net.IP {
	ip := make(net.IP, 4)
	binary.BigEndian.PutUint32(ip, nn)
	return ip
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
		conn, ok := activeConns[ip2int(ip.SrcIP)]
		activeConnsLock.RUnlock()

		if !ok {
			log.Println("new connection: ", ip.SrcIP, " -> ", *args.monitorIp)
			newconn := make(chan gopacket.Packet)
			go checkConnection(newconn, warn, ip.SrcIP)

			activeConnsLock.Lock()
			activeConns[ip2int(ip.SrcIP)] = newconn
			activeConnsLock.Unlock()

			newconn <- packet
		} else {
			conn <- packet
		}

	}

	return
}

func checkConnection(conn chan gopacket.Packet, warn chan net.IP, srcIP net.IP) {
	checkTimer := time.NewTicker(time.Second)
	defer checkTimer.Stop()

	timeoutTimer := time.NewTicker(10 * time.Second)
	defer timeoutTimer.Stop()

	defer close(conn)

	defer func() {
		activeConnsLock.Lock()
		delete(activeConns, ip2int(srcIP))
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

				payloadStr := string(applicationLayer.Payload())
				// Search for a string inside the payload
				if strings.Contains(payloadStr, "HTTP") && strings.Contains(payloadStr, "GET") {
					count++
				}
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
				log.Println("Warning: ", srcIP, " -> ", *args.monitorIp, " count: ", count)
				warn <- srcIP
			}
			log.Println("count: ", srcIP, " -> ", *args.monitorIp, " count: ", count)
			count = 0
		case <-timeoutTimer.C:
			if !used {
				log.Println("Connection Timeout, closing: ", srcIP, " -> ", *args.monitorIp)
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
		_, err := conn.Write([]byte(string(jsonMsg) + "\n"))
		if err != nil {
			log.Println(err)
			return
		}
	}
}

func main() {
	flag.Parse()
	activeConns = make(map[uint32]chan gopacket.Packet)
	//open flow client to send warnings
	warn := make(chan net.IP)
	go flowServer(warn)

	//open pcap to get packets
	var handle *pcap.Handle
	var err error

	if *args.fname != "" {
		if handle, err = pcap.OpenOffline(*args.fname); err != nil {
			log.Fatal("PCAP OpenOffline error:", err)
		}
	} else {
		if !deviceExists(*args.iface) {
			log.Fatal("Unable to open device ", *args.iface)
		}
		handle, err = pcap.OpenLive(*args.iface, int32(*args.snaplen), true, pcap.BlockForever)

		if err != nil {
			log.Fatal(err)
		}
		defer handle.Close()
	}

	var bpffilter string
	if *args.monitorIp != "" {
		ip := net.ParseIP(*args.monitorIp)
		if ip == nil {
			log.Fatal("monitor ip has wrong format: ", *args.monitorIp)
		}
		bpffilter = "tcp and dst host " + *args.monitorIp
	} else {
		bpffilter = "tcp"
	}
	if err = handle.SetBPFFilter(bpffilter); err != nil {
		log.Fatal("BPF filter error:", err)
	}

	source := gopacket.NewPacketSource(handle, handle.LinkType())

	for packet := range source.Packets() {
		getPacketInfo(packet, warn)
	}
}
