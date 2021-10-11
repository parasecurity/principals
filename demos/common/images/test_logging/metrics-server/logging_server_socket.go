package main

import (
	"os"
	"io"
	"fmt"
	"encoding/binary"
	"os/signal"
	"syscall"
	"net"
	"bufio"
	"log"
	"sync"
)

var (
	logger *server_logger
)

type server_logger struct {
	m sync.Mutex
	file *os.File
}

func init_server_logger(fname string) (*server_logger) {
	f, err := os.OpenFile(fname, os.O_APPEND|os.O_RDWR|os.O_CREATE, 0664)
	if err != nil {
		print("error opening logs' file")
		return nil
	}
	sl := new(server_logger)
	sl.file = f
	return sl
}

func (sl *server_logger) Write(p []byte) (n int, err error) {
	sl.m.Lock()
	n, err = sl.file.Write(p)
	sl.m.Unlock()
	return n, err
}

func init() {

	// Open log file
	// logFile, err := os.OpenFile("logs", os.O_APPEND|os.O_RDWR|os.O_CREATE, 0644)
	// if err != nil {
	// 	log.Println(err)
	// 	return
	// }
	logger = init_server_logger("logs.log")

	log.SetFlags(log.Ldate | log.Lmicroseconds | log.LUTC)
	log.SetOutput(logger)
	log.SetPrefix("logging server: ")


	// Setup signal catching
	sigs := make(chan os.Signal, 1)
	// Catch all signals since not explicitly listing
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)
	// Method invoked upon seeing signal
	go func() {
		s := <-sigs
		log.Printf("RECEIVED SIGNAL: %s", s)
		os.Exit(1)
	}()
}

func handleConnection(c net.Conn){

	defer func() {
		c.Close()
		log.Printf("Connection closed: %s", c.RemoteAddr())
	}()

	reader := bufio.NewReader(c)
	log.Printf("Serving %s", c.RemoteAddr().String())
	cli_log := log.New(log.Writer(), c.RemoteAddr().String()+": ", 0)

	for {

		str, err := reader.ReadString('\n')

		if err != nil {
			if err == io.EOF {
				break;
			} else {
				log.Println(err)
			}
		}
		cli_log.Print(str)
	}
}

func handleMetricsConnection(c net.Conn){
	//vers0.2
	// Recieve metric data from client
	reader := bufio.NewReader(c)
	for {
		//str, err := reader.ReadString('\n')

		b := make([]byte, 8)
		var err error
		b[0], err = reader.ReadByte()
		b[1], err = reader.ReadByte()
		b[2], err = reader.ReadByte()
		b[3], err = reader.ReadByte()
		b[4], err = reader.ReadByte()
		b[5], err = reader.ReadByte()
		b[6], err = reader.ReadByte()
		b[7], err = reader.ReadByte()

		str, err := reader.ReadString('\n')

		if err != nil {
			if err == io.EOF {
				break;
			} else {
				log.Println(err)
			}
		} else {
			//log.Printf(str)
			t := binary.LittleEndian.Uint64(b[0:8])
			msg := fmt.Sprintf("%v %s", t, str)
			log.Printf(msg)
		}
	}
	c.Close()
	log.Printf("Connection closed: %s", c.RemoteAddr())
}

func main() {

	sock_addr, err := net.ResolveUnixAddr("unix", "/tmp/fastlog.sock")

	if err := os.RemoveAll(SockAddr); err != nil {
        log.Fatal(err)
    }

	listener, err := net.ListenUnix("unix", sock_addr)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Listenning on port 4321")

	for {
		cli, err := listener.AcceptUnix()
		if err != nil {
			log.Fatal("Accept failed:", err.Error())
			break
		}
		log.Printf("Connection open: %s ", cli.RemoteAddr())
		go handleConnection(cli)
	}
	listener.Close()
}
