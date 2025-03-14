package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	log "logging"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

type command struct {
	Action   string   `json:"action"`
	Argument argument `json:"argument"`
}

type argument struct {
	Ip           string `json:"ip"`
	Limit        string `json:"limit"`
	Port         string `json:"port"`
	Honeypot_ip  string `json:"honeypot_ip"`
	Honeypot_mac string `json:"honeypot_mac"`
}

var (
	args struct {
		server      *string
		serverCIDR  *string
		broadcaster *string
		logPath     *string
		noNtpSync     *bool
	}
	subnet *net.IPNet
)

func init() {
	args.server = flag.String("c", "localhost:12345", "The server listening connection in format ip:port")
	args.serverCIDR = flag.String("s", "10.0.0.0/24", "The subnet the server belongs to")
	args.broadcaster = flag.String("bc", "localhost:23456", "The broadcaster connection that the server will connect to in format ip:port")
	args.logPath = flag.String("lp", "./server.log", "The path to the log file")
	args.noNtpSync = flag.Bool("no-ntp", false, "Do ntp sync")
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
	signal.Notify(sigs, syscall.SIGTERM, syscall.SIGINT)
	// method invoked upon seeing signal
	go func() {
		s := <-sigs
		log.Printf("RECEIVED SIGNAL: %s", s)
		os.Exit(1)
	}()
}

func execCommand(netData []byte, toBroadcaster chan []byte, broadcastEnabled bool) {
	var cmd command
	err := json.Unmarshal(netData, &cmd)
	if err != nil {
		log.Println(err)
		return
	}
	var out bytes.Buffer
	var stderr bytes.Buffer
	if cmd.Action == "block" {
		cmd1 := exec.Command("/usr/bin/ovs-ofctl", "add-flow", "br-int", "ip,nw_dst="+cmd.Argument.Ip+",actions=drop")
		cmd2 := exec.Command("/usr/bin/ovs-ofctl", "add-flow", "br-int", "ip,nw_src="+cmd.Argument.Ip+",actions=drop")
		cmd1.Stdout = &out
		cmd1.Stderr = &stderr
		cmd2.Stdout = &out
		cmd2.Stderr = &stderr
		log.Println("Executing ", cmd1)
		err = cmd1.Run()
		if err != nil {
			log.Println(err, ": ", stderr.String())
			return
		}
		log.Println("Result: " + out.String())
		log.Println("Executing ", cmd2)
		err = cmd2.Run()
		if err != nil {
			log.Println(err, ": ", stderr.String())
			return
		}
		log.Println("Result: " + out.String())

		ip := net.ParseIP(cmd.Argument.Ip)
		if !subnet.Contains(ip) && broadcastEnabled {
			toBroadcaster <- netData
		}
	} else if cmd.Action == "unblock" {
		cmd1 := exec.Command("/usr/bin/ovs-ofctl", "del-flows", "--strict", "br-int", "ip,nw_src="+cmd.Argument.Ip)
		cmd2 := exec.Command("/usr/bin/ovs-ofctl", "del-flows", "--strict", "br-int", "ip,nw_dst="+cmd.Argument.Ip)
		cmd1.Stdout = &out
		cmd1.Stderr = &stderr
		cmd2.Stdout = &out
		cmd2.Stderr = &stderr
		log.Println("Executing ", cmd1)
		err = cmd1.Run()
		if err != nil {
			log.Println(err, ": ", stderr.String())
			return
		}
		log.Println("Result: " + out.String())
		log.Println("Executing ", cmd2)
		err = cmd2.Run()
		if err != nil {
			log.Println(err, ": ", stderr.String())
			return
		}
		log.Println("Result: " + out.String())

		ip := net.ParseIP(cmd.Argument.Ip)
		if !subnet.Contains(ip) && broadcastEnabled {
			toBroadcaster <- netData
		}
	} else if cmd.Action == "throttle" {
		limit, err := strconv.Atoi(cmd.Argument.Limit)
		if err != nil {
			log.Println(err)
			return
		}
		barrier := limit * 100
		limit = limit * 1000
		cmd1 := exec.Command("/usr/bin/ovs-vsctl", "set", "interface", cmd.Argument.Port, "ingress_policing_rate="+strconv.Itoa(limit))
		cmd2 := exec.Command("/usr/bin/ovs-vsctl", "set", "interface", cmd.Argument.Port, "ingress_policing_burst="+strconv.Itoa(barrier))
		cmd1.Stdout = &out
		cmd1.Stderr = &stderr
		cmd2.Stdout = &out
		cmd2.Stderr = &stderr
		log.Println("Executing ", cmd1)
		err = cmd1.Run()
		if err != nil {
			log.Println(err, ": ", stderr.String())
			return
		}
		log.Println("Result: " + out.String())
		log.Println("Executing ", cmd2)
		err = cmd2.Run()
		if err != nil {
			log.Println(err, ": ", stderr.String())
			return
		}
		log.Println("Result: " + out.String())

		if broadcastEnabled {
			toBroadcaster <- netData
		}
	} else if cmd.Action == "forward" {
		cmd1 := exec.Command("/usr/bin/ovs-ofctl", "add-flow", "br-int", "table=70,ip,nw_dst="+cmd.Argument.Ip+",priority=300,actions=drop")
		cmd2 := exec.Command("/usr/bin/ovs-ofctl", "add-flow", "br-int", "table=70,tcp,tcp_dst=80,nw_dst="+cmd.Argument.Ip+",actions=mod_nw_dst:"+cmd.Argument.Honeypot_ip+",mod_dl_dst:"+cmd.Argument.Honeypot_mac+",goto_table:71")
		cmd3 := exec.Command("/usr/bin/ovs-ofctl", "add-flow", "br-int", "table=10,ip,dl_src="+cmd.Argument.Honeypot_mac+",nw_src="+cmd.Argument.Honeypot_ip+",actions=mod_nw_src:"+cmd.Argument.Ip+",goto_table:29")
		cmd1.Stdout = &out
		cmd1.Stderr = &stderr
		cmd2.Stdout = &out
		cmd2.Stderr = &stderr
		cmd3.Stdout = &out
		cmd3.Stderr = &stderr
		log.Println("Executing ", cmd1)
		err = cmd1.Run()
		if err != nil {
			log.Println(err, ": ", stderr.String())
			return
		}
		log.Println("Result: " + out.String())
		log.Println("Executing ", cmd2)
		err = cmd2.Run()
		if err != nil {
			log.Println(err, ": ", stderr.String())
			return
		}
		log.Println("Result: " + out.String())
		log.Println("Executing ", cmd3)
		err = cmd3.Run()
		if err != nil {
			log.Println(err, ": ", stderr.String())
			return
		}
		log.Println("Result: " + out.String())
	} else if cmd.Action == "tarpit" {
		cmd1 := exec.Command("/usr/bin/ovs-ofctl", "add-flow", "br-int", "ip,nw_dst="+cmd.Argument.Ip+",action=set_queue:100,goto_table:10")
		cmd2 := exec.Command("/usr/bin/ovs-ofctl", "add-flow", "br-int", "ip,nw_src="+cmd.Argument.Ip+",action=set_queue:100,goto_table:10")
		cmd1.Stdout = &out
		cmd1.Stderr = &stderr
		cmd2.Stdout = &out
		cmd2.Stderr = &stderr
		log.Println("Executing ", cmd1)
		err = cmd1.Run()
		if err != nil {
			log.Println(err, ": ", stderr.String())
			return
		}
		log.Println("Result: " + out.String())
		log.Println("Executing ", cmd2)
		err = cmd2.Run()
		if err != nil {
			log.Println(err, ": ", stderr.String())
			return
		}
		log.Println("Result: " + out.String())

		ip := net.ParseIP(cmd.Argument.Ip)
		if !subnet.Contains(ip) && broadcastEnabled {
			toBroadcaster <- netData
		}
	} else {
		log.Println("command ", cmd)
	}
	log.Println("executed command ", string(netData))
}

