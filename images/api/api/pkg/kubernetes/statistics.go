package kubernetes

import (
	"log"
	"strings"
	"bufio"
	"encoding/json"
	"net"
)

type APIcommand struct {
	NodeName     string `json:"nodename"`
	Primitive    string `json:"primitive"`
}


func getData(args []string, shortArg string, longArg string) string {
	argsLength := len(args)
	for counter := 0; counter < argsLength; counter++ {
		value := args[counter]
		isCommand := strings.Contains(value, shortArg) || strings.Contains(value, longArg)
		if isCommand {
			valueArray := strings.Split(value, "=")
			return valueArray[1]
		}
	}

	return "nil"
}


func Statistics(command Command, c net.Conn) {
	nodeName := getData(command.Arguments, "-n", "-node") 
	req := APIcommand{
		nodeName,
		command.Target, // this contains the the target primitive like dns-stitching
	}

	jsonMsg, err := json.Marshal(req)
	if err != nil {
		log.Println(err)
	}
	jsonMsg = append(jsonMsg, []byte("\n")...)
	log.Println(string(jsonMsg))
	_, err = c.Write(jsonMsg)
	if err != nil {
		log.Println(err)
	}

	reader := bufio.NewReader(c)
	netData := make([]byte, 4096)
	_, err = reader.Read(netData)
	if err != nil {
		log.Println(err)
	}
	dataString := string(netData)
	log.Println("Statistics: ", dataString)
}
