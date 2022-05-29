package statistics

import (
	"net"
	"log"
	"github.com/hpcloud/tail"
	"sync"
	"time"
	"encoding/json"
)

type receivedData struct {
	Primitive string  `json:"primitive"`
	Data 	string   `json:"data"`
}

func connectServer(ip *string, port *string) net.Conn {
	serverAddress := *ip + ":" + *port
	connection, err := net.Dial("tcp4", serverAddress)
	if err != nil {
		log.Println(err)
		return nil
	}

	return connection
}

func sendStatistics(primitive *string, connection net.Conn, statistics *string) {
	msg := receivedData{
		*primitive,
		*statistics,
	}
	jsonMsg, _ := json.Marshal(msg)
	jsonMsg = append(jsonMsg, []byte("\n")...)
	log.Println(string(jsonMsg))

	_, err := connection.Write(jsonMsg)
	if err != nil {
		log.Println(err)
		return
	}
}

func trackStatistics(logFile *string, data *string, mutex *sync.Mutex) {
	t, err := tail.TailFile(*logFile, tail.Config{Follow: true})
	if err != nil {
		return
	}

	for line := range t.Lines {
		mutex.Lock()
		*data = *data + line.Text 
		mutex.Unlock()
	}
}

func HandleStatistics(primitive *string, ip *string, port *string, logFile *string, pollingRate int) {
	var mutex sync.Mutex
	data:= ""
	conn := connectServer(ip, port)
	go trackStatistics(logFile, &data, &mutex)
	
	for {
		mutex.Lock()
		if data != "" {
			sendStatistics(primitive, conn, &data)
			data = ""
		}
		mutex.Unlock()
		time.Sleep(5 * time.Second)
	}
}