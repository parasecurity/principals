package statistics

import (
	"net"
	"log"
	"github.com/hpcloud/tail"
	"sync"
	"time"
)

func connectServer(ip *string, port *string) net.Conn {
	serverAddress := *ip + ":" + *port
	connection, err := net.Dial("tcp4", serverAddress)
	if err != nil {
		log.Println(err)
		return nil
	}

	return connection
}

func sendStatistics(connection net.Conn, statistics *string) {
	_, err := connection.Write([]byte(*statistics))
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

func HandleStatistics(ip *string, port *string, logFile *string, pollingRate int) {
	var mutex sync.Mutex
	data:= ""
	conn := connectServer(ip, port)
	go trackStatistics(logFile, &data, &mutex)
	
	for {
		mutex.Lock()
		if data != "" {
			sendStatistics(conn, &data)
			data = ""
		}
		mutex.Unlock()
		time.Sleep(5 * time.Second)
	}
}