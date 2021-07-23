package main

import (
	"bytes"
	"encoding/binary"
	"flag"
	"log"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/containerd/cgroups"
	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
	specs "github.com/opencontainers/runtime-spec/specs-go"
)

var (
	iface         *string
	code          *string
	cpu_quota_us  *int64
	cpu_period_us *uint64
	mem           *int64
	logPath       *string
	control       cgroups.Cgroup
	source        *gopacket.PacketSource
	err           error
	endProgram    chan bool
	handle        *pcap.Handle
)

type ExitStatus struct {
	Signal os.Signal
	Code   int
}

func toByte(number *int64) *int64 {
	modNumber := *number * 1024 * 1024
	return &modNumber
}

/* CPU limit to remember:
*  cpuLimit	| cpu.cfs_quota_us | cpu.cfs_period_us
*  1 	    | 100000           | 100000
*  2 	    | 200000           | 100000
*  3	    | 300000           | 100000
*  m	    | m*100000         | 100000
*
 */
func init() {
	iface = flag.String("i", "eth0", "Interface to read packets from")
	//code = flag.String("code", "while(true) {var x = 1}", "The code that you want to execute in the pod")
	code = flag.String("code", `var counter =0;var client = net.createConnection("/tmp/echo.sock");client.on("data", (data)=>{counter+=JSON.stringify(data).length});`, "The code that you want to execute in the pod")
	cpu_quota_us = flag.Int64("cpu_quota", 50000, "Add cpu contraints to the program you want to execute")
	cpu_period_us = flag.Uint64("cpu_period", 100000, "Add cpu contraints to the program you want to execute")
	mem = flag.Int64("mem", 64, "Add RAM constrains in the program you want to run in MBs")
	logPath = flag.String("lp", "./executor.log", "The path to the log file")

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

	// checking for all goroutines to finish
	endProgram = make(chan bool)
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

func startCmd() {
	cmd := exec.Command("/usr/bin/nodejs", "-e", *code)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	// start app
	if err = cmd.Start(); err != nil {
		log.Println("Error on executing the user command", err)
	}

	log.Println("add pid", cmd.Process.Pid)
	if err = control.Add(cgroups.Process{Pid: cmd.Process.Pid}); err != nil {
		log.Println("Error adding on the groupds", err)
	}

	if err = cmd.Wait(); err != nil {
		log.Println("cmd return with error:", err)
	}

	defer cmd.Process.Kill()
	endProgram <- true
}

func openPcap() {
	if !deviceExists(*iface) {
		log.Fatal("Unable to open device ", *iface)
	}
	handle, err = pcap.OpenLive(*iface, int32(65536), true, pcap.BlockForever)

	if err != nil {
		log.Fatal(err)
	}

	source = gopacket.NewPacketSource(handle, handle.LinkType())
}

func createCgroup() {
	control, err = cgroups.New(cgroups.V1, cgroups.StaticPath("/test"), &specs.LinuxResources{
		Memory: &specs.LinuxMemory{
			Limit: toByte(mem),
		},
		CPU: &specs.LinuxCPU{
			Quota:  cpu_quota_us,
			Period: cpu_period_us,
		},
	})

	if err != nil {
		log.Println("Something went wrong with the group creation", err)
	}

}

func sendResponse(conn *net.UnixConn, data []byte) {
	buf := new(bytes.Buffer)
	msglen := uint32(len(data))

	binary.Write(buf, binary.BigEndian, &msglen)
	data = append(buf.Bytes(), data...)

	conn.Write(data)
}

func handleConnection(conn *net.UnixConn) {
	// Close connection when finish handling
	defer func() {
		conn.Close()
	}()
	log.Println("packet")

	// Write recursively all the packets
	for packet := range source.Packets() {
		sendResponse(conn, []byte(packet.Data()))
	}

}

func createSocket() {
	unixSocket := "/tmp/echo.sock"

	// Remove the existing socket file
	os.Remove(unixSocket)

	// Get unix socket address based on file path
	uaddr, err := net.ResolveUnixAddr("unix", unixSocket)
	if err != nil {
		log.Println(err)
		return
	}

	// Listen on the address
	unixListener, err := net.ListenUnix("unix", uaddr)
	if err != nil {
		log.Println(err)
		return
	}

	/* Close listener when close this function,
	*  you can also emit it because this function
	*  will not terminate gracefully
	 */
	defer unixListener.Close()

	// Monitor request and process
	for {
		conn, err := unixListener.AcceptUnix()
		if err != nil {
			log.Println(err)
			continue
		}

		// Handle request
		go handleConnection(conn)
	}
}

func main() {
	flag.Parse()

	/* Create the restricted cgroup for
	*  for all the code that is to be executed
	 */
	createCgroup()
	defer control.Delete()

	/* Start reading packets from the selected interface
	*  and create the source file
	 */
	openPcap()
	defer handle.Close()

	/* Create the unix socket from bytes
	 *  steam of packages
	 */
	go createSocket()

	/* Execute the command
	*
	 */
	go startCmd()

	/* Waiting for all subroutines to finish
	*
	 */
	<-endProgram
}
