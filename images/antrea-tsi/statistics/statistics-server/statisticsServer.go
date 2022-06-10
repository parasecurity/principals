package main

import (
	"bufio"
	"flag"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"encoding/json"
)

type nodesData struct {
	nodes map[string]bool
	data map[string]string
	nodeCounter int
}

type primitiveData struct {
	primitives map[string]bool
	data map[string]nodesData
	primitiveCounter int
	mutex sync.RWMutex
}

var (
	args struct {
		statisticsAddress    *string
		APIstatisticsAddress *string
		logPath        *string
	}
)

type statisticsData struct {
	NodeName     string `json:"nodename"`
	Primitive    string `json:"primitive"`
	Data         string `json:"data"`
}

type APIcommand struct {
	NodeName     string `json:"nodename"`
	Primitive    string `json:"primitive"`
}

func sendData(c net.Conn, cmd APIcommand, data string) {
	resp := statisticsData{
		cmd.NodeName,
		cmd.Primitive,
		data,
	}
	jsonMsg, err := json.Marshal(resp)
	if err != nil {
		log.Println(err)
	}
	jsonMsg = append(jsonMsg, []byte("\n")...)
	log.Println(string(jsonMsg))
	_, err = c.Write(jsonMsg)
	if err != nil {
		log.Println(err)
	}
}

func initNodeData(nodeName string, data string) nodesData {
	var tmpNodeData nodesData

	tmpNodeData.nodes = make(map[string]bool)
	tmpNodeData.data = make(map[string]string)
	tmpNodeData.nodeCounter = 0

	tmpNodeData.nodes[nodeName] = true
	tmpNodeData.data[nodeName] = data
	tmpNodeData.nodeCounter += 1
	return tmpNodeData
}

func init() {
	args.statisticsAddress = flag.String("c", "localhost:30000", "The server listening connection in format ip:port")
	args.APIstatisticsAddress = flag.String("ac", "localhost:30001", "The api listening connection in format ip:port")
	args.logPath = flag.String("lp", "./statisticsServer.log", "The path to the log file")
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
	signal.Notify(sigs)
	// method invoked upon seeing signal
	go func() {
		s := <-sigs
		log.Printf("RECEIVED SIGNAL: %s", s)
		os.Exit(1)
	}()
}

func handleAPIConnection(c net.Conn, primitiveList *primitiveData) {
	log.Printf("Serving API %s\n", c.RemoteAddr())
	reader := bufio.NewReader(c)
	for {
		netData, err := reader.ReadBytes('\n')
		if err != nil {
			log.Println(err)
			break
		}
		log.Println("from API ", c.RemoteAddr(), ": ", string(netData))

		var cmd APIcommand
		err = json.Unmarshal(netData, &cmd)
		if err != nil {
			log.Println(err)
			return
		}

		// whenever the API server sends data we reply with all the statistics
		primitiveList.mutex.Lock()
		found := false
		if primitiveList.primitives[cmd.Primitive] {
			reqPrimitive := primitiveList.data[cmd.Primitive]
			for node, data := range reqPrimitive.data {
				if node == cmd.NodeName {
					sendData(c, cmd, data)
					found = true
					break
				}
			}
		}
		if !found {
			sendData(c, cmd, "")
		}
		primitiveList.mutex.Unlock()
	}
	c.Close()
	log.Println("API Connection closed ", c.RemoteAddr())
	// if the API server connection is closed we let the handler terminate
}

func handleConnection(c net.Conn, primitiveList *primitiveData) {
	log.Printf("Serving %s, idx\n", c.RemoteAddr())
	reader := bufio.NewReader(c)
	for {
		netData, err := reader.ReadBytes('\n')
		if err != nil {
			log.Println(err)
			break
		}

		var cmd statisticsData
		err = json.Unmarshal(netData, &cmd)
		if err != nil {
			log.Println(err)
			return
		}
		log.Printf("The received data are " + cmd.NodeName + " " + cmd.Primitive + " " + cmd.Data)
		
		primitiveList.mutex.Lock()
		// if this primitive type exists
		if primitiveList.primitives[cmd.Primitive] {
			if primitiveList.data[cmd.Primitive].nodes[cmd.NodeName] {
				primitiveList.data[cmd.Primitive].data[cmd.NodeName] += cmd.Data
			} else {
				primitiveList.data[cmd.Primitive] = initNodeData(cmd.NodeName, cmd.Data)
			}
		} else {
			primitiveList.primitives[cmd.Primitive] = true
			primitiveList.data[cmd.Primitive] = initNodeData(cmd.NodeName, cmd.Data)
			primitiveList.primitiveCounter++;
		}
		primitiveList.mutex.Unlock()

	}
	c.Close()
	log.Println("Connection closed ", c.RemoteAddr())
	//closeOutConn(c_idx, nodeList.c[c_idx], nodeList)
}

func main() {
	listener, err := net.Listen("tcp4", *args.statisticsAddress)
	if err != nil {
		log.Println(err)
		return
	}
	defer listener.Close()

	primitiveList := new(primitiveData)
	primitiveList.primitives = make(map[string]bool)
	primitiveList.data = make(map[string]nodesData)
	primitiveList.primitiveCounter = 0
	
	go func() {
		for {
			c, err := listener.Accept()
			if err != nil {
				log.Println(err)
				return
			}

			go handleConnection(c, primitiveList)
		}
	}()

	// port to listen to input nodesData (API server)
	APIlistener, err := net.Listen("tcp4", *args.APIstatisticsAddress)
	if err != nil {
		log.Println(err)
		return
	}
	defer APIlistener.Close()

	// whenever an API server connects we open a new handler
	for {
		c, err := APIlistener.Accept()
		if err != nil {
			log.Println(err)
			return
		}

		go handleAPIConnection(c, primitiveList)
	}
}