func connectionReader(c net.Conn, toBroadcaster chan []byte, broadcastEnabled bool) {
	defer func() {
		log.Printf("Reader Connection closed %s\n", c.RemoteAddr().String())
		c.Close()
	}()

	log.Printf("Serving reader %s\n", c.RemoteAddr().String())
	reader := bufio.NewReader(c)
	for {
		netData, err := reader.ReadBytes('\n')
		if err != nil {
			log.Println(err)
			break
		}

		log.Println("received command from ", c.RemoteAddr().String(), ": ", string(netData))
		go execCommand(netData, toBroadcaster, broadcastEnabled)
	}
	// if a flow controller connection is closed we let the handler terminate
}

func connectionWriter(c net.Conn, toBroadcaster chan []byte) {
	defer func() {
		log.Printf("Writer Connection closed %s\n", c.RemoteAddr().String())
		c.Close()
	}()

	log.Printf("Serving writer %s\n", c.RemoteAddr().String())
	for {
		message := <-toBroadcaster
		_, err := c.Write(message)
		if err != nil {
			log.Println(err)
			os.Exit(1)
		}
	}
}

func main() {
	// port to listen to input connections (flow controllers)
	var err error
	_, subnet, err = net.ParseCIDR(*args.serverCIDR)
	if err != nil {
		log.Println(err)
		return
	}
	log.Println("Local subnet:", subnet)

	var retries int = 0
	var connBroadcaster net.Conn

	if !(*args.noNtpSync) {
		ntpSync()
	}

	for retries < 10 {
		connBroadcaster, err = net.Dial("tcp4", *args.broadcaster)
		if err == nil {
			break
		}

		log.Println(err)
		retries++
		if retries < 10 {
			log.Println("Retrying")
		} else {
			log.Println("Failed to connect to broadcaster")
		}
		time.Sleep(5 * time.Second)
	}

	toBroadcaster := make(chan []byte)

	// we open a nnnew writer routine for the broadcaster connection
	go connectionWriter(connBroadcaster, toBroadcaster)

	// we open a new reader routine for the broadcaster connection
	go connectionReader(connBroadcaster, nil, false)

	listener, err := net.Listen("tcp4", *args.server)
	if err != nil {
		log.Println(err)
		return
	}
	defer listener.Close()
	// whenever a flow controller connects we open a new reader routine
	for {
		c, err := listener.Accept()
		if err != nil {
			log.Println(err)
			return
		}
		go connectionReader(c, toBroadcaster, true)
	}
}
