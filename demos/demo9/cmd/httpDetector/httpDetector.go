package main

import (
	"encoding/binary"
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

type IPFlowTuplet struct {
	SrcIP net.IP
	DstIP net.IP
}

var (
	args struct {
		iface     *string
		fname     *string
		snaplen   *int
		promisc   *bool
		monitorIp *string
		threshold *int
		logPath   *string
	}
	activeConns     map[uint32]chan gopacket.Packet
	activeConnsLock sync.RWMutex
)

func init() {
	args.iface = flag.String("i", "eth0", "Interface to read packets from")
	args.fname = flag.String("r", "", "Filename to read from, overrides -i")
	args.snaplen = flag.Int("s", 65536, "Snap length (number of bytes max to read per packet")
	args.monitorIp = flag.String("ip", "", "Set monitor ip, if empty monitor all")
	args.threshold = flag.Int("t", 100, "Set the packet threshold, the value is packets per second")
	args.logPath = flag.String("lp", "./detector.log", "The path to the log file")
	flag.Parse()

	// open log file
	logFile, err := os.OpenFile(*args.logPath, os.O_APPEND|os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		log.Println(err)
		return
	}

	log.SetFlags(log.Ldate | log.Ltime)
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
	if len(ip) == 16 {
		return binary.BigEndian.Uint32(ip[12:16])
	}
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

func getPacketInfo(packet gopacket.Packet, warn chan IPFlowTuplet) {
	// we can get the MAC Addr, but it's not very usefull
	// ethernetLayer := packet.Layer(layers.LayerTypeEthernet)
	// if ethernetLayer != nil {
	// 	ethernetPacket, _ := ethernetLayer.(*layers.Ethernet)
	// 	pinfo.SrcMAC = ethernetPacket.SrcMAC
	// 	pinfo.DstMAC = ethernetPacket.DstMAC
	// }
	//get IP Addr
	var srcIPuint uint32
	var ipTuplet IPFlowTuplet
	ipLayer := packet.Layer(layers.LayerTypeIPv4)
	if ipLayer != nil {
		ip, _ := ipLayer.(*layers.IPv4)
		srcIPuint = ip2int(ip.SrcIP)
		ipTuplet = IPFlowTuplet{ip.SrcIP, ip.DstIP}
	}

	// log.Println("Packet: ", pinfo.Proto, " - ", pinfo.SrcMAC, "/", pinfo.DstMAC, " - ", pinfo.SrcIP, "/", pinfo.DstIP, " - ", pinfo.SrcPort, "/", pinfo.DstPort)
	// log.Println("Hash: ", pinfo.hashV4Flow())

	activeConnsLock.RLock()
	conn, ok := activeConns[srcIPuint]
	activeConnsLock.RUnlock()

	if !ok {
		log.Println("new connection: ", ipTuplet.SrcIP, "/", ipTuplet.DstIP)
		newconn := make(chan gopacket.Packet)
		go checkConnection(newconn, warn, ipTuplet)

		activeConnsLock.Lock()
		activeConns[srcIPuint] = newconn
		activeConnsLock.Unlock()

		newconn <- packet
	} else {
		conn <- packet
	}
}

func checkConnection(conn chan gopacket.Packet, warn chan IPFlowTuplet, pinfo IPFlowTuplet) {
	checkTimer := time.NewTicker(time.Second)
	defer checkTimer.Stop()

	timeoutTimer := time.NewTicker(10 * time.Second)
	defer timeoutTimer.Stop()

	defer close(conn)

	defer func() {
		activeConnsLock.Lock()
		delete(activeConns, ip2int(pinfo.SrcIP))
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

				// Search for a string inside the payload
				if strings.Contains(string(applicationLayer.Payload()), "HTTP") && strings.Contains(string(applicationLayer.Payload()), "GET") {
					count++
					log.Println("Count: ", count)
				}
			}

		case <-checkTimer.C:
			if count > *args.threshold {
				warn <- pinfo
			}
			count = 0
		case <-timeoutTimer.C:
			if !used {
				log.Println("Connection Timeout, closing: ", pinfo.SrcIP, "/", pinfo.DstIP)
				break
			}
			used = false
		}
	}
}

func flowClient(warn chan IPFlowTuplet) {
	for pinfo := range warn {
		log.Println("Warning: ", pinfo.SrcIP, "/", pinfo.DstIP)
	}
}

func main() {
	flag.Parse()
	activeConns = make(map[uint32]chan gopacket.Packet)
	//open flow client to send warnings
	warn := make(chan IPFlowTuplet)
	go flowClient(warn)

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
		bpffilter = "tcp dst net " + *args.monitorIp
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
